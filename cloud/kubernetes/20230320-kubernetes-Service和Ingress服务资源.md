## Service
为什么需要该资源，因为Pod会不停的创建销毁，需要拥有一个固定的接入口，这就是Service存在的意义，Service提供的IP和端口不会变，客户端对于该Service的路由会再路由到该服务背后任意的Pod上。服务的后端可以有不止一个Pod,Service队所有的后端Pod的连接时负载均衡的。

### Service资源定义
```yaml
# Service资源版本为v1
apiVersion: v1
kind: Service
metadata:
    name: kubia
spec:
    ports:
    - port: 80 # 该Service的可用端口
      targetPort: 8080 # 服务将连接转发到的容器端口
    selector:
        app: kubia # Pod标签为app，值为kubia的都被该Service代理
```

操作演示：
- 分配给该Service的IP为10.97.1.224，这个IP是集群的IP地址，只能在集群内部访问。该服务的主要目的是使集群内部其他Pod可以访问当前Service代理的后端Pod。当然也可以通过其他手段使得集群外部可以访问该Service。
```shell
[dairongpeng@dev workspace]$ cat service-test.yaml
apiVersion: v1
kind: Service
metadata:
    name: kubia
spec:
    ports:
    - port: 80
      targetPort: 8080
    selector:
        app: kubia
[dairongpeng@dev workspace]$ kubectl create -f service-test.yaml
service/kubia created
[dairongpeng@dev workspace]$ kubectl get svc
NAME         TYPE        CLUSTER-IP    EXTERNAL-IP   PORT(S)   AGE
kubernetes   ClusterIP   10.96.0.1     <none>        443/TCP   11d
kubia        ClusterIP   10.97.1.224   <none>        80/TCP    9s
```

如果希望来自同一个客户端IP的请求，都可以命中同一个Service后端的Pod上，可以在Service增加该配置项`spec.sessionAffinity: ClientIP`

### Service代理多端口例子
```yaml
apiVersion: v1
kind: Service
metadata:
    name: kubia
spec:
    ports:
    - name: http # Service的Http代理
      port: 80
      targetPort: 8080
    - name: https # Service的Https代理
      port: 443
      targetPort: 8443
    selector:
        app: kubia
```

### 关于服务发现
- 可以通过环境变量进行服务发现
一般来说当Service先于Pod(RC或者RS或Deployment负责调度)，该Pod可以发现该服务，写入这个Pod的环境变量中了，如果Pod在Service前创建，可以通过更改期望为0，再重新创建Pod，来发现所有的Service。

- 可以通过DNS进行服务发现
利用的是kube-dns的pod，运行了dns服务。在k8s集群中，其他的pod都被配置成使用其作为dns（k8s通过修改每个容器/etc/resolv.conf文件实现）。kube-dns知道系统中所有服务的地址。

> `服务名称.命名空间.svc.cluster.local`确定了一个服务的条目，特别的，如果客户端和服务在同一个命名空间下，可以使用服务名称作为主机名，代替`服务名称.命名空间.svc.cluster.local`，大大简化了服务的访问。

### Endpoint
在服务和Pod之间，存在Endpoint资源。服务的spec中定义Pod选择器，选择器用于构建IP和端口列表，然后存储到Endpoint资源中。当服务被链接时，服务代理选择背后Endpoints中的一个，转发过去。

### 暴露服务
- 通过将服务的类型设置成NodePort。每个集群节点都会在自身节点上打开一个对应的端口，并将该端口上接收到的流量重定向到基础服务，达到节点上的专用端口，转发到集群内部service的目的。
- 将服务的类型设置成LoadBalance，NodePort类型的一种扩展。负载均衡器将流量重定向到跨所有节点的节点端口。客户端可以通过负载均衡器的IP从而连接到集群中的服务。
- 创建Ingress资源，这是一个完全不同的机制，通过一个IP地址公开多个服务。

#### NodePort类型的服务
```yaml
apiVersion: v1
kind: Service
metadata:
  name: kubia-nodeport
spec:
  type: NodePort # NodePort服务类型
  ports:
  - port: 80 # 服务端口号
    targetPort: 8080 # 背后Pod的目标端口号
    nodePort: 30123 # 通过集群各个节点30123端口，可以转发到该服务（80端口）
  selector:
    app: kubia
```

