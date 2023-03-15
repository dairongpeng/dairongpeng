## Kubernetes组件

### Master控制节点
- ApiServer：称为Api服务器，负责和各组件进行通信
- Scheculer：调度应用，为应用的每个可部署的节点分配一个工作节点
- Controller Manager：集群级别，处理例如复制组件，持续跟踪工作节点，处理失败节点等。
- Etcd：分布式kv存储，持久化集群配置。

一般来说控制节点Master仅负责控制集群状态，不扮演Node工作节点的角色，即不运行应用程序，应用程序由工作节点承载。

### Node工作节点
- ContainerD: 容器运行时，一般是Docker/rtk或其他容器类型，提供容器运行环境
- Kubelet：与控制节点中的ApiServer组件通信，管理所在节点的容器
- Kube-Proxy：负责组件之间的负载均衡网络流量

例如，当我们在本地开发机上调用创建Pod的命令时`kubectl run nginx --image=nginx --restart=Never`，本地kubectl会转化为post请求，请求k8s master的apiserver，
master节点上的调度器，会去通知一个Node节点上的kubelet，请求把该pod创建在该node上，当前node节点的kubelet负责调度本node上的容器运行时（Docker）负责拉取运行镜像，
至此一个pod就被创建在了其中一个Node节点上。
