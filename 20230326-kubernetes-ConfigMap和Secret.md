## 通过资源配置传递配置到容器
```yaml
kind: Pod
spec:
    containers:
    - image: xxx
      command: ["/bin/xxx"]
      args: ["arg1", "arg2", "arg3"]
```

> 在k8s资源定义中，存在该对应关系。command对应Docker的ENTRYPOINT（容器中运行的可执行文件），args对应Docker的CMD（传给可执行文件的参数）

## 通过为容器添加环境变量传递配置
```yaml
kind: Pod
spec:
    containers:
    - image: luksa/forture:env
      env: # 为容器预置环境变量，注意该环境变量不是pod级别
      - name: FIRST_ENV_VAR
        value: "ab"
      - name: SECOND_ENV_VAR
        vaule: "$(FIRST_ENV_VAR)c" ## 后续环境变量可以使用前面的环境变量，这里为abc
```

> 除此之外，k8s会自动暴露相同命名空间下每个service对应的环境变量，属于自动注入的，达到pod自动发现service的目的。

## ConfigMap
配置与Pod分离，进行解耦，把配置选项分离到单独的资源对象ConfigMap中。ConfigMap本质上就是一个键值对迎着。其中值可以是短字面量，也可以是完整的配置文件。

应用感知不到ConfigMap的存在，ConfigMap中的键值对通过环境变量或者卷文件的形式传递给容器。

### ConfigMap定义
```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: my-configmap # ConfigMap的名字
data: # ConfigMap中存储的数据，"|"表示将数据项中的多行文本转化为字符串，"-"符号表示将第一行和最后一行的空格和换行符去掉
  app.config: |-
    server.port=8080
    db.host=localhost
    db.port=3306
  app.env: |-
    SPRING_PROFILES_ACTIVE=dev
    LOG_LEVEL=debug
```

### ConfigMap的使用
在容器中使用ConfigMap：
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
  - name: my-container
    image: my-image
    envFrom: # 通过envFrom引用ConfigMap, 实现该Config的配置全部注入
    - configMapRef:
        name: my-configmap
```

在容器中使用ConfigMap中的某一配置项：
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: my-pod
spec:
  containers:
  - name: my-container
    image: my-image
    env: # 通过env，单个引入Config中的配置项, 需要什么配置项引入什么配置项
    - name: SERVER_PORT
      valueFrom:
        configMapKeyRef:
          name: my-configmap
          key: app.config
    - name: DB_HOST
      valueFrom:
        configMapKeyRef:
          name: my-configmap
          key: app.config
    - name: LOG_LEVEL
      valueFrom:
        configMapKeyRef:
          name: my-configmap
          key: app.env
```

### ConfigMap当成卷挂载到容器内
> ConfigMap的卷，需要被挂载到文件夹/etc/nginx/conf.d下让程序使用，也可以使用items配置项，选择configMap中的哪些项被挂载到卷中
```yaml
spec:
  containers:
  - name: my-container
    image: my-image
    volumeMounts:
    ...
    - name: config
      mountPath: /etc/nginx/conf.d
      readOnly: true
    ...
  volumes:
  ...
  - name: config
    configMap:
      name: my-configmap
```

将ConfigMap暴露为卷，可以达到应用热更新的效果，无需新建Pod或重启容器。ConfigMap被更新，卷中引用它的所有文件也会相应更新，进程发现文件被改变之后进行重载。

> Config通常被用来存储非敏感的数据，敏感数据一般使用Secret资源。

## Secret
与ConfigMap类似，Secret也可以：
- 将Secret条目作为环境变量传递给容器
- 将Secret条目暴露为卷中的文件

注意：Secret只会存储在节点的内存中，不会写盘。

### 默认令牌Secret
每个Pod都被默认挂载一个令牌的Secret，通过k describe pod可以看到。该Secret对应成的卷被挂在容器的/var/run/secrets/kubernetes.io/serviceaccount/下。
```shell
kubectl exec mypod ls /var/run/secrets/kubernetes.io/serviceaccount/
ca.crt
namespace
token
```

### Secret资源定义
```yaml
apiVersion: v1
stringData:
  foo.string: plain text
data:
  foo: YmFyCg==
  https.cert: LSOtLSlCRUdJTiBDRVJUSUZJQOFURSOtLSOtCklJSURCekNDQ...
  https.key: LSOtLSlCRUdJTiBDRVJUSUZJQOFURSOtLSOtCklJSURCekNDQ...
kind: Secret
...
```

一般Secret数据条目会被Base64格式编码，特殊需要纯文本展示需要使用stringData配置，而ConfigMap直接以纯文本展示。由于Secret存在内存的原因，一个Secret大小限制为1MB。

stringData是只写的，当我们写入stringData，再查该Secret资源时，看不到stringData中的内容，转而生成一条被加密的条目。

### Pod中使用Secret
通过Secret卷把Secret暴露被容器后，Secret条目会被转化为环境变量，且被解码以真实的形式写入对应的文件。作为应用程序无需解码，直接读取对应的文件或者环境变量即可

1. 创建一个包含密码的文件，比如 `password.txt`，并将密码写入文件中。
2. 执行以下命令创建 Secret：
```shell
kubectl create secret generic db-password --from-file=password.txt
```

这将创建一个名为 db-password 的 Secret，其中包含了名为 password.txt 的文件中的密码。

接下来，我们需要将这个 Secret 挂载到一个容器中。这里我以一个简单的 Nginx 容器为例：
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: nginx-pod
spec:
  containers:
    - name: nginx
      image: nginx
      volumeMounts:
        - name: db-password
          mountPath: /etc/db-password
          readOnly: true
  volumes:
    - name: db-password
      secret:
        secretName: db-password
```

这个 Pod 包含一个名为 nginx 的容器，它将 `db-password` Secret 挂载到 /etc/db-password 目录下，并且设置为只读模式。

最后，我们可以在容器中访问这个 Secret 中的密码。比如，我们可以在容器中执行以下命令查看密码：
```bash
cat /etc/db-password/password.txt
```