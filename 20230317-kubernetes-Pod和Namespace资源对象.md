## Kubernetes资源对象

几乎所有的Kubernetes资源yaml定义都可以找到三大重要部分：
- metadata: 包括名称、命名空间、标签和关于该容器的其他信息。
- spec: 包含pod内容的实际说明，例如pod的容器，卷和其他数据。
- status: 包含运行中pod的当前信息，例如pod所处的条件，每个容器的描述和状态，以及内部IP和其他信息，一般对资源进行`-o yaml`时可看到。而在创建资源时，不需要提供。

## Pod
资源yaml定义：

```yaml
apiVersion: v1 # api版本号
kind: Pod # 资源类型
metadata:
    name: kubia-manual # Pod的名称
spec:
    containers:
    - image: luksa/kubia # 容器创建所需的镜像
      name: kubia # 容器的名称
      ports:
      - containerPort: 8080 # 应用监听的端口
        protocol: TCP
```

### 运用kubectl帮助我们理解资源定义
例如：我们使用`kubectl explain pods`可以看到k8s对于定义资源Pod相关字段的解释，例如存在一个spec，我们还想知道spec下的资源解释，则可以使用`kubectl explain pod.spec`

### 通过资源定义启动Pod
命令：`kubectl create -f kubia-manual-pod-test.yaml`

```shell
[dairongpeng@dev workspace]$ cat kubia-manual-pod-test.yaml
apiVersion: v1 # api版本号
kind: Pod # 资源类型
metadata:
    name: kubia-manual # Pod的名称
spec:
    containers:
    - image: luksa/kubia # 容器创建所需的镜像
      name: kubia # 容器的名称
      ports:
      - containerPort: 8080 # 应用监听的端口
        protocol: TCP
[dairongpeng@dev workspace]$ kubectl create -f kubia-manual-pod-test.yaml
pod/kubia-manual created
[dairongpeng@dev workspace]$ k get pod | grep kubia
kubia-manual   1/1     Running   0          10m
```

### 查看运行时的Pod定义文件
可以使用`k get po kubia-manual -o yaml`或者`k get po kubia-manual -o json`, 可以看到相比于定义资源，运行时的yaml字段更为丰富。
```shell
[dairongpeng@dev workspace]$ k get po kubia-manual -o yaml
apiVersion: v1
kind: Pod
metadata:
  creationTimestamp: "2023-03-16T15:24:43Z"
  name: kubia-manual
  namespace: default
  resourceVersion: "110371"
  uid: a146c2c6-0d99-4e22-9169-090940b72188
spec:
  containers:
  - image: luksa/kubia
    imagePullPolicy: Always
    name: kubia
    ports:
    - containerPort: 8080
      protocol: TCP
    resources: {}
    terminationMessagePath: /dev/termination-log
    terminationMessagePolicy: File
    volumeMounts:
    - mountPath: /var/run/secrets/kubernetes.io/serviceaccount
      name: kube-api-access-jjzqd
      readOnly: true
  dnsPolicy: ClusterFirst
  enableServiceLinks: true
  nodeName: docker-desktop
  preemptionPolicy: PreemptLowerPriority
  priority: 0
  restartPolicy: Always
  schedulerName: default-scheduler
  securityContext: {}
  serviceAccount: default
  serviceAccountName: default
  terminationGracePeriodSeconds: 30
  tolerations:
  - effect: NoExecute
    key: node.kubernetes.io/not-ready
    operator: Exists
    tolerationSeconds: 300
  - effect: NoExecute
    key: node.kubernetes.io/unreachable
    operator: Exists
    tolerationSeconds: 300
  volumes:
  - name: kube-api-access-jjzqd
    projected:
      defaultMode: 420
      sources:
      - serviceAccountToken:
          expirationSeconds: 3607
          path: token
      - configMap:
          items:
          - key: ca.crt
            path: ca.crt
          name: kube-root-ca.crt
      - downwardAPI:
          items:
          - fieldRef:
              apiVersion: v1
              fieldPath: metadata.namespace
            path: namespace
status:
  conditions:
  - lastProbeTime: null
    lastTransitionTime: "2023-03-16T15:24:43Z"
    status: "True"
    type: Initialized
  - lastProbeTime: null
    lastTransitionTime: "2023-03-16T15:34:36Z"
    status: "True"
    type: Ready
  - lastProbeTime: null
    lastTransitionTime: "2023-03-16T15:34:36Z"
    status: "True"
    type: ContainersReady
  - lastProbeTime: null
    lastTransitionTime: "2023-03-16T15:24:43Z"
    status: "True"
    type: PodScheduled
  containerStatuses:
  - containerID: docker://14b20c7479469c791e6c8e358c64b056234639e59ca01bf8154f6f7e56aaf4bd
    image: luksa/kubia:latest
    imageID: docker-pullable://luksa/kubia@sha256:3f28e304dc0f63dc30f273a4202096f0fa0d08510bd2ee7e1032ce600616de24
    lastState: {}
    name: kubia
    ready: true
    restartCount: 0
    started: true
    state:
      running:
        startedAt: "2023-03-16T15:34:35Z"
  hostIP: 192.168.65.4
  phase: Running
  podIP: 10.1.0.41
  podIPs:
  - ip: 10.1.0.41
  qosClass: BestEffort
  startTime: "2023-03-16T15:24:43Z"
[dairongpeng@dev workspace]$
```

