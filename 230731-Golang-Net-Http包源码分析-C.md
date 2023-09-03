`golang`标准库`http`客户端主要包含以下结构客户端`Client`，请求结构`Request`。客户端`Client`发送请求的时候需要携带请求结构`Request`，后`client.Do(req)`发给服务端。在标准库中，可以看到有一些请求类似直接`Client.Get(url)`请求，其内部也是构造了一个默认`Request`后，请求服务端的。

## Client客户端
构造一个客户端，可以选配四个字段，默认的客户端是`Client{}`，表示不配置任何`option`的客户端。

```golang
type Client struct {
    Transport RoundTripper

    CheckRedirect func(req *Request, via []*Request) error

    Jar CookieJar

    Timeout time.Duration
}
```

1. `Transport`：这是一个`http.RoundTripper`接口类型的字段，用于配置`HTTP`客户端的传输行为。它决定了客户端发送请求和接收响应的方式。默认情况下，如果不指定该字段，会使用`http.DefaultTransport`，它在大多数情况下都能够满足需求。如果需要自定义传输行为，可以通过创建自定义的`http.RoundTripper`实现来设置该字段。
2. `CheckRedirect`：这是一个函数字段，用于指定在进行`HTTP`重定向时的行为。默认情况下，`http.Client`会遵循所有的`HTTP`重定向，直到达到最终的目标 URL 或达到最大重定向次数（默认是 10 次）。通过设置该字段，开发者可以自定义重定向的行为。例如，可以指定在重定向时是否继续请求，或者在特定情况下中止重定向。
3. `Jar`：这是一个`http.CookieJar`接口类型的字段，用于管理 HTTP 客户端的`Cookie`。`Cookie`是一种在客户端和服务器之间传递信息的机制，用于维持会话状态。默认情况下，`http.Client`使用`http.DefaultCookieJar`来管理`Cookie`。通过设置该字段，开发者可以使用自定义的`http.CookieJar`实现来控制`Cookie`的处理。
4. `Timeout`：这是一个`time.Duration`类型的字段，用于指定客户端在发出请求后等待服务器响应的最大时间。如果服务器在指定的时间内没有响应，客户端将放弃该请求并返回超时错误。默认情况下，`http.Client`的 `Timeout`字段为零，表示不设置超时。如果需要对请求进行超时控制，可以设置该字段为一个合适的时间间隔。

通过上文的解释，一般来说，日常开发中，我们只需要酌情考虑配置`RoundTripper`就行。大多数情况下可以完全默认。扩展`RoundTripper`时，需要考虑并发安全问题，即多个请求都调用`RoundTripper`的`RoundTrip`方法时，不会存在并发问题。

分析源码可以看到，在`client.Do(req)`的时候，会最终执行到`resp, didTimeout, err = send(req, c.transport(), deadline)`,其中`c.transport()`就是检查我们是否配置了`RoundTripper`，如果没有配置，就使用默认的`DefaultTransport`对应的实现。

```golang
func (c *Client) transport() RoundTripper {
    if c.Transport != nil {
        return c.Transport
    }
    // 如果没配置`RoundTripper`，使用默认的实现
    return DefaultTransport
}
```

首先我们从`DefaultTransport`来分析，到底发生了什么。从上面的分析中，我们知道，无论我们有没有配置`RoundTripper`, 在`c.transport()`总会得到一个实现，而`DefaultTransport`实现，就是`http.Transport`结构, 该结构及其复杂，是http请求得以成功的保证，内部封装了一些网络请求细节, 很多http请求网络问题，大可能在这里可以找到答案。

```golang
// DefaultTransport is the default implementation of Transport and is
// used by DefaultClient. It establishes network connections as needed
// and caches them for reuse by subsequent calls. It uses HTTP proxies
// as directed by the environment variables HTTP_PROXY, HTTPS_PROXY
// and NO_PROXY (or the lowercase versions thereof).
var DefaultTransport RoundTripper = &Transport{
    Proxy: ProxyFromEnvironment,
    DialContext: defaultTransportDialContext(&net.Dialer{
        Timeout:   30 * time.Second,
        KeepAlive: 30 * time.Second,
    }),
    ForceAttemptHTTP2:     true,
    MaxIdleConns:          100,
    IdleConnTimeout:       90 * time.Second,
    TLSHandshakeTimeout:   10 * time.Second,
    ExpectContinueTimeout: 1 * time.Second,
}
```