```shell
[dairongpeng@dev workspace]$ cat nodeport-service-test.yaml
apiVersion: v1
kind: Service
metadata:
  name: kubia-nodeport
spec:
  type: NodePort
  ports:
  - port: 80
    targetPort: 8080
    nodePort: 30123
  selector:
    app: kubia
[dairongpeng@dev workspace]$ k create -f nodeport-service-test.yaml
service/kubia-nodeport created
# <none>表明服务可以通过集群任何节点的IP地址访问
[dairongpeng@dev workspace]$ k get svc | grep kubia-nodeport
kubia-nodeport   NodePort    10.100.168.135   <none>        80:30123/TCP   21s
```

> 当采用NodePort的方式，如果客户端配置访问Node1的IP和端口，当Node1宕机时，客户端便法务访问该服务，所以更需要负载均衡的方式来访问服务。
#### LoadBalance服务
```yaml
apiVersion: v1
kind: Service
metadata:
  name: kubia-loadbalancer
spec:
  type: LoadBalancer # 该服务会从k8s集群的基础架构中获取负载均衡器
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: kubia
```

```shell
[dairongpeng@dev workspace]$ k create -f loadbalance-service-test.yaml
service/kubia-loadbalancer created
# 分配的负载均衡的IP为localhost，访问localhost即可达到对该服务负载均衡访问的方式
[dairongpeng@dev workspace]$ k get svc kubia-loadbalancer
NAME                 TYPE           CLUSTER-IP      EXTERNAL-IP   PORT(S)        AGE
kubia-loadbalancer   LoadBalancer   10.98.254.197   localhost     80:30666/TCP   14s
[dairongpeng@dev workspace]$
```

> 负载均衡器负责转发流量到各个节点的NodePort，且是负载均衡的方式转发，被转发到的NortPort再去转发到集群的service，集群service再把流量转发到自己代理的其中一个pot（从Endpoints中选一个）



服务对于集群外部被负载均衡访问到的流量，只转发到和被负载均衡转发的Node同节点的Pod上，避免远距离流量转发，增加接口响应时间，可以使用该配置`spec.externalTrafficPolicy: Local`，因为在分配到Node之前，已经负载均衡了，service无需再对后端Pod随机转发了。注意：如果配置了该选项，且被转发的Node节点没有代理的pod，则service不进行转发了，所以请确保负载均衡器将流量转发给的节点，至少存在该Pod

> 使用LoadBalancer存在缺点，如果希望避免Node访问远距离Pod，一般加上`spec.externalTrafficPolicy: Local`，此时访问的负载完全是由于LoadBalancer转发到不同的Node来决定的，假设我们有两个Node, 三个被Service代理的Pod，NodeA分配了PodA, NodeB分配了PodB、PodC，这样的方式，会造成PodA拥有50%的流量，PodB和PodC各拥有25%的流量，这样显然不是合理的。

## Ingress
为什么需要Ingress资源呢，一个重要的原因是每个Loadbalancer服务都需要自己的负载均衡器，以及独有的IP地址，而Ingress资源，只需要一个共有IP地址，就能为多个服务提供访问，当客户端向Ingress发送HTTP请求时，Ingress会根据请求的主机名和路径决定请求转发到的服务。且由于Ingress工作在网络第七层，可以提供一些服务Service（网络第四层）不能实现的功能，例如Cookie。

Ingress控制器是Ingress资源在k8s集群中运行的前提，不同的k8s版本可能会有不同的控制器实现，甚至有些并不提供默认的Ingress控制器。

### Ingress资源定义
> 在运行Ingress之前，需要先检查Ingress控制器默认是否存在。

