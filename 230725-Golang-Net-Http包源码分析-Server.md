在进入`Golang net/http`包细节之前，先简单了解下网络通信。关于`TCP/IP`协议中通信的五个要素被称为五元组。
- 源IP地址 (Source IP Address)：指发送数据包的计算机的IP地址。
- 目标IP地址 (Destination IP Address)：指接收数据包的计算机的IP地址。
- 源端口号 (Source Port)：指发送数据包的应用程序的端口号。
- 目标端口号 (Destination Port)：指接收数据包的应用程序的端口号。
- 传输协议 (Protocol)：指数据包所使用的传输层协议，通常是TCP或UDP。

这五个要素共同标识了一个网络连接的两个端点，确保数据包能够正确地从源地址发送到目标地址，并被正确地交付给相应的应用程序。

在Linux中，TCP/IP数据包的封装是通过套接字(Socket)来实现的。套接字是应用程序与网络通信之间的接口。当应用程序发送数据时，数据会通过套接字传递到TCP/IP协议栈进行封装，形成一个完整的数据包，然后通过网络传输到目标主机的协议栈，再经过解封装交付给目标应用程序。

具体封装过程如下：
1. 应用程序通过套接字发送数据。
2. 数据被传递到传输层协议栈，根据套接字中指定的传输协议（TCP或UDP）选择相应的协议处理函数。
3. 协议栈根据目标IP地址和端口号查找路由表，确定数据包的下一跳目标（可能是本地网络或网关）。
4. 数据包被传递到网络层协议栈，根据路由表确定数据包的下一跳地址，并加上IP头部（包含源IP地址和目标IP地址）。
5. 数据包进入数据链路层协议栈，加上数据链路层的头部和尾部，形成完整的数据帧，其中包含源MAC地址和目标MAC地址。
6. 数据帧通过网络接口发送到目标主机。

在目标主机上，相应的操作会逆序进行，进行解封装，将数据包从数据链路层一直还原到应用程序层，然后交付给目标应用程序。这样，通过五元组的封装和解封装过程，TCP/IP协议栈保证了应用程序之间的可靠通信。

`net/http`包内，`server`和`client`就对应一组CS。网络通信的底层实现已经封装在`net.Dial`和`net.Listen`等函数中，它们直接利用了操作系统提供的`socket`套接字进行通信。用户无需直接操作`socket`套接字，而是通过`http.Client`和`http.Server`等高级抽象来进行HTTP通信。由于`net/http`的Client和Server是更高级别的抽象，理所当然，Client发送请求时，通信目标通过url已经给定。Server也会注定自身监听在哪个端口上。

## 服务端Server
在官方的示例中，启动一个Http服务只需要几行代码。例如我们使用默认定义，启动一个HTTP服务。我们现在剖析这个最简单，没有特别配置的HTTP服务。
- 路由配置使用默认的`DefaultServeMux`，当我们`http.HandleFunc("/hello", sayHello)`时，会有默认路由，注册该处理函数。处理函数通过定义`func(w http.ResponseWriter, r *http.Request)`来实现。
- 可以一指定处理结构，需要接口实现`ServeHTTP(w http.ResponseWriter, r *http.Request)`方法，通过Handle函数注册到默认路由。`http.Handle("/world", &helloHandler{})`
- 监听`8080`端口，`http.ListenAndServe(":8080", nil)`, 传递nil会使用默认路由匹配。该函数会返回一个err，一般是阻塞监听客户端请求，处于挂起状态。err会在服务退出时出现返回值，正常退出会有特别的错误标识`http.ErrServerClosed`。

```golang
func sayHello(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "got /hello request\n")
}

func main() {
	http.HandleFunc("/hello", sayHello)
	http.Handle("/world", &helloHandler{})

	err := http.ListenAndServe(":8080", nil)
	if errors.Is(err, http.ErrServerClosed) {
		fmt.Printf("server one closed\n")
	} else if err != nil {
		fmt.Printf("error listening for server one: %s\n", err)
	}
}

type helloHandler struct{}

func (*helloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "got /hello request\n")
}
```

### 多路复用
路由注册使用`DefaultServeMux`，当我们通过http包指定Handle或者HandleFunc时，默认写入到默认的多路复用器中。在后续启动服务的阶段，如果我们指定自己的多路复用器，就传入mux，而不是默认传入nil即可。

在多路复用结构中，路由表是使用一个map结构实现的, 其中map的key是路由路径`path`即`pattern`，value是`muxEntry`, 含有`h`和`pattern`，在我们上文`Handle`和`HandleFunc`中实际上目的都是为了填充`h`，`h`是接口类型，含有`ServeHTTP(w http.ResponseWriter, r *http.Request)`方法。当我们使用函数注册路由时，该函数会强制转换为`HandlerFunc`函数类型，该类型也实现了`Handler`接口的`ServeHTTP`方法；当我们使用结构体填充路由时，也需要结构体实现`Handler`接口的`ServeHTTP`方法。`muxEntry`的`h`使用接口类型，很合理，从而实现了扩展。

