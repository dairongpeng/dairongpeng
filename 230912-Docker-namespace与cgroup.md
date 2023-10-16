## Docker用到的一些技术
Docker容器本质上是宿主机上的进程。Docker通过namespace实现了资源隔离，通过cgroups实现了资源限制，通过写时复制机制（copy-on-wite）实现了高效的文件操作。

### Namespace资源隔离
- 容器拥有独立的文件系统，所以需要资源隔离
- 为了在分布式的环境下进行通信和定位，容器也必须要有独立的IP，端口，路由等，所以需要网络隔离
- 容器需要主机名
- 容器间通信需要隔离
- 容器的用户和用户组权限隔离
- 容器中的进程PID需要和宿主机隔离

对应linux的6项namespace的隔离：

|  namespace   | 系统调用参数      | 隔离内容                 |
|  ---         | ---             | ---                     |
| UTS          | CLONE_NEWUTS    | 主机名与域名              |
| IPC          | CLONE_NEWIPC    | 信号量，消息队列和共享内存  |
| PID          | CLONE_NEWPID    | 进程编号                 |
| Newwork      | CLONE_NEWNET    | 网络设备，网络栈，端口等    |
| Mount        | CLONE_NEWNS     | 挂载点（文件系统）         |
| User         | CLONE_NEWUSER   | 用户和用户组              |

通过namespace构建一个相对隔离的环境（shell环境也可以称作容器），处在同一个namespace下的进程，可以彼此感知，而当前namespace内的整体对外界进程无感知。

### Cgroup资源限制
Cgroups是Linux内核提供的一种机制，这种机制可以根据需求把一系列系统任务及子任务整合（或分割）到按资源划分等级的不同组内，从而为系统资源管理提供一个统一的框架。

Cgruops（control groups）不仅可以限制被namespace隔离起来的资源，还可以为资源设置权重，计算使用量，操控任务（进程或线程）启停等。本质上来说，cgroup是内核附加在程序上的一系列钩子（hook），通过程序运行时对资源的调度触发相应的钩子达到资源追踪和限制的目的。

Docker daemon会在单独挂载了每个子系统的控制组目录（例如/sys/fs/cgroup/cup）创建一个docker控制组，在docker控制组里面，再为每个容器创建一个以容器ID为名称的容器控制组，这个容器的所有进程号都会写入该控制组的tasks中，并且在控制文件（比如cpu_cfs_quota_us）中写入预设的限制参数值。例如：

```shell
tree cgroup/cpu/docker
cgroup/cpu/docker/
|--{容器ID}
|   |--cgroup.clone_children
|   |--cgroup.procs
|   |--cpu.cfs_quota_us
|   |--cpu.rt_rt_runtime_us
......
```

### copy-on-wite
Docker镜像使用的一种策略，叫写时复制。在多个容器之间共享镜像，每个容器在启动的时候，并不需要单独复制一份镜像文件，而是将所有镜像层以只读的方式挂载到一个挂载点，再在上面覆盖一个可读写的容器层。在未更改文件内容时，所有容器共享同一份数据，只有在Docker容器运行过程中文件系统发生变化时，才会把变化的文件内容写到可读写层，并隐藏只读层中的老版本文件。写时复制配合镜像分层机制，减少了镜像对磁盘空间的占用和容器启动时间。