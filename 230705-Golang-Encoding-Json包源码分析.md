`encoding/json`包是`golang`标准库中编解码`json`相关的包, `RFC-7159`。`go`官方[blog](https://golang.org/doc/articles/json_and_go.html)也有相关的详细介绍。

## 字符编码问题
首先是`encoding`编码相关的知识大概说一下。字符编码问题，在计算机发展的初期，是比较混乱的，由于计算机最早流行在美国，美国人对字符编码，考虑到了大小写英文字母和符号，也就是`ASCII`编码。后面其他国家也需要把自己的文字编码到计算机中，于是出现了中国的`GB2312`编码，日本的`Shift_JIS`编码等。这种各自为营的方式在程序需要混用语言的时候就容易出现乱码。

于是大一统的`Unicode`字符集就统一了这一乱象。在`Unicode`编码中，一般字符分配2个字节，特殊字符，需要用到4个字节。现在操作系统和编程语言都天然支持Unicode编码。但是也存在弊端，当程序大部分使用的都是英文时，实际上每个英文字符用`ASCII`编码只需要一个字节即可，直接使用`Unicode`会造成存储浪费。

由于上面的问题，`Unocode`后面发展成`可变长编码`的`UTF-8`编码。`UTF-8`编码会把`Unicode`不同的字符选择存储为`1~6`个字节。
- 如果`Unicode`字符是一个英文字母等`Ascii`编码的字符, `UTF-8`会使用一个字节存储。
- 如果`Unicode`字符是一个中文汉子, `UTF-8`一般会使用三个字节存储, 生僻字有可能会使用4-6个字节存储。

`Go`语言里的字符串的内部实现使用`UTF-8`编码. 通过`rune`类型, 可以方便地对每个UTF-8字符进行访问。
- 当只需要处理`ASCII`编码的字符时，可以直接按顺序从字符串底层字节数组中取元素即可。
- 当要获取`UTF-8`编码的字符，`Go`中使用`rune`类型类接收。我们直接使用`range`遍历一个字符串，默认返回的就是`rune`字符。使用下标遍历字符串时，默认返回的是`byte`

## Json编码（序列化）
`json`编码也被称为`json`序列化, 在`golang`标准库中对`json`序列化是比较清晰的，主要就是两个公共方法，`Marshal`和`MarshalIndent`。其他私有方法，都是内部调用，为公共方法服务的。感兴趣可以进入标准库研究下私有方法，如何为`Marshal`和`MarshalIndent`提供支持。需要提一下，`Go`中，`json`编码，只会作用到`Go`结构体的public可导出字段（即首字母大写的字段）。

```golang
// 传入一个golang结构体，对该结构体进行json序列化，返回序列化后的字节数组，以及是否有报错（可以理解为返回压缩json）
func Marshal(v any) ([]byte, error)

// 传入一个golang结构体，对该结构体进行json序列化并添加缩进，增加json的可读性和美观。（可以理解为返回美化json）
// 返回序列化后的字节数组，以及是否有报错
// prefix表示前缀字符串: 如果要在生成的JSON中添加前缀（每行记录的开始，一般传递空白符），可以在这里指定。如果不需要前缀，可以传递一个空字符串。
// indent表示缩进字符串: 通常使用空格或制表符进行缩进，也可以传递一个空字符串来表示不使用缩进。
func MarshalIndent(v any, prefix, indent string) ([]byte, error)
```

### Json编码规则

`json`的编码大致需满足一下四条规则（摘抄自官方文档）：
- JSON objects only support strings as keys; to encode a Go map type it must be of the form map[string]T (where T is any Go type supported by the json package).
- Channel, complex, and function types cannot be encoded.
- Cyclic data structures are not supported; they will cause Marshal to go into an infinite loop.
- Pointers will be encoded as the values they point to (or ’null’ if the pointer is nil).

结构体匿名字段的编码，分两种情况。
- 如果匿名结构体中存在和主要结构体同名的字段（`json`的`tag`名也相同），则忽略匿名结构体中的该字段。
- 如果匿名结构体中存在和主要结构体同名的字段（`json`的`tag`名不相同），则主要结构体，和匿名结构体中的字段都参与编码。

### Json编码自定义选项
1. 可以通过指定`tag`来覆写`golang`结构体中默认字段名。
2. 可以通过`omitempty`控制字段为默认零值时，忽略该字段的编码。
3. 可以通过`-`的`tag`标识，直接控制某个字段不参与`json`编码。
4. 可以通过`string`的`tag`标识，控制`字符串`, `number`类型的值，编码为`string`类型。

### Marshal编码案例
```golang
func main() {
    stu := Student{
        NameA:      "欧阳",
        NameB:      "浩南",
        nameC:      "南哥",
        Age:        31,
        Weight:     77.3,
        ParentName: "南哥父亲",
        School: School{
            NameA: "欧阳学校",
            NameB: "浩南学校",
        },
    }

    b1, err := json.Marshal(stu)
    if err != nil {
        panic(err)
    }

    // 增加't'前缀
    // 通过制表符缩进，制表符一般是4个空格
    b2, err := json.MarshalIndent(stu, "", "\t")
    if err != nil {
        panic(err)
    }
    fmt.Println(string(b1))
    fmt.Println("=============")
    fmt.Println(string(b2))

    // Output:
    // {"name_a":"欧阳","name_b":"浩南","age":31,"weight":"77.3","name_sch_b":"浩南学校"}
    // =============
    // {
    //     "name_a": "欧阳",
    //     "name_b": "浩南",
    //     "age": 31,
    //     "weight": "77.3",
    //     "name_sch_b": "浩南学校"
    // }

}

type Student struct {
    NameA string `json:"name_a"`
    NameB string `json:"name_b"`
    // 非public字段，不参与序列化
    nameC string `json:"name_c"`
    Age   int64  `json:"age"`
    // 不存在，则不序列化该字段
    Height float64 `json:"height,omitempty"`
    // 把weight序列化为string类型
    Weight float64 `json:"weight,string"`
    // json tag为-，不参与json序列化
    ParentName string `json:"-"`
    // 嵌套匿名结构
    School
}

type School struct {
    // 与Stu的name_a相同，显示Stu的name_a
    NameA string `json:"name_a"`
    // 与Stu的name_b不相同，都显示
    NameB string `json:"name_sch_b"`
}
```


## Json解码（反序列化）

### 静态解码（固定结构接收）
`json`解码也被称为`json`反序列化。即通过字节流，重新构建`Golang`结构体对象。需要注意的是，解码一个`json`字节流，需要提前定义一个可接收该`json`字节流的结构体。另外一点是指定接收字节流反序列化的对象，必须使用指针地址接收。在经过解码函数的作用后，未产生报错，`json`字节流会被存储到该对象中。

当`json`字节流与需要反序列化的结构字段，无法一一匹配时。则只序列化匹配上的字段。当我们希望从一个大的`json`体中只序列化某些字段时，定义正确接受的结构体即可部分接收。且目标结构的非`public`字段，不参与反序列化。

```golang
func main() {
    b := []byte(`{"Name":"ZhangSan","Age":18,"Schools":["Qinghua","Beida"]}`)
    var message Message

    err := json.Unmarshal(b, &message)
    if err != nil {
        panic(err)
    }

    fmt.Printf("message is: %+v", message)
    
    // Output:
    // message is: {Name:ZhangSan Schools:[Qinghua Beida]}
}

// 选择接受，不接受json中的Age
type Message struct {
    Name    string
    Schools []string
}
```

### 动态解码（通用结构接收）
在`golang`中，我们可以使用`map[string]interface{}`或者`[]interface{}`结构组合，来接收任意的`json`反序列化的内容。其中，`interface{}`后续的类型判断时：
- bool for JSON booleans（bool值，用来接受json的booleans变量）
- float64 for JSON numbers（float64，用来接收json的数值类型）
- string for JSON strings（string, 用来接收json的字符串类型）
- nil for JSON null（nil用来接收json的null类型）

```golang
func main() {
    b := []byte(`{"Name":"ZhangSan","Age":18,"Schools":["Qinghua","Beida"]}`)
    var message interface{}

    err := json.Unmarshal(b, &message)
    if err != nil {
        panic(err)
    }

    // 第一层，首先断言为map[string]interface{}
    m := message.(map[string]interface{})
    for k, v := range m {
        switch vv := v.(type) {
        case string:
            fmt.Println(k, "is string", vv)
        case float64:
            fmt.Println(k, "is float64", vv)
        case []interface{}:
            fmt.Println(k, "is an array:")
            for i, u := range vv {
                fmt.Println(i, u)
            }
        default:
            fmt.Println(k, "is of a type I don't know how to handle")
        }
    }

    // Output:
    // Name is string ZhangSan
    // Age is float64 18
    // Schools is an array:
    // 0 Qinghua
    // 1 Beida
}
```

## Json编解码流
在Go语言的标准库中，`json.Decoder`和`json.Encoder`是用于JSON数据的解码和编码的结构。它们提供了一种方便的方式来处理JSON数据流，`json.Decoder`从输出流中读取`json`数据。`json.Encoder`将`json`数据写入到输入流中。

### Decoder
```golang
// 初始化一个json解码器Decode，从输入流读取
func NewDecoder(r io.Reader) *Decoder

// 配置Decoder编码使用Numer类型(string)处理数字，而不是float64处理。默认是float64处理数字
func (dec *Decoder) UseNumber()

// 配置未知字段报错。如果解码的过程中，接受的结构体中不存在json中的某个字段的key匹配的字段，则报错。
// 即不支持部分字段的接收，默认是支持的。
func (dec *Decoder) DisallowUnknownFields()

// 解码输入流中的一个JSON值并将其存储在v指向的结构中。v可以是任何Go语言的数据结构，通过将其指针传递给Decode方法，可以将JSON数据解码为该结构。
func (dec *Decoder) Decode(v any) error

// 返回用于解码的底层缓冲区的io.Reader接口。此方法可用于检查缓冲区中剩余的未处理数据。
func (dec *Decoder) Buffered() io.Reader

// 报告输入流中是否还有更多的JSON值可供读取。
func (dec *Decoder) More() bool

// 返回输入流中的下一个JSON令牌（Token）。令牌可以是基本类型（如数字、字符串、布尔值等）或分隔符（如{、}、[、]等）。
// 在输入流的末尾，Token返回nil, io.EOF。Token保证分隔符是正确的嵌套和匹配的
func (dec *Decoder) Token() (Token, error)

// InputOffset返回当前解码器位置的输入流字节偏移量。偏移量给出了最近返回的令牌的末尾位置和下一个符号的开头。
// 即当前读取到了输入流的哪个位置。
func (dec *Decoder) InputOffset() int64
```

### Encoder
```golang
// 初始化一个json编码器，写入到输出流
func NewEncoder(w io.Writer) *Encoder

// 将golang结构v，进行json编码，写入到输出流，每写完一个v，紧跟一个换行符。
func (enc *Encoder) Encode(v any) error

// 设置json编码器的默认美化效果，指定了美化json的格式，再Encode时，就写入美化后的json，而不是压缩json
func (enc *Encoder) SetIndent(prefix, indent string)

// 设置json编码器编码的json，要是html页面安全的json。对一些html标签的字符，做Unicode转义
func (enc *Encoder) SetEscapeHTML(on bool)
```

### RowMessage
`RowMessage`是json包内定的一个新的结构，用来接收`json`编码后的值。本质是通过`[]byte`得来。该对象包含两个方法，`MarshalJSON`和`UnmarshalJSON`。

```golang
type RawMessage []byte

// 把RowMessage转化为json编码的[]byte
// m -> []byte
func (m RawMessage) MarshalJSON() ([]byte, error)

// 通过json编码的[]byte，转化为RowMessage类型的m
// []byte -> m
func (m *RawMessage) UnmarshalJSON(data []byte) error
```

### Delim
一个`json`流中，存在以下集中元素.
1. `json`每个`key`，是一个字符串类型。
2. `key`对应的`value`，是基本类型`booleans`, `numbers`, `string`, `null`
3. `json`的分隔符，其中含有`[`, `]`, `{`, `}`。在这里这四个分割服，被`golang`标准库的`json`包定义为`Delim`类型。

```golang
type Delim rune

// 返回Delim对应符号的字符串表示。
func (d Delim) String() string
```

### 案例
```golang
func main() {
    // 模拟一个包含多个JSON对象的数据流。这里固定一个内容, 正常业务流程，这里应该是json编码器Encoder的写入流内的数据。
    data := `{"name": "John", "age": 30}
    {"name": "Alice", "age": 25}
    {"name": "Bob", "age": 35}`

    decoder := json.NewDecoder(strings.NewReader(data))

    for {
        var obj map[string]interface{}
        err := decoder.Decode(&obj)
        if err != nil {
            // 如果遇到错误或流结束，则跳出循环
            break
        }

        fmt.Println(obj)

        // 判断是否还有更多的JSON值可供读取
        if !decoder.More() {
            break
        }
    }
    
    // Output:
    // map[age:30 name:John]
    // map[age:25 name:Alice]
    // map[age:35 name:Bob]
}
```

## 其他扩充功能

### Json转换
```golang
// 把原始json字节数组src，进行字符转化，转化为HTML页面显示安全的json。
// 一般的，<，>，&，U+2028和U+2029等html敏感的字符更改为\u003c， \u003e， \u0026， \u2028， \u2029。
func HTMLEscape(dst *bytes.Buffer, src []byte)

// 可以理解为，json压缩。美化后的json，存在很多空白符，该方法可以去除。
func Compact(dst *bytes.Buffer, src []byte) error

// json美化，类似于MarshalIndent。与MarshalIndent不同的是，这里是直接通过json序列化后的结果进行美化，MarshalIndent是在结构体序列化的过程中美化。
func Indent(dst *bytes.Buffer, src []byte, prefix, indent string) error
```

### 校验Json
```golang
// 给定一个字节流data，判断是否是满足json编码的字节流
func Valid(data []byte) bool
```

```golang
func main() {
    src := []byte(`
    {
     "name_a": "欧阳",
     "name_b": "浩南",
     "age": 31,
     "weight": "77.3",
     "name_sch_b": "浩南学校"
    }`,
    )

    fmt.Println(string(src))

    b := make([]byte, 0)
    dst := bytes.NewBuffer(b)
    err := json.Compact(dst, src)
    if err != nil {
        panic(err)
    }
    fmt.Println("=========")

    fmt.Println(string(dst.Bytes()))

    // Output:
    // {
    //         "name_a": "欧阳",
    //         "name_b": "浩南",
    //         "age": 31,
    //         "weight": "77.3",
    //         "name_sch_b": "浩南学校"
    // }
    // =========
    // {"name_a":"欧阳","name_b":"浩南","age":31,"weight":"77.3","name_sch_b":"浩南学校"}
}
```

## 总结
`golang`标准库中`json`编码解码的功能，是非常清晰的，提供给上层调用的仅仅只有三个方法。`Marshal`, `MarshalIndent`, `Unmarshal`。后面又熟悉了标准库提供的`json`流操作，提供了更复杂`json`处理的方法。以上，基本囊括了日常使用中，对`json`的各种处理。