### 查看该Pod日志
容器化的应用程序通常会将日志记录到标准输出和标准错误流， 而不是将其写入文件， 这就允许用户可以通过
简单、 标准的方式查看不同应用程序的日志。

如果一个pod内包含多个容器，我们希望指定查看某个容器的日志可以使用`kubectl logs -c [容器名]`
```shell
[dairongpeng@dev workspace]$ k logs kubia-manual -c kubia
Kubia server starting...
[dairongpeng@dev workspace]$
```

如果你的pod崩溃后重启了，由于`kubectl log`命令默认只能看到当前Pod的日志，如果希望看到上一个被kill掉的pod的日志，可以使用`kubectl log [pod] --previous`

### 手动配置端口转发
如果我们还没接触过service，那么如何和我们建立的pod进行通信，可以通过`kubectl port-forward`端口转发来实现，例如我们可以将机器的本地端口的8888转发到kubia-manual pod的端口8080。
```shell
[dairongpeng@dev workspace]$ kubectl port-forward kubia-manual 8888:8080
Forwarding from 127.0.0.1:8888 -> 8080
Forwarding from [::1]:8888 -> 8080
```

此时正在进行端口转发，本地另起终端测试转发情况：
```shell
[dairongpeng@dev workspace]$ curl 127.0.0.1:8888
You've hit kubia-manual
[dairongpeng@dev workspace]$
```

### 为资源打标签
```yaml
apiVersion: v1 # api版本号
kind: Pod # 资源类型
metadata:
    name: kubia-manual-v2 # Pod的名称
    labels: # 为Pod增加两个标签
      creation_method: manual
      env: prod
spec:
    containers:
    - image: luksa/kubia # 容器创建所需的镜像
      name: kubia # 容器的名称
      ports:
      - containerPort: 8080 # 应用监听的端口
        protocol: TCP
```

- 创建带标签的资源

```shell
[dairongpeng@dev workspace]$ k create -f kubia-manual-pod-test2.yaml
pod/kubia-manual-v2 created
[dairongpeng@dev workspace]$ k get pod | grep kubia
kubia-manual      1/1     Running   0          31m
kubia-manual-v2   1/1     Running   0          44s
[dairongpeng@dev workspace]$ k get pod --show-labels | grep kubia-manual-v2
kubia-manual-v2   1/1     Running   0          100s   creation_method=manual,env=prod
```

- 选中带有特定标签的资源(标签选择器)
```shell
# 选出标签key为creation_method其value为manual的pod资源
[dairongpeng@dev workspace]$ k get po -l creation_method=manual
NAME              READY   STATUS    RESTARTS   AGE
kubia-manual-v2   1/1     Running   0          7m44s
# 列出包含env标签的所有pod，无论其值如何
[dairongpeng@dev workspace]$ k get po -l env
NAME              READY   STATUS    RESTARTS   AGE
kubia-manual-v2   1/1     Running   0          8m56s
# 列出没有标签env的pod资源
[dairongpeng@dev workspace]$ k get po -l '!env'
NAME           READY   STATUS    RESTARTS   AGE
kubia-manual   1/1     Running   0          41m
nginx          0/1     Error     0          24h
```
标签的更多用法：
- env in (prod, dev)
- env!=prod
- env not in (prod, dev)

### 节点选择器
```yaml
apiVersion: v1 # api版本号
kind: Pod # 资源类型
metadata:
    name: kubia-manual-v3 # Pod的名称
spec:
    nodeSelector: # 添加节点选择器，让该pod运行在标签为gpu，标签值为"true"的特定节点Node上
      gpu: "true"
    containers:
    - image: luksa/kubia # 容器创建所需的镜像
      name: kubia # 容器的名称
      ports:
      - containerPort: 8080 # 应用监听的端口
        protocol: TCP
```

## Namespace
Kubernetes命名空间简单地为对象名称提供一个作用域。例如我们可以按照Namespace把资源划分为不同环境，测试环境，开发环境等。

定义命名空间资源:
```yaml
apiVersion: v1
kind: Namespace # 资源类型为命名空间
metadata:
  name: dev # 命名空间的名称
```

- 使用资源配置创建namespace
```shell
[dairongpeng@dev workspace]$ cat namespace-dev.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: dev
[dairongpeng@dev workspace]$ k create -f namespace-dev.yaml
namespace/dev created
[dairongpeng@dev workspace]$ k get ns | grep dev
dev               Active   8s
[dairongpeng@dev workspace]$
```

当然我们也可以快捷创建一个ns，使用`k create ns dev`即可，后续命令如果需要指定在哪个命名空间下，只需要在命令后增加`-n [命名空间]`即可。

- 设置快速切换命名空间, 在.bashrc中新增`alias kcd='kubectl config set-context $(kubectl config current-context) --namespace'`, 再source

```shell
[dairongpeng@dev workspace]$ kcd dev
Context "docker-desktop" modified.
[dairongpeng@dev workspace]$ k get pod
No resources found in dev namespace.
[dairongpeng@dev workspace]$ kcd default
Context "docker-desktop" modified.
[dairongpeng@dev workspace]$ k get pod
NAME              READY   STATUS    RESTARTS   AGE
kubia-manual      1/1     Running   0          75m
kubia-manual-v2   1/1     Running   0          44m
nginx             0/1     Error     0          24h
[dairongpeng@dev workspace]$
```