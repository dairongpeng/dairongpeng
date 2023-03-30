## DaemonSet
当你希望在集群中每个节点上运行时，并且每个节点正好需要运行一个Pod实例。例如每个节点运行日志收集和资源监控的Agent, 或是Kubernetes自己的kube-proxy进程也需要在所有节点上运行才能服务。

RC和RS需要期望副本数的概念，而DaemonSet不需要，因为它的工作是确保一个pod匹配它的选择器并在每个Node节点上运行。DaemonSet设计出来的目的是运行系统服务，即使在不可调度的节点上，系统服务也会被分配。另外DaemonSet也是从自身的配置模板中创建pod。


### DaemonSet定义模板
```yaml
# daemonSet在apps组中，版本为v1beta2
apiVersion: apps/v1beta2
kind: DaemonSet
metadata:
    name: ssd-monitor
spec:
    selector:
        matchLabels:
            app: ssd-monitor
    template:
        metadata:
            labels:
                app: ssd-monitor
        spec:
            nodeSelector: # 不加节点选择器，默认为每个节点运行一个，需要确保存在该标签的node
                disk: ssd
            containers:
            - name: main
              image: luksa/ssd-monitor
```

## Job
ReplicationController、 ReplicaSet和 DaemonSet资源会持续运行任务，RS和RC会维护期望，DaemonSet会维持自己的标签选择的Node上运行的Pod情况，其都是永远达不到完成态。而Job资源即是为了解决这个问题的，用来定义一个可以被完成的任务，任务完成后进程退出，不应该再重新启动。


### Job定义模板
```yaml
# Job再batch组中，版本为v1
apiVersion: batch/v1
kind: Job
metadata:
    name: batch-job
spec:
    template: # 这里没指定pod选择器，会根据pod模板创建
        metadata:
            labels:
                app: batch-job
        spec:
            restartPolicy: OnFailure # Job不能使用Always为默认的重新启动策略。这里会防止容器再完成任务时重新启动
            containers:
            - name: main
              image: luksa/batch-job
```

Job演示：
```shell
[dairongpeng@dev workspace]$ cat job-test.yaml
apiVersion: batch/v1
kind: Job
metadata:
    name: batch-job
spec:
    template:
        metadata:
            labels:
                app: batch-job
        spec:
            restartPolicy: OnFailure
            containers:
            - name: main
              image: luksa/batch-job
[dairongpeng@dev workspace]$ kubectl create -f job-test.yaml
job.batch/batch-job created
[dairongpeng@dev workspace]$ kubectl get job
NAME        COMPLETIONS   DURATION   AGE
batch-job   0/1           56s        56s
# 两分钟后显示job已经完成退出
[dairongpeng@dev workspace]$ kubectl get jobs
NAME        COMPLETIONS   DURATION   AGE
batch-job   1/1           2m41s      3m
# 此时Job对应的pod任务已经显示完成
[dairongpeng@dev workspace]$ kubectl get po -A | grep batch
default       batch-job-gl6nf                          0/1     Completed   0               6m11s
[dairongpeng@dev workspace]$ kubectl logs batch-job-gl6nf
Sat Mar 18 07:31:17 UTC 2023 Batch job starting
Sat Mar 18 07:33:17 UTC 2023 Finished succesfully
[dairongpeng@dev workspace]$
```
### Job多任务（多Pod实例）
#### 顺序运行Job Pod
```yaml
# 相关串行运行Job中Pod的配置，其他属性神略
spec:
    completions: 5
    template:
```
#### 并行运行Job Pod
```yaml
# 相关并行运行Job中Pod的配置，其他属性神略
spec:
    completions: 5 # 该任务必须确保五个Pod成功完成
    parallelism: 2 # 最多同时运行两个pod
    template:
```

- 通过`k scale job [job名] --replicas [期望数]`可以调整Job运行的期望。
- 通过在pod中设置activeDeadlineSeconds属性，可以限制pod的时间，如果超时，系统尝试终止该pod，并将Job标记为失败，通过`spec.backoffLimit`可以配置Job失败后重试次数，默认为6次。

## Job定期运行和定时运行（CronJob）
```yaml
# CronJob的api组为batch，版本为v1beta1
apiVersion: batch/v1beta1
kind: CronJob
metadata:
    name: batch-job-every-fifteen-minutes
spec:
    schedule: "0,15,30,45 * * * *" # 每小时0，15，30，45分的时候运行
    jobTemplate: # 资源模板
        spec:
            template:
                metadata:
                    labels:
                        app: periodic-batch-job
                spec:
                    restartPolicy: OnFailure
                    containers:
                    - name: main
                      image: luksa/batch-job
```

> cron表达式：从左到右为 分钟、小时、每月中的第几天、月、星期几。例如希望每月的第一天的每半小时执行一次，那么为：`0,30 * 1 * *`，如果希望每个星期天的凌晨3点运行，则为：`0 3 * * 0`

### 配置任务截至日期
```yaml
# 只显示关键配置，其他配置项略
apiVersion: batch/v1beta1
kind: CronJob
metadata:
    name: batch-job-every-fifteen-minutes
spec:
    schedule: "0,15,30,45 * * * *" # 每小时0，15，30，45分的时候运行
    startingDeadlineSecond: 15 # Pod必须在定时时间后的15秒内开始，否则被认为失败
```

## 使用Job类资源需要注意的点
- Job或者CronJob类定时或定期的任务，必须设计成幂等的，即任务执行一次或者执行多次，或并发执行，都可以得到相同的期望结果。
- 请确保下一个任务运行完成本应该上一次（错过的或者失败的）运行完成的任何工作，即当前任务的失败，会累计到下一次任务的基础上，确保任务不会无缘无故遗漏。