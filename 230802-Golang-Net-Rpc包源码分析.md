`远程过程调用 (RPC)`是提供操作系统中使用的高级通信范例的协议。 RPC 假定存在低级别传输协议，例如传输控制协议/Internet Protocol (TCP/IP) 或用户数据报协议 (UDP) ，用于在通信程序之间传输消息数据。 RPC 实现专为支持网络应用程序而设计的逻辑客户机到服务器通信系统。

`RPC`框架即`远程调用框架（Remote Procedure Call）`。远程过程调用可以理解为，被调用函数不在本地，而是在服务端。通过网络请求，客户端调用服务端的某个方法，服务端把方法结果，返回给客户端。`RPC`框架也非常多，比如google的grpc，百度的brpc, 新浪的rpcx等等。大家可以根据情况选择。实际上在`golang`标准库中就存在rpc的简单实现，叫做`jsonrpc`。

## gorpc服务端
`jsonrpc`服务端，有两个入口，第一个是直接`net/rpc`包级别函数启动服务，使用的是`DefaultServer`, 第二个是自己通过`rpc.NewServer`获取一个自定义的`rpc.Server`结构，通过调用该结构对应的结构体方法启动服务。实际上包级别的函数也是通过`DefaultServer`调用自身的结构体函数，这一点需要提前了解下。所以我们只需要分析`rpc.Server`的结构体方法即可。

```golang
// Server represents an RPC Server.
type Server struct {
    // 类似与路由信息，key如果不指定, 就是server.Register(new(MyService))中MyService的名字, 如果指定不能重复；
    // value是是通过MyService反射加工出来的信息，比如该结构对应多少结构体方法，结构体TypeOf和ValueOf
	serviceMap sync.Map   // map[string]*service
    
    // reqLock 和 freeReq 用于保护空闲的请求（Request）对象的互斥访问。
    // 在处理 RPC 请求时，需要从一个可用的 Request 对象池中获取一个请求对象，以减少内存分配的开销。
    // reqLock 用于对 freeReq 进行加锁，以确保在多个 goroutine 同时获取请求对象时不会发生竞争条件。
	reqLock    sync.Mutex // protects freeReq
    // 在处理 RPC 请求时，通过调用 server.getRequest 方法可以获取一个空闲的 Request 对象。
    // 在处理完请求后，可以通过调用 server.freeRequest 方法将请求对象放回对象池，以便其他请求可以继续复用它。
	freeReq    *Request
	
    // respLock 和 freeResp 用于保护空闲的响应（Response）对象的互斥访问。
    // 与请求对象一样，响应对象也需要进行复用，以减少内存分配的开销。
    respLock   sync.Mutex // protects freeResp
    // 在处理 RPC 请求时，通过调用 server.getResponse 方法可以获取一个空闲的 Response 对象。
    // 在处理完请求后，可以通过调用 server.freeResponse 方法将响应对象放回对象池，以便其他请求可以继续复用它。
	freeResp   *Response
}
```

`rpc.Server`的一些方法：

```golang
// 注册服务，写入到serviceMap字段，key默认为结构体名字，value默认是rcvr的一些反射信息。
func (server *Server) Register(rcvr any) error

// 注册服务，自定义服务的key名称，其他流程和Register一样
func (server *Server) RegisterName(name string, rcvr any) error

// 通过ServeConn来启动服务端监听，默认使用gobServerCodec编码器，来解码请求Request和编码Response
// ServeConn 方法用于处理单个连接上的连续 RPC 请求。它接受一个 net.Conn 类型的参数，表示一个网络连接。
// 该方法会持续地从连接中读取请求，并通过注册的服务和方法来处理这些请求。它会在连接关闭之前一直阻塞，用于处理来自客户端的连续请求。
// 使用场景：当你希望在一个长期保持的连接上连续处理多个 RPC 请求时，可以使用 ServeConn 方法。例如，在一个长连接的 WebSocket 连接上提供 RPC 服务。
func (server *Server) ServeConn(conn io.ReadWriteCloser)

// 通过ServeCodec来启动服务端监听，自定义编码器，来解码请求和编码响应。并发处理请求
// 与ServeConn的区别就在于一个是自定义编码器，一个使用默认的编码器。
func (server *Server) ServeCodec(codec ServerCodec)

// 通过ServeRequest来启动服务端监听，自定义编码器，来解码请求和编码响应。与ServeCodec类似，区别是串行处理请求
// ServeRequest 方法用于处理单个 RPC 请求。它接受一个 rpc.Request 类型的参数，表示一个 RPC 请求。这个方法在处理完请求后会立即返回，而不会等待下一个请求。
// 使用场景：当你希望独立地处理单个 RPC 请求时，可以使用 ServeRequest 方法。这种方式适用于批处理或在特定时间处理单个请求。
func (server *Server) ServeRequest(codec ServerCodec) error

// 服务监听net.Listener的链接，并为获取到的链接请求，启动一个go程，处理。
func (server *Server) Accept(lis net.Listener)
```

