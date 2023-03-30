## 卷（Volume）
与Docker概念中的卷（Volume）类似，其中kubernetes中卷的类型很多，常用的一般为：
- emptyDir: 用于存储临时数据的简单空目录。
- hostPath: 用于将目录从工作节点的文件系统挂载到pod中。
- gitRepo: 通过Git仓库内容类初始化卷。
- nfs: 挂在到pod中的NFS共享卷。
- configMap、secret、downwardAPI：用于将k8s部分资源信息挂载到Pod中。
- persistenVolumeClaim: 一种预置或者动态配置的持久存储类型。

其中单个容器，可以选择使用或不使用卷，也可以同时使用不同类型的卷。configMap、secret、downwardAPI资源主要是将k8s元数据公开给pod中的程序。

### emptyDir卷
该卷从一个空目录开始，生命周期和使用它的pod相同，当删除pod时，该卷会丢失。emptyDir对于一个pod中多个容器的文件共享，会非常有用。也可以用于单个pod单个容器数据的临时写盘，比如大量数据的排序，内存不够，临时选择写磁盘。由于有些容器的文件系统可能是不可写的，所有有些时候只能选择写emptyDir存储卷。

```yaml
apiVersion: v1
kind: Pod
metadata:
    name: fortune
spec:
    containers:
    - image: luksa/fortune
        name: html-generator # pod中的容器1
        volumeMounts:
        - name: html # 使用的卷名
          mountPath: /var/htdocs # 该卷挂载到Pod的/var/htdocs路径，当前容器会往/var/htdocs路径下写入一些文件
    - image: nginx:alpine
        name: web-server # pod中的容器2
        volumeMounts:
        - name: html # 使用的卷名
          mountPath: /usr/share/nginx/html # 该卷挂载到pod的/usr/share/nginx/html路径。该容器会读取路径
          readOnly: true # 设为只读（作用到该容器）
    ports:
    - containerPort: 80
      protocol: TCP
volumes:
- name: html # 该pod中的所有容器都可以共用该卷
  emptyDir: {}
```

emptyDir存储卷是在承载该pod的集群节点上的实际磁盘上创建的，因此该卷的读写性能取决于该节点上的磁盘类型。如果我们要基于节点的内存创建该卷那么配置如下：
```yaml
volumes:
- name: html # 该pod中的所有容器都可以共用该卷
  emptyDir: {}
    medium: Memory
```

## gitRepo存储卷
gitRepo卷类似于emptyDir卷。本质就是使用git的repo来作为挂载点，但是一般来说创建gitRepo卷后，Pod并不能及时的和repo保持同步，如果我们使用RC管理Pod,可以使用“减一加一”，新创建出来的Pod就会和gitRepo同步了。一般使用该卷，还是倾向于gitRepo中存储的是静态资源，一般不更改，若更改，需要使用控制器销毁pod重新创建，来保持同步。
```yaml
volumes:
- name: html # 该pod中的所有容器都可以共用该卷
  repository: https://www.github.com/dairongpeng/gitRepo-volume-example.git
  revision: master # 检出主分支
  directory: . # 将repo克隆到卷的根目录
```

注意，gitRepo卷也非持久卷，随着pod的销毁，卷及其内容也会被删除。

### hostPath卷
一般来说，大多数pod应该选择不依赖自身主机节点的文件系统。但是某些系统级别的pod（通常由DaemonSet管理）确实需要读取节点的文件系统来访问节点设备等。k8s提供了hostPath卷，来实现这一点。

hostPath卷，是持久卷。如果在删除了pod，新创建的pod仍指向相同的挂载目录，则新pod会发现上一个pod留下的数据文件。前提是，新启动的pod和上个被销毁的pod，同在一个主机节点上。

所以使用hostPath卷，依赖节点，新旧pod不在一个节点则新pod还是找不到数据，所以使用hostPath卷之前需要慎重考虑。一般DaemonSet确保每个节点分布一个Pod, 可以尝试。

### 其他共享卷
由于hostPath持久卷的问题，要保证任何集群节点访问持久存储都没问题，则需要考虑使用网络存储共享卷（eg:NAS）。

例如使用NFS卷：
```yaml
volumes:
- name: mongodb-data
  nfs: # 该卷受nfs共享支持
    server: 1.2.3.4 # nfs服务器的ip
    path: /some/path #  服务器提供的路径
```

