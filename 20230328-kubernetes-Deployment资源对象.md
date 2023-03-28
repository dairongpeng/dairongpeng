## Deployment的演变
一开始为了Pod的调度和维持期望，Kubernetes存在ReplicationController, 后续为了更强的Pod选择器能力，RC升级成为了ReplicaSet（RS）。但是通常我们并不会直接创建RS，而是创建更高级别的资源Deployment，该资源会自动创建RS。

RS标签选择器，不必在selector属性中直接列出pod需要的标签，而是在selector.matchLabels下指定它们。这是在ReplicaSet定义标签选择器的更简单方式。

- RC资源的apiVersion属于v1
- RS资源的apiVersion属于apps/v1beta2

RC资源定义关注点：

```yaml
# 仅显示关键配置
# 核心api组不需要指名，只需要写api版本即可
apiVersion: v1
kind: ReplicationController
# RC标签选择器
spec:
    selector:
        app: kubia
```

RS资源定义关注点：

```yaml
# 仅显示关键配置
# api组为apps，版本为v1beta2
apiVersion: apps/v1beta2
kind: ReplicaSet
# RS标签选择器
spec:
    selector:
        matchLabls:
            app: kubia
```

RS更强大的标签选择：

```yaml
# 仅显示标签选择相关配置
selector:
    matchExpressions:  # 要求被选择的pod含有app标签，其值为kubia
        - key: app
          operator: In # 可用的运算符包含 In、NotIn、Exists、DoesNotExist
          values:
            - kubia
```

## Deployment简介
Deployment时一种更高阶的资源，相比RS和RC（被认作更底层）。每当创建一个Deploymentg时，RS资源也会随之创建，Deployment控制ReplicaSet，RS控制Pods。

在原RS和RC的滚动升级中，需要引入额外的一个RS或者RC，而Deployment依附在k8s控制器进程，就是解决这个问题的。所以使用Deployment可以更容器的更新应用程序。

### 创建一个Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp-deployment
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - name: myapp-container
        image: your-docker-registry/myapp-image:latest
        ports:
        - containerPort: 80
```

在这个文件中，我们定义了一个名为“myapp-deployment”的deployment，它将使用“myapp”标签来选择pod，并且我们希望部署3个replica。此外，我们还定义了一个容器，其中包含我们的应用程序镜像。

使用kubectl apply命令创建我们的deployment。

```shell
kubectl apply -f deployment.yaml
```

这将创建一个新的deployment，它将在Kubernetes集群中运行我们的应用程序。

### 升级Deployment
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: myapp-deployment
spec:
  replicas: 3
  selector:
    matchLabels:
      app: myapp
  template:
    metadata:
      labels:
        app: myapp
    spec:
      containers:
      - name: myapp-container
        image: your-docker-registry/myapp-image:v2 # 这里替换为需要升级的镜像版本
        ports:
        - containerPort: 80
```

使用kubectl apply命令将更新的deployment.yaml文件应用于我们的deployment。

```shell
kubectl apply -f deployment.yaml
```

此时，Kubernetes将自动升级我们的deployment，它将在后台将旧的pod删除并创建新的pod。在这个过程中，我们的应用程序将不会停机，因为我们已经定义了3个replica，其中的至少2个将始终可用。

Deployment的滚动升级，是通过创建一个新的RS，新的Pod由新的RS管理，由于时自动创建的，用户无感知。升级后的RS实际上不一定会被删除掉，因为Deployment的升级依赖这个RS, 可以通过控制Deployment的revisionHistoryLimit来限制保存的历史版本的RS数量。

### 回滚Deployment
假设我们发现新的应用程序版本有一些问题，并且需要回滚到旧版本。
1. 查找可用的deployment历史版本。
```shell
kubectl rollout history deployment/myapp-deployment
```

2. 找到我们想要回滚到的历史版本的版本号，并使用kubectl命令回滚deployment。
```shell
kubectl rollout undo deployment/myapp-deployment --to-revision=2
```

Kubernetes将自动回滚我们的deployment，它将删除新版本的pod并创建旧版本的pod。

> 默认情况下，在10分钟内不能完成滚动升级，会被认为升级失败。可以通过设置spec的progressDeadlineSeconds来修改该默认值。