关于`ServeConn`, `ServeCodec`, `ServeRequest`: 如果你希望持续处理多个请求并在连接关闭之前保持连接，可以使用 ServeConn。如果你希望自定义编解码器并处理请求，可以使用 ServeCodec。如果你需要独立地处理单个请求，可以使用 ServeRequest。根据具体的需求，你可以选择最适合的方法来实现 RPC 服务器的功能。

### 构建rpc服务端
```golang
// 定义Service
type Service int
// Service提供给外部调用的方法
// 注意：
// 1. 服务必须可以导出，即Service的首字母S要大写
// 2. 服务提供的方法可以导出，即SayHello的首字母S要大写
// 3. 服务端提供的可被调用的方法，定义的参数要固定，形如：MethodName(req Any, resp *Any) error； 其中resp必须为指针类型
// 如果违背以上原则，可能会得到"rpc.Register: type Service has no exported methods of suitable type"的问题，显示服务方法不能被识别注册
func (*Service) SayHello(name string, output *string) error {
	*output = fmt.Sprintf("hello %s", name)
	return nil
}

func main() {
    // 获取一个rpc Server结构
	server := rpc.NewServer()
	// 注册rpc服务
	svc1 := new(Service)
	server.Register(svc1)

    // 服务监听在哪个ip:port, 默认是本地ip
	l, err := net.Listen("tcp", ":1024")
	if err != nil {
		log.Printf("Error: net.Listen err=%v", err.Error())
		panic(err)
	}

    // 使用rpc库自带的Accept，使用默认的编码解码器。
    // 会监听请求，开go程处理
	server.Accept(l)
}
```

## gorpc客户端
`rpc`原生客户端比较简单，源码也就300多行。客户端对应一些关键的方法介绍：

```golang
// 获取一个rpc客户端， 传入一个链接，并且使用默认的编码解码器gobClientCodec
func NewClient(conn io.ReadWriteCloser) *Client

// 获取一个rpc客户端，传入定制化的编码解码器，编码解码器中携带了conn信息。例如json编码解码器等。
func NewClientWithCodec(codec ClientCodec) *Client

// 获取一个rpc客户端，传入协议和地址，方法内部会创建conn，直接获取到Client, 使用默认编码解码器
func Dial(network, address string) (*Client, error)

// 关闭客户端链接
func (client *Client) Close() error

// client请求服务端，获得一个*Call的回调，当*Call.Done信号产生时，请求完毕。可以理解为异步客户端调用
func (client *Client) Go(serviceMethod string, args any, reply any, done chan *Call) *Call

// client请求服务端，获得一个错误信息。Call内部实质上还是调用Go，只是显示的等待*Call.Done信号，判断成功是否，返回给客户端调用处，可以理解为同步客户端调用
func (client *Client) Call(serviceMethod string, args any, reply any) error
```

一个客户端调用服务端的例子：