```shell
# 不存在Ingress资源控制器，需要安装一个。参考：https://github.com/kubernetes/ingress-nginx/blob/main/docs/deploy/index.md
[dairongpeng@dev workspace]$ kubectl get pod --all-namespaces | grep ingress
[dairongpeng@dev workspace]$ kubectl apply -f https://raw.githubusercontent.com/kubernetes/ingress-nginx/controller-v0.41.2/deploy/static/provider/cloud/deploy.yaml
The connection to the server raw.githubusercontent.com was refused - did you specify the right host or port?
# 遇到网络问题，部署文件下载到本地
[dairongpeng@dev workspace]$ kubectl create -f ingress-controller-local.yaml
namespace/ingress-nginx created
serviceaccount/ingress-nginx created
serviceaccount/ingress-nginx-admission created
role.rbac.authorization.k8s.io/ingress-nginx created
role.rbac.authorization.k8s.io/ingress-nginx-admission created
clusterrole.rbac.authorization.k8s.io/ingress-nginx created
clusterrole.rbac.authorization.k8s.io/ingress-nginx-admission created
rolebinding.rbac.authorization.k8s.io/ingress-nginx created
rolebinding.rbac.authorization.k8s.io/ingress-nginx-admission created
clusterrolebinding.rbac.authorization.k8s.io/ingress-nginx created
clusterrolebinding.rbac.authorization.k8s.io/ingress-nginx-admission created
configmap/ingress-nginx-controller created
service/ingress-nginx-controller created
service/ingress-nginx-controller-admission created
deployment.apps/ingress-nginx-controller created
job.batch/ingress-nginx-admission-create created
job.batch/ingress-nginx-admission-patch created
ingressclass.networking.k8s.io/nginx created
validatingwebhookconfiguration.admissionregistration.k8s.io/ingress-nginx-admission created
# 如果遇到网络原因，等待资源的安装完成
[dairongpeng@dev workspace]$ kubectl get po -n ingress-nginx
NAME                                        READY   STATUS              RESTARTS   AGE
ingress-nginx-admission-create-x9ssp        0/1     Completed           0          76s
ingress-nginx-admission-patch-7k4ld         0/1     Completed           0          76s
ingress-nginx-controller-6b94c75599-vmwz9   0/1     ContainerCreating   0          76s
[dairongpeng@dev workspace]$ k get svc -n ingress-nginx
NAME                                 TYPE           CLUSTER-IP     EXTERNAL-IP   PORT(S)                      AGE
ingress-nginx-controller             LoadBalancer   10.110.26.8    localhost     80:30261/TCP,443:30590/TCP   97s
ingress-nginx-controller-admission   ClusterIP      10.97.172.81   <none>        443/TCP                      97s
# ingress-nginx-controller-6b94c75599-vmwz9 启动完成，说明ingress-controller启动成功
[dairongpeng@dev workspace]$ kubectl get po -n ingress-nginx
NAME                                        READY   STATUS      RESTARTS   AGE
ingress-nginx-admission-create-x9ssp        0/1     Completed   0          2m23s
ingress-nginx-admission-patch-7k4ld         0/1     Completed   0          2m23s
ingress-nginx-controller-6b94c75599-vmwz9   1/1     Running     0          2m23s
```

- Ingress资源
```yaml
# before Kubernetes 1.19, only v1beta1 Ingress resources are supported
# from Kubernetes 1.19 to 1.21, both v1beta1 and v1 Ingress resources are supported
# in Kubernetes 1.22 and above, only v1 Ingress resources are supported
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: kubia
spec:
  rules:
  - host: kubia.example.com # ingress会将kubia.example.com转发到对应的服务
    http:
      paths:
      - path: "/" # 将所有发往当前ingress的流量都进行转发
        pathType: Prefix
        backend:
          service:
            name: kubia-nodeport # 当前ingress资源需要转发到的服务名
            port:
              number: 80 # 当前ingress资源需要转发到的服务端口
```

```shell
[dairongpeng@dev workspace]$ cat ingress-test.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: kubia
spec:
  rules:
  - host: kubia.example.com
    http:
      paths:
      - path: "/"
        pathType: Prefix
        backend:
          service:
            name: kubia-nodeport
            port:
              number: 80
[dairongpeng@dev workspace]$ kubectl create -f ingress-test.yaml
ingress.networking.k8s.io/kubia created
# ADDRESS会延迟显示，因为Ingress控制器在幕后调配负载均衡器
[dairongpeng@dev workspace]$ k get ingresses
NAME    CLASS    HOSTS               ADDRESS   PORTS   AGE
kubia   <none>   kubia.example.com             80      67s
```