我们自己实现多路复用的路由器时，也可以按照类似的方法，注册我们的路由，最终传递到Server中去。

```golang
type ServeMux struct {
	mu    sync.RWMutex
	m     map[string]muxEntry // 路由结构
	es    []muxEntry // slice of entries sorted from longest to shortest.
	hosts bool       // whether any patterns contain hostnames
}

type muxEntry struct {
	h       Handler
	pattern string
}

// 接口类型，如果通过函数注册路由，会强制转换为HandlerFunc类型，使得其实现ServeHTTP方法。就是本身的f调用。
type HandlerFunc func(ResponseWriter, *Request)

// ServeHTTP calls f(w, r).
func (f HandlerFunc) ServeHTTP(w ResponseWriter, r *Request) {
	f(w, r)
}
```

### 监听路由
上面最简单的http服务，处理路由注册，数`http.ListenAndServe`方法最为耀眼，`http.ListenAndServe`内发生了什么，使得其可以挂起，接受来自客户端的请求。

```golang
func ListenAndServe(addr string, handler Handler) error {
    // 初始化一个server结构，地址和端口使用传入的addr。路由器使用传入的多路复用器，如果传入空，使用默认的多路复用的路由器。
	server := &Server{Addr: addr, Handler: handler}
    // 获取到最基本的运行条件，调用ListenAndServe，启动服务，并监听在addr上。
	return server.ListenAndServe()
}
```

深入到`server.ListenAndServe()`内部，我们可以看到两个关键的方法，其一是调用封装的网络库`dial.go`，监听端口的方法`net.Listen("tcp", addr)`; 其二是通过第一步返回的监听器`Listener`启动我们的服务`srv.Serve(ln)`。

```golang
func (srv *Server) ListenAndServe() error {
	if srv.shuttingDown() {
		return ErrServerClosed
	}
	addr := srv.Addr
	if addr == "" {
		addr = ":http"
	}
    // 调用基于Linux Socket封装的网络库。这里屏蔽细节
	ln, err := net.Listen("tcp", addr)
	if err != nil {
		return err
	}
    // 启动服务。
	return srv.Serve(ln)
}
```

启动服务后，不停的监听`Listener`是不是有新的链接接入。由于我们这里监听的是TCP, 即如果有新的TCP链接进入，`Server`会启动一个`goroutine`来处理这次请求, 后续会查找这个接入请求要路由到多路复用器的哪个路由上。

```golang
	for {
        // 这里阻塞，当有新的连接加入时，这里会存在返回值。我们启动的是HTTP服务，这里返回的应该是TCP链接，TCPConn
		rw, err := l.Accept()
		if err != nil {
			if srv.shuttingDown() {
				return ErrServerClosed
			}
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				srv.logf("http: Accept error: %v; retrying in %v", err, tempDelay)
				time.Sleep(tempDelay)
				continue
			}
			return err
		}
		connCtx := ctx
		if cc := srv.ConnContext; cc != nil {
			connCtx = cc(connCtx, rw)
			if connCtx == nil {
				panic("ConnContext returned nil")
			}
		}
		tempDelay = 0
        // 构造处理用的conn，基于接入的Conn
		c := srv.newConn(rw)
        // 标记当前链接是新链接
		c.setState(c.rwc, StateNew, runHooks) // before Serve can return
        // 启动一个协程处理该conn
        // 1. 获取该链接Conn, 对应的的五元组中，远程地址信息, 本地地址信息
        // 2. 获取req信息，包含请求头，请求体等信息
        // 3. serverHandler{c.server}.ServeHTTP(w, w.req) 查询该req匹配到的多路复用器的路由地址
        // 4. 如果匹配到路由了，就去调用路由中对应的方法，处理该次链接。对应的就是一条回调函数，回调到路由注册的地方，例如上文第一个例子的`sayHello`函数内。
        // 5. 回调没有问题的话，标记当前请求被成功处理了，设置链接状态为StateIdle, 再做一些收尾工作，至此这个链接被处理完成。
		go c.serve(connCtx)
	}
```

## 定制化服务端
### 定制化多路复用器
```golang
// 构造一个新的路由匹配规则
mux := http.NewServeMux()
// 写入一条规则
mux.Handle("/", &helloHandler{})
```

### 定制化服务配置
```golang
		server := &http.Server{
		Addr:                         ":8080",    // 指定服务监听在哪个端口上或地址上。"127.0.0.1:8080"或者":8080"
		Handler:                      nil,        // 指定多路复用器，实现http.Handel接口的对象。实现路由匹配，不指定的话默认使用DefaultServeMux
		TLSConfig:                    nil,        // 如果你需要在 HTTPS 上运行服务器，可以设置该字段来配置 TLS。可以使用 tls.Config 类型的对象进行设置。
		ReadTimeout:                  0,          // 允许读取请求体的最大时间
		ReadHeaderTimeout:            0,		  // 允许读取请求头的最大时间
		WriteTimeout:                 0,          // 允许写入响应的最大时间
		IdleTimeout:                  0,          // 表示空闲连接的最大时间。如果客户端连接在一段时间内没有活动，超过该时间将会被关闭。这有助于释放闲置的资源。
		MaxHeaderBytes:               0,          // 用于限制请求头的大小。如果请求头太大，服务器会返回一个 400 错误。
		ConnState:                    nil,        // 一个可选的回调函数，用于监听连接状态的改变，比如新连接的建立、连接的关闭等。你可以在这里进行日志记录或其他处理。
		ErrorLog:                     nil,        // 用于记录服务器错误的日志记录器。可以是 log.Logger 对象，用于记录服务器运行时的错误信息。
		BaseContext:                  nil,        // 它指定返回该服务器使用的上下文的函数。
		ConnContext:                  nil,        // 它修改服务器接受的每个新连接的基本上下文。
	}
```