```golang
func main() {
    // 配置rpc服务客户端需要请求的服务端地址，获取链接
	conn, err := net.Dial("tcp", "127.0.0.1:1024")
	if err != nil {
		panic(err)
	}

	req := "gorpc"
	var resp string

    // 获取一个rpc客户端，使用默认的编码解码器，需要扩展编码解码器可以使用rpc.NewClientWithCodec(customerCodec), 
    // 注意：需要服务端也使用customerCodec编码解码器
	client := rpc.NewClient(conn)
    // 调用rpc服务，Service服务的，SayHello方法。Service需要在rpc服务端注册
	err = client.Call("Service.SayHello", req, &resp)
	if err != nil {
		panic(err)
	}

	fmt.Println(resp)

    // Output:
    // hello gorpc
}
```

## jsonrpc
上面golang原生的rpc调用内容，已经可以达到客户端和服务端的通信，但是上文提到过很多次编码解码器，那么编码解码器到底是个什么东西。为什么会出现这个概念呢？编码解码器，就是服务端和客户端约定按何种方式传输数据，如果以一种通用的格式数据传输数据，那么客户端和服务端就可以不用关心各自是什么语言实现的了。

需要使用jsonrpc也比较简单，golang标准库里面已经提供了相应的编码解码器，在`rpc/jsonrpc`包下, 简单改造下服务端，让其称为json编码器的rpc服务。

```golang
// 定义Service
type Service int
// Service提供给外部调用的方法
// 注意：
// 1. 服务必须可以导出，即Service的首字母S要大写
// 2. 服务提供的方法可以导出，即SayHello的首字母S要大写
// 3. 服务端提供的可被调用的方法，定义的参数要固定，形如：MethodName(req Any, resp *Any) error； 其中resp必须为指针类型
// 如果违背以上原则，可能会得到"rpc.Register: type Service has no exported methods of suitable type"的问题，显示服务方法不能被识别注册
func (*Service) SayHello(name string, output *string) error {
	*output = fmt.Sprintf("hello %s", name)
	return nil
}

func main() {
    // 获取一个rpc Server结构
	server := rpc.NewServer()
	// 注册rpc服务
	svc1 := new(Service)
	server.Register(svc1)

    // 服务监听在哪个ip:port, 默认是本地ip
	l, err := net.Listen("tcp", ":1024")
	if err != nil {
		log.Printf("Error: net.Listen err=%v", err.Error())
		panic(err)
	}

   	// 与原生Server不同，这里自己实现监听接入服务端的链接！
	for {
		conn, err := l.Accept()
		if err != nil {
			log.Println("Warn: get conn err")
			time.Sleep(100 * time.Millisecond)
			continue
		}

		// 启动一个协程去处理当前连进来的链接，使用jsonrpc的编码解码器
		go server.ServeCodec(jsonrpc.NewServerCodec(conn))
	}
}
```

客户端改写成使用jsonrpc的客户端：

```golang
func main() {
    // 配置rpc服务客户端需要请求的服务端地址，获取链接
	conn, err := net.Dial("tcp", "127.0.0.1:1024")
	if err != nil {
		panic(err)
	}

	req := "gorpc"
	var resp string

    // 这里使用jsonrpc.NewClientCodec标识，客户端使用jsonrpc的编码解码方式
	client := rpc.NewClientWithCodec(jsonrpc.NewClientCodec(conn))
    // 调用rpc服务，Service服务的，SayHello方法。Service需要在rpc服务端注册
	err = client.Call("Service.SayHello", req, &resp)
	if err != nil {
		panic(err)
	}

	fmt.Println(resp)

    // Output:
    // hello gorpc
}
```

改写成jsonrpc的好处是，如果我们服务端是用Golang实现的，那么由于是一个通用的传输结构，服务端无论用什么语言，也用json数据格式传输rpc服务，就能够实现客户端和服务端的通信。这里由于json比较通用，我们先来检查一下当前客户端服务端交互中，网络传输的数据流到底是什么格式。方便我们定位到，如果其他语言客户端想要和我们的golang服务端通信，需要发送什么样的网络请求数据包。