> 本地host增加`{ADDRESS} kubia.example.com`的DNS解析，即可使用域名访问Ingress资源代理，从而访问服务。`curl http://kubia.example.com`

### Ingress工作原理
cli（域名） -> dns（ip） -> Ingress Controller -> Ingress ->  Service -> Endpoints -> Pod

### Ingress资源代理多个Service
```yaml
# 其他无关项略
spec:
  rules:
  - host: kubia.example.com
    http:
      paths:
      - path: "/a"
        pathType: Prefix
        backend:
          service:
            name: service-a # 将kubia.example.com/a 请求转发到service-a服务
            port:
              number: 80
      - path: "/b"
        pathType: Prefix
        backend:
          service:
            name: service-b # 将kubia.example.com/b 请求转发到service-b服务
            port:
              number: 80         
```

也可以根据不同主机host来确定不同的转发规则：
```yaml
# 其他无关项略
spec:
  rules:
  - host: a.example.com
    http:
      paths:
      - path: "/"
        pathType: Prefix
        backend:
          service:
            name: service-a # 将a.example.com 请求转发到service-a服务
            port:
              number: 80
  - host: b.example.com
    http:
      paths:
      - path: "/"
        pathType: Prefix
        backend:
          service:
            name: service-a # 将b.example.com 请求转发到service-a服务
            port:
              number: 80
```

### Ingreee支持TLS
Ingress支持TLS，一般客户端和Ingress控制器之间的通信是加密的，而控制器和后端pod之间的通信则不是，只需要把证书和私钥附加到Ingree资源配置中。
准确来说，证书和私钥存储在kubernetes的Secret资源中。而Ingress的Ingress manifest引用Secret即可。

1. 创建私钥及颁布证书
2. 创建Secret资源，使用我们创建的私钥和证书
3. 创建Ingress资源，选择使用该Secret

```yaml
# 其他无关项略
spec:
  tls: # 包含Ingress tls相关配置
  - hosts:
    - a.example.com # 允许接受主机a.example.com的tls链接
    secretName: tls-secret # 从tls-secret中获取之前创建的私钥和证书配置
  rules:
  - host: a.example.com
    http:
      paths:
      - path: "/a"
        pathType: Prefix
        backend:
          service:
            name: service-a # a.example.com 请求转发到service-a服务
            port:
              number: 80
```

## Service和Ingress背后Pod的就绪探针
Pod存在两种探针，存活探针和就绪探针。容器的就绪探针返回成功，则表示容器已经准备好接受外来请求。Pod就绪探针需要程序研发人员判断何时返回OK。

就绪探针也是在启动容器后一定时间，开始探测。与存活探针不同，就绪探针中如果容器没通过就绪检查，不会强制退出。

为容器配置就绪探针在配置项`spec.template.spec.containers`下，例如：
```yaml
readinessProbe:
  exec:
    commond:
    - ls
    - /var/ready
```

应该始终为容器定义就绪探针，即使就是很简单的http get "ok"，因为当我们的应用程序启动需要一定的时间时，这时的客户端流量没有就绪探针来控制，会打向未就绪的程序，从而引起大量请求报错。
同样的，在程序退出的时候，明确的把就绪探针置为失败的状态。

## kubernetes服务问题排查
如果无法通过服务访问服务后端的Pod：
- 查看集群内部是否可以访问该服务
- 不要通过ping服务的IP来排查问题（服务的集群IP是虚拟IP，ping不通）
- 使用就绪探针，如果就绪探针未返回OK，确保该POD不会被Service代理
- 查看服务的Endpoints中是否明确包含想要访问的容器（POD）
- FQDN访问不通（服务名称.命名空间.svc.cluster.local访问或其他简写方式），试一下直接通过集群IP访问是否连通、
- 检查客户端访问的端口，是服务暴露的端口，而不是服务代理的目标端口（TargetPort）
- 集群内部直接访问Pod IP及服务端口，确认Pod上访问是正常的
- 如果通过Pod IP也无法访问，确认容器端口是否和主机端口存在映射