注意： 理论上选择何种外部存储对应的持久卷，决定权应该留给集群管理员，基础设施相关的同学。但就目前来说，在Pod配置中直接选择要使用何种卷，并未达到解耦的作用，研发人员应该只需要告诉pod应该使用卷，至于卷的实现，可以存在各种，这就是后续的卷的接口隔离思想（持久卷声明）。

## 持久卷（PersistentVolume 简:PV）和持久卷声明（PersistentVolumeClaim 简：PVC）
当我们需要持久化存储时，首先声明PVC，标识出需要的最低容量和访问模式，交给k8s，k8s寻找到可匹配的持久卷，并绑定到该持久卷的声明上。可以类比PVC为定义接口，接口的实现PV由k8s决定。

### 创建持久卷（基础设施角色）
以mongodb举例：

```yaml
apiVersion: v1
kind: PersistentVolume
metadata:
    name: mongodb-pv
spec:
    capacity: # 定义该PV的大小
        storage: 1Gi
    accessModes:
    - ReadWriteOnce # 可以被单个客户端挂载为读写模式
    - ReadOnlyMany # 可以被多个客户端挂载为只读模式
    persistentVolumeReclaimPolicy: Retain # 当声明被释放后，pv将被保留（不删除和清理）
    gcePersistentDisk: # pv指定支持GCE持久磁盘
        pdName: mongodb
        fsType: ext4
```

通过`k create -f [持久卷yaml]`，持久卷被当作资源创建在K8s中，可以通过`k get pv`查看。

注意：持久卷不属于任何名称空间，是集群层面的资源，整个集群可以使用。

### 创建持久卷声明（交付研发角色）
创建持久卷声明：

```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
    name: mongodb-pvc
spec:
    resources:
        requests:
            storage: 1Gi # 申请1GB的存储空间
    accessModes:
    - ReadWriteOnce # 允许单个客户端访问（支持读写）
    storageClassName: "" # 用户控制该pvc需要使用的存储类，将空字符串指定为存储类名可确保PVC 绑定到预先配置的 PV, 而不是动态配置新的PV
```

通过`k create -f [持久卷声明yaml]`，持久卷声明被当作资源创建在K8s中，可以通过`k get pvc`查看。与pv不同，持久卷声明pvc创建需要指定命名空间。

k8s会自动为我们绑定pv, 查看pvc时，AccessModes一般为：
- RWO: ReadWriteOnce, 允许单个节点挂载读写
- ROX: ReadOnlyMany, 允许多个节点挂载只读
- RWX: ReadWriteMany, 允许多个节点挂载读写该卷

再次查看pv`k get pv`，发现上文创建的pv的STATUS状态由Availavble变为Bound，限制已经被绑定使用了。

### Pod中使用持久卷声明
```yaml
spec:
    containers:
    - image: mongo
      name: mongodb
      volumeMounts:
      - name: moungodb-data
        mountPath: /data/db
      ports:
      - containerPort: 27017
        protocol: TCP
    volumes:
    - name: mongodb-data
      persistentVolumeClaim: # 在pod的资源配置定义中，通过pvc的名称，声明要使用的pvc，需要保证pvc和当前pod在同一个ns
      claimName: mongodb-pvc
```

### 持久卷的释放
当我们删除pod，删除相应的pvc，那么此时pv显示的STATUS时Released，不在允许被其他pvc进行绑定。基于数据安全的考虑，也是不应该直接让一个新的声明，使用之前的存储。

可以选择手动清理pv，或者自动清理pv，略。达到pv可以被反复利用的效果。

## StorageClass
用户处理指定pvc，并且也在Pvc中引用了StorageClass。集群管理员负责提前创建一些StorageClass，再创建一些持久卷。

```yaml
apiVersion: storage.k8s.io/v1
kind: StorageClass
metadata:
    name: fast
provisioner: kubernetes.io/gce-pd # 用于配置持久卷的插件
parameters: # 传递parameters的参数
    type: pd-ssd
    zone: europe-west1-b
```

可使用`k get sc`来查看StorageClass资源。

在pvc中指定存储类：
```yaml
apiVersion: v1
kind: PersistentVolumeClaim
metadata:
    name: mongodb-pvc
spec:
    resources:
        requests:
            storage: 1Gi # 申请1GB的存储空间
    accessModes:
    - ReadWriteOnce # 允许单个客户端访问（支持读写）
    storageClassName: fast # !! 该pvc请求自定义存储类
```

StorageClass的好处：声明是通过名称引用它们的，因此只要StorageClass名称在所有这些名称中相同，PVC就可以跨不同的集群进行移植，队集群的迁移大有好处。