```golang
// 服务端代码，在获取到conn后，调用这个方法，打印conn里面的数据流。
func readConnData(conn net.Conn) {
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		log.Printf("Server read returned error: %s", err)
		return
	}
	if n != 0 || err != io.EOF {
		log.Printf("Read = %v, %v, wanted %v, %v", n, err, 0, io.EOF)
	}

	requestBody := strings.TrimSpace(string(buf[0:n]))
	fmt.Printf("remoteAddr: %v | requestBody: %s", conn.RemoteAddr(), requestBody)

	conn.Write([]byte("ack!"))
	conn.Close()
}

// Output:
// remoteAddr: 127.0.0.1:52971 | requestBody: {"method":"Service.SayHello","params":["gorpc"],"id":0}
```

收到了请求包为：`{"method":"Service.SayHello","params":["gorpc"],"id":0}`, 也就是说，当我们从任何设备发送TCP请求到我们的服务端，携带的是`{"method":"Service.SayHello","params":["gorpc"],"id":0}`这个数据包，就可以调用的到我们服务端提供的rpc方法。通过观察源码，我们还可以发现，req的id字段，来源是客户端发送数据包的序列号（sequence），由于做了并发控制，支持并发。客户端发送请求后，把序列号给到req，后传递到服务端，服务端响应该数据包的时候，也要告诉客户端回应的是哪个序列号。

```golang
// rpc.Client
type Client struct {
    // ... 略 ...
	seq      uint64
    // ... 略 ...
}

// client的seq携带给了req的Seq
type Request struct {
    // ... 略 ...
	Seq           uint64   // sequence number chosen by client
    // ... 略 ...
}

// 在json rpc的编码解码器中，把Seq写入到了c.req.Id，通过网络发送到了服务端。
func (c *clientCodec) WriteRequest(r *rpc.Request, param any) error {
	c.mutex.Lock()
	c.pending[r.Seq] = r.ServiceMethod
	c.mutex.Unlock()
	c.req.Method = r.ServiceMethod
	c.req.Params[0] = param
	c.req.Id = r.Seq
	return c.enc.Encode(&c.req)
}
```

下面我们把服务端代码恢复成正常的rpc服务端代码，来做一些验证。

1. 先启动rpc服务端
2. 启动一个裸tcp客户端，不依赖rpc。即所有的编程语言都可以实现该客户端。
3. 启动客户端，获取网络链接，发送数据包`{"method":"Service.SayHello","params":["gorpc"],"id":0}`
4. 获取链接的响应，即我们rpc服务，给予该链接的响应。
5. 读取并打印链接的输出流

```golang
func main() {
    // 获取rpc服务所在服务端的链接
	conn, err := net.Dial("tcp", "127.0.0.1:1024")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

    // 往该网络链接中写入数据流
	_, err = conn.Write([]byte(`{"method":"Service.SayHello","params":["gorpc"],"id":0}`))
	if err != nil {
		println("Write data failed:", err.Error())
		os.Exit(1)
	}

    // 读取该链接的输出流
	received := make([]byte, 1024)
	_, err = conn.Read(received)
	if err != nil {
		println("Read data failed:", err.Error())
		os.Exit(1)
	}

    // 打印链接返回给tcp客户端的输出流
	fmt.Printf("responseBody: %s", string(received))
}

// Output:
// responseBody: {"id":0,"result":"hello gorpc","error":null}
```

## 总结
本篇，我们进入了`net/rpc`包中，观察了golang中原生`rpc`的实现，剖析了`rpc.Client`，`rpc.Server`结构。后面又演示了原生`rpc`调用方法，以及实现了跨语言通信`jsonrpc`的调用方法。如果个人项目希望实现`rpc`的调用完全可以使用原生的`rpc`或者`jsonrpc`。众所周知，`json`虽然是各个编程共同认可的结构，但是解析和传输效率上来说，确实有他的不足之处。由此，开源社区出现了很多的`rpc`框架，我理解大部分也都是在数据的编码解码上面做的文章，具体我并没有去研究。如果需要更高性能的`rpc`，大家也可以尝试更完善的`rpc`框架，例如`grpc`, `rpcx`, `brpc`等，更细节的选择问题，大家各自调研就好。