这个结构对应的`RoundTrip`方法，是其内部的`roundTrip方法`, 该方法是`http.Client`和`http.Request`联合，请求服务端，得到响应的根基，涉及到获取链接，请求服务端，得到`resp`等。

```golang
// RoundTrip implements the RoundTripper interface.
//
// For higher-level HTTP client support (such as handling of cookies
// and redirects), see Get, Post, and the Client type.
//
// Like the RoundTripper interface, the error types returned
// by RoundTrip are unspecified.
func (t *Transport) RoundTrip(req *Request) (*Response, error) {
    // 这个方法，是golang http请求服务端的，核心方法。
    return t.roundTrip(req)
}

// roundTrip implements a RoundTripper over HTTP.
func (t *Transport) roundTrip(req *Request) (*Response, error) {
    // ... 省略：一些http请求参数非法的校验代码 ...

    for {
        // 检查是否有取消请求
        select {
        case <-ctx.Done():
            req.closeBody()
            return nil, ctx.Err()
        default:
        }

        // treq gets modified by roundTrip, so we need to recreate for each retry.
        treq := &transportRequest{Request: req, trace: trace, cancelKey: cancelKey}
        // 通过request获取connectMethod
        cm, err := t.connectMethodForRequest(treq)
        if err != nil {
            req.closeBody()
            return nil, err
        }

        // Get the cached or newly-created connection to either the
        // host (for http or https), the http proxy, or the http proxy
        // pre-CONNECTed to https server. In any case, we'll be ready
        // to send it requests.
        // 获取http/https的网络链接
        pconn, err := t.getConn(treq, cm)
        if err != nil {
            t.setReqCanceler(cancelKey, nil)
            req.closeBody()
            return nil, err
        }

        var resp *Response
        if pconn.alt != nil {
            // HTTP/2 path.
            // http2的路径，一般我们http请求是1.1，走下一个分支
            t.setReqCanceler(cancelKey, nil) // not cancelable with CancelRequest
            resp, err = pconn.alt.RoundTrip(req)
        } else {
            // 一般来说，我们会进入这个分支。通过req和网络链接pconn， 获取网络资源resp
            resp, err = pconn.roundTrip(treq)
        }
        if err == nil {
            // 让正确响应体，把请求信息待会给客户端。
            resp.Request = origReq
            return resp, nil
        }

        // ... 省略：失败场景的一些善后工作，重试之前的一些准备工作...
    }
}
```

### 如何扩展客户端的RoundTripper
通过上面的分析，我们得知，`RoundTripper`实际上提供了一个请求到响应的一个路径，也发现默认的`http.Transport`内部的网络细节。扩展`http.Client`的`RoundTripper`主要思路，还是使用其默认的`RoundTripper`,复用`golang`网络库的完备性，我们仅仅需要一个时机，像切面一样，在默认的请求`RoundTrip`方法完成后，我们后置做一些`收尾`, `日志`, `打点`等统一动作。