一般我们生产上只需要选择配置相关参数，满足自身业务即可。例如：
```golang
    server := &http.Server{
        Addr:           ":8080",
        Handler:        handler,
        ReadTimeout:    10 * time.Second,
        WriteTimeout:   10 * time.Second,
        IdleTimeout:    30 * time.Second,
        MaxHeaderBytes: 1 << 20, // 1 MB
        ErrorLog:       nil,     // 使用默认日志记录器
    }
```

完整的启动`net/http`包中Server的示例：
```golang
func main() {
	mux := http.NewServeMux()
	mux.Handle("/hello", &helloHandler{})

	server := &http.Server{
		Addr:           ":8080",
		Handler:        mux,
		ReadTimeout:    10 * time.Second,
		WriteTimeout:   10 * time.Second,
		IdleTimeout:    30 * time.Second,
		MaxHeaderBytes: 1 << 20, // 1 MB
		ErrorLog:       nil,     // 使用默认日志记录器
	}

	// 创建系统信号接收器
	done := make(chan os.Signal)
	// os.Interrupt 和 syscall.SIGINT 都表示 Ctrl+C 中断信号.
	// syscall.SIGTERM 则表示程序终止请求。kill -15
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-done

		if err := server.Shutdown(context.Background()); err != nil {
			log.Fatal("Shutdown server:", err)
		}
	}()

	log.Println("Starting HTTP server...")
	err := server.ListenAndServe()
	if err != nil {
		if err == http.ErrServerClosed {
			log.Print("Server closed under request")
		} else {
			log.Fatal("Server closed unexpected")
		}
	}
}

type helloHandler struct{}

func (*helloHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hasFirst := r.URL.Query().Has("first")
	first := r.URL.Query().Get("first")
	hasSecond := r.URL.Query().Has("second")
	second := r.URL.Query().Get("second")

	contentType := r.Header.Get("Content-Type")

	body, err := io.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	fmt.Println(hasFirst, first, hasSecond, second, contentType, string(body))

    w.Header().Set("Hello", "World")
	fmt.Fprintf(w, string(body))
}

// Output:
// 2023/07/27 11:47:11 Starting HTTP server...
// true 1 true 2 application/json {"name": "zhangsan","age": 18}
// 2023/07/27 11:47:24 Server closed under request // 优雅退出
// Exiting.
```

```shell
➜  ~ curl -XPOST -H "Content-type: application/json" -d '{"name": "zhangsan","age": 18}' '127.0.0.1:8080/hello?second=2&first=1' -v
Note: Unnecessary use of -X or --request, POST is already inferred.
*   Trying 127.0.0.1:8080...
* Connected to 127.0.0.1 (127.0.0.1) port 8080 (#0)
> POST /hello?second=2&first=1 HTTP/1.1
> Host: 127.0.0.1:8080
> User-Agent: curl/7.79.1
> Accept: */*
> Content-type: application/json
> Content-Length: 30
> 
* Mark bundle as not supporting multiuse
< HTTP/1.1 200 OK
< Hello: World
< Date: Thu, 27 Jul 2023 05:59:12 GMT
< Content-Length: 30
< Content-Type: text/plain; charset=utf-8
< 
* Connection #0 to host 127.0.0.1 left intact
{"name": "zhangsan","age": 18}
```

关于操作系统终端信号：
- os.Interrupt 和 syscall.SIGINT是`Ctrl+C`信号
- `Ctrl+Z`是syscall.SIGTSTP信号，会挂起当前进程（注意：终止服务使用Ctrl+Z）时，终端上退出了，但是可能持有的句柄没有释放，现象是端口继续被占用。
- syscall.SIGTERM 信号, 等价于`kill -15`，可以优雅退出

## 总结
`net/http`包下的`Server`是非常清晰易懂的，里面使用到了一些网络知识，但是被很好的封装了起来，屏蔽了不少复杂的技术细节，该包是TCP通信更高级别的抽象。想要启动一个HTTP服务，如果比较简单的接口和业务，大可直接选择基于原生的`net/http`包启动，凸现golang的轻量优势。了解到很多开源框架，都是基于这里的概念做的一系列封装，当业务复杂的时候，选择这些优秀的开源框架，是比较好的。希望以上的梳理，对于看到这里的小伙伴会有所帮助。