建议看到这里的小伙伴，可以看一些开源社区，一些对于`http.RoundTripper`接口的扩展案例，这里简单分析一下`opentelemetry`对与`net/http`的扩展[案例](https://github.com/open-telemetry/opentelemetry-go-contrib/blob/384acfd0d4e2fb4a837797032cda719e585db5dd/instrumentation/net/http/otelhttp/transport.go#L32C2-L32C2)。

```golang
// Transport implements the http.RoundTripper interface and wraps
// outbound HTTP(S) requests with a span.
type Transport struct {
    rt http.RoundTripper

    tracer            trace.Tracer
    propagators       propagation.TextMapPropagator
    spanStartOptions  []trace.SpanStartOption
    filters           []Filter
    spanNameFormatter func(string, *http.Request) string
    clientTrace       func(context.Context) *httptrace.ClientTrace
}

// NewTransport wraps the provided http.RoundTripper with one that
// starts a span and injects the span context into the outbound request headers.
//
// If the provided http.RoundTripper is nil, http.DefaultTransport will be used
// as the base http.RoundTripper.
func NewTransport(base http.RoundTripper, opts ...Option) *Transport {
    if base == nil {
        // 这里建议不要自己实现base http.RoundTripper, 传递nil就好
        base = http.DefaultTransport
    }

    t := Transport{
        rt: base,
    }

    defaultOpts := []Option{
        WithSpanOptions(trace.WithSpanKind(trace.SpanKindClient)),
        WithSpanNameFormatter(defaultTransportFormatter),
    }

    c := newConfig(append(defaultOpts, opts...)...)
    t.applyConfig(c)

    return &t
}

// RoundTrip creates a Span and propagates its context via the provided request's headers
// before handing the request to the configured base RoundTripper. The created span will
// end when the response body is closed or when a read from the body returns io.EOF.
func (t *Transport) RoundTrip(r *http.Request) (*http.Response, error) {
    for _, f := range t.filters {
        if !f(r) {
            // Simply pass through to the base RoundTripper if a filter rejects the request
            // rt就是base，一般也就是http.DefaultTransport
            return t.rt.RoundTrip(r)
        }
    }

    opts := append([]trace.SpanOption{}, t.spanStartOptions...) // start with the configured options

    ctx, span := t.tracer.Start(r.Context(), t.spanNameFormatter("", r), opts...)

    r = r.WithContext(ctx)
    span.SetAttributes(semconv.HTTPClientAttributesFromHTTPRequest(r)...)
    t.propagators.Inject(ctx, propagation.HeaderCarrier(r.Header))

    // rt就是base，一般也就是http.DefaultTransport, 复用了golang http网络请求的能力
    res, err := t.rt.RoundTrip(r)
    if err != nil {
        span.RecordError(err)
        span.End()
        return res, err
    }

    // 在调用base的请求后，拿到resp, 这里可以提供日志，metric，trace等指标收集工作。
    // 上层使用http.Client的时候，传入我们定义的这个Transport
    span.SetAttributes(semconv.HTTPAttributesFromHTTPStatusCode(res.StatusCode)...)
    span.SetStatus(semconv.SpanStatusFromHTTPStatusCode(res.StatusCode))
    res.Body = &wrappedBody{ctx: ctx, span: span, body: res.Body}

    return res, err
}
```

`Transport`在其方法`RoundTrip`中，在调用base的请求后，拿到resp后, 后续可以提供日志，metric，trace等指标收集工作。上层使用http.Client的时候，传入我们定义的这个`Transport`，那么我们就完成了`net/http/Client`的`RoundTripper`扩展。

## Request请求结构
请求结构`Request`，是`http`请求中不可或缺的内容，任何请求或显式或隐式的或构造该结构后，发送给服务端。例如`http.Get(url)`，看起来没有传入`request`，本质是内部进行了封装。

另外，我们知道如果我们使用默认的`Client`，即`Client{}.Do(req)`发起请求时，本质上是使用的`Client`结构默认的`DefaultTransport.RoundTrip(req)`来实现。所以，对于客户端发起请求，最重要的结构，实质上是`http.Request`和`http.DefaultTransport`。

```golang
type Request struct {
    Method string                           // 请求的方法，支持：GET, POST, PUT等等
    URL *url.URL                            // 该字段是一个指向 url.URL 类型的指针，表示请求的 URL 信息，包括主机、路径、查询参数等。
    Proto      string                       // 该字段表示请求使用的协议版本，通常为 "HTTP/1.1"(ProtoMajor=1, ProtoMinor=1) 或 "HTTP/2"。
    ProtoMajor int                          // 该字段表示请求使用的主要协议版本号，通常为 1。
    ProtoMinor int                          // 该字段表示请求使用的次要协议版本号，通常为0或者1。
    Header Header                           // 该字段是一个 http.Header 类型的映射，表示请求的头部信息，包括请求头字段和对应的值。
    Body io.ReadCloser                      // 该字段是一个 io.ReadCloser 接口类型，表示请求的主体 (body) 数据。在 POST、PUT 等请求中，可以通过该字段来读取请求体中的数据, 读取后需要手动关闭。
    GetBody func() (io.ReadCloser, error)   // 类似Body，不同点在于，GetBody获取请求新的io.ReadCloser，从而达到可以重复读取Body的目的。每次读取仍需关闭获取到的io.ReadCloser
    ContentLength int64                     // 该字段表示允许请求主体Body的长度，以字节为单位。
    TransferEncoding []string               // 该字段表示请求的传输编码方式，如 "chunked"。
    Close bool                              // Close方法是用于关闭请求主体的方法。通常是在POST、PUT等请求中包含表单数据或JSON数据。通过Close，我们可以显式地关闭请求主体的数据流。
    Host string                             // 该字段表示请求的主机名，如果请求中没有指定主机名，则会使用请求的 URL 中的主机名。
    Form url.Values                         //该字段是一个url.Values类型的映射，表示解析后的表单数据。在POST请求中，如果请求头部的Content-Type是 application/x-www-form-urlencoded，则可以通过该字段访问表单数据。
    PostForm url.Values                     // 该字段是一个 url.Values 类型的映射，表示解析后的POST表单数据。与Form不同的是，PostForm只在POST请求中有效，而且仅当请求头部的Content-Type是 application/x-www-form-urlencoded时有效。
    MultipartForm *multipart.Form           // 该字段是一个 *multipart.Form 类型的指针，表示解析后的多部分表单数据。在 POST 请求中，如果请求头部的 Content-Type 是 multipart/form-data，则可以通过该字段访问多部分表单数据。
    
    
    Trailer Header                          // Trailer字段是一个http.Header类型的映射，表示请求Trailer头部（HTTP 规范中称为尾部头）。Trailer头部是一种允许在HTTP消息主体之后发送额外头部的方式，但在发送主体之前，无法确定Trailer头部的内容。通常Trailer头部在Transfer-Encoding: chunked 的情况下使用。
    RemoteAddr string                       // RemoteAddr字段是一个字符串，表示发起请求的客户端的地址。它通常是客户端的IP地址加上端口号。
    RequestURI string                       // RequestURI 字段是一个字符串，表示请求的完整 URL。它包含了主机名、路径以及查询参数等信息。
    TLS *tls.ConnectionState                // TLS 字段是一个 *tls.ConnectionState 类型的指针，用于表示请求的 TLS 连接状态。如果请求是通过 HTTPS 发送的，那么 TLS 字段将包含关于 TLS 握手和证书等信息。否则，TLS 字段将为 nil。
    Cancel <-chan struct{}                  // Cancel是用于取消HTTP请求的方法。用于在 HTTP 请求处理期间检测是否需要取消请求。在某些情况下，如果客户端关闭了连接或发生了其他错误，我们可能需要提前终止请求处理，以避免浪费服务器资源。Cancel 字段通常与 context.Context 配合使用。
    Response *Response                      // Response 字段是一个 *http.Response 类型的指针，表示正在处理的HTTP响应。这个字段在HTTP处理器中是空的，它是为了方便在某些情况下在请求处理期间进行错误处理时使用的。
    ctx context.Context                     // ctx 字段是一个 context.Context 类型，用于传递上下文信息。它在请求处理的整个生命周期中都有效，并且可以用于传递请求相关的数据和控制请求的取消操作。通过在 http.Request 对象中设置 ctx 字段，我们可以在不同的处理器中共享上下文信息，并在需要时对请求进行取消。
}
```

简单说一下`Request`结构中的扩展类型`url.URL`, 这个结构在`net.http`平级的包中，是`统一资源定位符`的抽象。在统一资源定位符的概念中，主要包含一下几种成分组成：`服务器地址`, `路径`, `查询参数`, `片段`组成。前面三个比较好理解，`服务器地址`形如`http://127.0.0.1:80`, `路径`形如`/user/group_1`, `查询参数`形如`?name=abc`。而`片段`一般用在更细粒度的资源定位上，比如请求地址是一个文档字段，片段可以细化到文档的哪一行。比如：`https://example.com/page.html#section1`，这里`#section1`就是`page.html`资源更细粒度的定位。

在`golang`标准库中`net/url`中`URL`结构对`统一资源定位符`进行了抽象，提供了一系列方法，比如获取URL的各种成分信息。也提供了URL各个成分的编码解码（eg: `PathEscape`, `QueryEscape`， `PathUnescape`, `QueryUnescape`），方便在网络传输中不会出现问题，当我们的path或者url是不确定是否存在需要转码的字符的时候，最好转码处理一下。更多的细节，请进入源码包继续深入。

### 封装Request

关于Request的最佳实践大概有下面几条：
1. **强制携带context**: 通过在 http.Request 对象的 WithContext 方法中传递 context.Context，可以在请求处理过程中传递数据、控制请求的取消操作等。这样可以避免使用全局变量或在函数之间传递大量参数。
2. **强制处理请求的错误**: 在处理 HTTP 请求时，要正确地处理请求和响应的错误。例如，在解析请求参数时，要检查是否出现错误，并正确返回错误信息给客户端。
3. **使用正确的解析器，解析请求的数据**: 对于 POST、PUT 等带有请求主体的请求，可以使用合适的解析器来处理请求数据。Golang 的 http 包提供了 http.Request.ParseForm 和 http.Request.ParseMultipartForm 等方法来解析表单数据和多部分表单数据, 需要对应响应的请求头，参考上文。
4. **使用 http.Client 发送 HTTP 请求**：如果应用程序需要发送 HTTP 请求，可以使用 Golang 提供的 http.Client。http.Client 提供了方便的方法来发送 GET、POST 等请求，并支持超时、重试等功能。其实还是有`http.DefaultTransport`来保证的。

`Request`通过http请求，分析http请求各阶段耗时问题，使用ctx传递参数的例子，本质是利用了`net/http/httptrace`包的功能：

```golang
func main() {
    // 使用http.NewRequest或者http.NewRequestWithContext构建一个Request,区别是使用传入的ctx或者让基础库自己生成一个。
    req, err := http.NewRequest("GET", "http://127.0.0.1/hello", nil)
    if err != nil {
        panic(err)
    }

    // 获取到req后，可以酌情填充需要的参数，例如填充请求头
    req.Header.Set("Lang", "zh-CN")

    // start: 请求开始发送的时刻
    // dns: 开始进行dns解析的时刻
    // tlsHandshake: 开始解析tls的时刻
    // connect: 开始获取连接的时刻
    var start, dns, tlsHandshake, connect time.Time
    // dnsTime: dns解析阶段的耗时情况
    // tlsTime: tls解析截断的耗时情况
    // connTime: 连接获取截断的耗时情况
    // startToDataTime: 从开始发起请求，到获取到数据的耗时情况
    // totalTime: 格外扩展的，从开始发起请求，到响应体数据被正确解析截断的耗时情况
    var dnsTime, tlsTime, connTime, startToDataTime, totalTime time.Duration

    // 采集http请求过程中，各个阶段的耗时情况。构建httptrace.ClientTrace，利用req的context携带信息的特征，把变量地址携带进请求中，记录。
    // 更多trace信息参考文档：https://blog.golang.org/http-tracing
    req = req.WithContext(httptrace.WithClientTrace(req.Context(), &httptrace.ClientTrace{
        DNSStart: func(dsi httptrace.DNSStartInfo) { dns = time.Now() },
        DNSDone: func(ddi httptrace.DNSDoneInfo) {
            dnsTime = time.Since(dns)
        },
        TLSHandshakeStart: func() { tlsHandshake = time.Now() },
        TLSHandshakeDone: func(cs tls.ConnectionState, err error) {
            tlsTime = time.Since(tlsHandshake)
        },
        ConnectStart: func(network, addr string) { connect = time.Now() },
        ConnectDone: func(network, addr string, err error) {
            connTime = time.Since(connect)
        },
        GotFirstResponseByte: func() {
            startToDataTime = time.Since(start)
        },
    }))

    start = time.Now()
    resp, err := http.DefaultTransport.RoundTrip(req)
    if err != nil {
        panic(err)
    }
    defer resp.Body.Close()

    respBody, err := io.ReadAll(resp.Body)
    if err != nil {
        panic(err)
    }
    totalTime = time.Since(start)

    log.Printf("httpRespStatusCode=%d | httpRespBody=%s", resp.StatusCode, string(respBody))

    log.Printf("dnsTime=%d | tlsTime=%d | connTime=%d | startToDataTime=%d | totalTime=%d",
        dnsTime, tlsTime, connTime, startToDataTime, totalTime)
}
```

## 总结
`net/http`关于客户端的源码分析就到这里了，由于`net/http`乃至`net`包，都属于标准库中比较庞大的包，庞大意味着大并且全，所以开源社区对golang在网络编程领域的优势都是认可的。通过上面两篇，我们仅仅是从一个使用者的角度，深入进去，验证正确的服务启动流程，客户端正确请求到服务端流程的两条路径展开，其中还扩展了一些额外的但是很重要的结构，比如扩展标准库路由的不足引入的第三方包，如何正确使用Request，如何扩展`net/http`的`http.DefaultTransport`, 后面简单认识了一下`net/http/httptrace`包，以及`net/url`包，最后介绍了如何对请求做阶段性耗时的追踪等。最后的最后，golang网络包比较庞大，封装了大量的网络细节，感兴趣的，可以继续深入。