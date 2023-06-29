`strings`包主要提供一些函数，用来操作`UTF-8`编码的字符串。在Go中，字符串本质上是`[]byte`类型的只读切片即`字节切片`。所以我们通过转换`string`
为`[]byte`时，每个位置是字节，并不是字符（如果一个字符不是一个字节存储时，就发现端倪, 例如`⌘`符号会使用三个字节存储，分别为`e2`, `8c`, `98`。
这些字节是十六进制值2318的UTF-8编码）。所以字符串底层是字节数组，字节数组用来存储字符串的各个字符的字节表示，有些字符一个字节存不下就会用多个字节存储。
字符串可以包含任意的字节，但是字符串的字面量，总是可以看成（几乎总是）`UTF-8`。

每个字符，都对应一个`Code point码点`，Go语言引入了一个更短的标识，来代表码点，即`rune`。可以理解`[]rune`可以每个元素存储一个`码点`，即存储一个字符。
Go语言`rune`是一个`int32`表示，占用4个字节。

字符串可以包含任意的字节，但其包内部，对于适配`UTF-8`编码，反而是其最核心的部分。

```golang
// 演示字符串中字符如何存储的例子
func main() {
	const placeOfInterest = `⌘`
	fmt.Printf("plain string: %s \n", placeOfInterest)

    // 对字符转义为ascii码，和引号包裹。无法转义为ascii的字符使用\u编码表示(Unicode 码点)。`⌘`码点是`U+2318`, 被转义的最终结果是`"\u2318"`
	fmt.Printf("quoted string: %+q \n", placeOfInterest)

	fmt.Printf("hex bytes: ")
	for i := 0; i < len(placeOfInterest); i++ {
        // 取每个字节对应ascii码的十六进制表示
		fmt.Printf("%x ", placeOfInterest[i])
	}
	fmt.Printf("\n")
	
	// Output:
	// plain string: ⌘ 
	// quoted string: "\u2318" 
	// hex bytes: e2 8c 98
}
```

对字符串进行`for`循环，每次迭代时，会解码出来一个`utf-8`字符。每次循环的索引，是当前`utf-8`字符的起始位置（以字节为单位）。每次循环出来的字符，使用代码点（对于golang是rune返回Unicode值）来表示, 可以使用`%#U`来接收。

```golang
func main() {
	var Str string = "我爱中国天安门"

    // 每次遍历的index是当前runeValue字符的起始字节位置
    // 每个runeValue类型是4字节表示的rune类型，表示Unicode码值
	for index, runeValue := range Str {
		fmt.Printf("%#U starts at byte position %d\n", runeValue, index)
	}
}

// Output:
// U+6211 '我' starts at byte position 0
// U+7231 '爱' starts at byte position 3
// U+4E2D '中' starts at byte position 6
// U+56FD '国' starts at byte position 9
// U+5929 '天' starts at byte position 12
// U+5B89 '安' starts at byte position 15
// U+95E8 '门' starts at byte position 18
```

也可以使用golang标准库的utf-8编码包，进行字符的提取，更底层。

```golang
func main() {
	var Str string = "我爱中国天安门"

	for index, w := 0, 0; index < len(Str); index += w {
		// 提取一个UTF-8编码的字符。返回字符，及该字符占用底层字节数组的宽度
		runeValue, width := utf8.DecodeRuneInString(Str[index:])
		// index为该字符的起始位置
		fmt.Printf("%#U starts at byte position %d\n", runeValue, index)
		// 下一个字符的起始位置，等于index+w
		w = width
	}

	// Output:
	// U+6211 '我' starts at byte position 0
	// U+7231 '爱' starts at byte position 3
	// U+4E2D '中' starts at byte position 6
	// U+56FD '国' starts at byte position 9
	// U+5929 '天' starts at byte position 12
	// U+5B89 '安' starts at byte position 15
	// U+95E8 '门' starts at byte position 18
}
```

golang标准库中，`strings`包关于字符串处理有较多内容，该包下主要文件有:

- `builder.go`提供了一个`String builder`可以用来级联调用构造一个字符串。
- `clone.go`提供了字符串复制拷贝的方法。
- `compare.go`提供了字符串字典序比较的方法。
- `reader.go`提供了一个字符串读取器。可以像操作文件一样，操作字符串数据。
- `replacer.go`提供了一个字符串的替换器, 提供了简单灵活的字符串替换功能。它可以用于一次性替换多个子串。例如`replacer := strings.NewReplacer("abc", "def", "opq", "xyz")`表示构建一个替换器，把`abc`替换为`def`; 把`opq`替换为`xyz`。
- `search.go`提供了一些字符串查找的能力，主要用于标准库内部使用。
- `strings.go`提供了处理字符串的常用方法，是`strings`包核心。

## 字符串构造器strings.Builder
`strings.Builder`是Go语言标准库中的类型，用于高效地构建字符串。它提供了一种可变大小的缓冲区，可以按需添加和拼接字符串，最终生成一个完整的字符串结果。
使用`strings.Builder`可以更高效率的拼接字符串，避免频繁创建和直接拼接字符串带来的性能开销。

```golang
// Builder结构
type Builder struct {
	addr *Builder // of receiver, to detect copies by value
	buf  []byte
}

// 构建一个Builder
builder := strings.Builder{}

// 追加字符串到b缓冲区。
func (b *Builder) WriteString(s string) (int, error)

// 追加一个rune字符，到b缓冲区
func (b *Builder) WriteRune(r rune) (int, error)

// 追加一个byte字节，到b缓冲区。
func (b *Builder) WriteByte(c byte) error

// 追加一个字节数组p，到b缓冲区。
func (b *Builder) Write(p []byte) (int, error)

// 扩容b的buf容量，如果可用空间（ cap(buf) - len(buf) < n ), 扩容n个单位
func (b *Builder) Grow(n int)

// 重置b, 释放b持有的addr和buf空间。
func (b *Builder) Reset()

// 获取b, buf的cap
func (b *Builder) Cap() int

// 获取b，buf的len
func (b *Builder) Len() int

// 将缓冲区中的内容转换为最终的字符串结果。
func (b *Builder) String() string
```

## 字符串拷贝
```golang
func main() {
	var s1 = "abc"
	s2 := strings.Clone(s1)

	fmt.Println(s2)

	// Output:
	// abc
}

func Clone(s string) string {
	// 如果字符串长度为0，则返回一个新的空串
	if len(s) == 0 {
		return ""
	}
	// 根据传入的字符串的底层字节数组的长度，构建一个新的字节数组
	b := make([]byte, len(s))
	// 调用copy方法，拷贝字节数组
	copy(b, s)
	// 构建字节数组为字符串，传入字节数组首地址，和字节数组的长度。
	return unsafe.String(&b[0], len(b))
}
```

## 字符串字典序比较
```golang
func main() {
	var s1 = "abc"
	var s2 = "def"
	// s1的字典序小于s2的字典序
	cmp := strings.Compare(s1, s2)

	fmt.Println(cmp)

	// Output:
	// -1
}

func Compare(a, b string) int {
	// 如果两个字符串字典序相等，返回0
	if a == b {
		return 0
	}
	// 如果a字符串字典序小于b字符串，返回-1
	if a < b {
		return -1
	}
	// 否则返回1
	return +1
}
```

## 字符串读取strings.Reader
`strings.Reader`结构通过实现`Read`从而实现了`Reader`接口。结构如下：

```golang
type Reader struct {
	s        string // 存储的待读取的字符串
	i        int64 // 当前读取到字符串的哪个位置
	prevRune int   // 当按照rune读取字符串时，该字段存储及维护rune的起始位置。从而实现了rune的读取及回滚读取
}

// 构建一个Reader
func NewReader(s string) *Reader

// r剩余的未被读取的字节数，如果已经读取完毕，返回0
func (r *Reader) Len() int

// r底层字符串字节数组的长度
func (r *Reader) Size() int64

// r按字节数组读取底层字符串，如果没有剩余未读字节数，返回io.EOF。
// 返回成功读取到多少个字节和一个错误信息。
// eg: 希望读取5个字节
// buffer := make([]byte, 5)
// n, _ := reader.Read(buffer)
func (r *Reader) Read(b []byte) (n int, err error)

// r从指定offset, 读取len(b)个字节。这种方式，可以定制化的，重复的读取reader的底层字符串
func (r *Reader) ReadAt(b []byte, off int64) (n int, err error)

// 读取r的一个字节。
func (r *Reader) ReadByte() (byte, error)

// 回滚一个被读取的字节（上一个被读取过的字节，重新置为未读取过）
func (r *Reader) UnreadByte() error

// 按字符，读取r的一个字符。返会读取成功的字符，及这个字符占用多少个字节。
func (r *Reader) ReadRune() (ch rune, size int, err error)

// 回滚一个被读取的字符，最近的被读取的字符，回滚未未被读取过。
func (r *Reader) UnreadRune() error

// 寻找一个特定的读取位置。一般也可以使用特定的，比如：
// io.SeekStart (0), 从头开始读取。
// io.SeekCurrent(1) 表示从当前读取位置，加offset的位置，开始后续的读取。
// io.SeekEnd(2) 表示从最大位置加offset的位置开始后续的读取
func (r *Reader) Seek(offset int64, whence int) (int64, error)

// 把字符串reader未读取的剩余内容，写入到io.Writer流中。返回写入成功的字节数
func (r *Reader) WriteTo(w io.Writer) (n int64, err error)

// 重置字符串读取器，通过s重新构建
func (r *Reader) Reset(s string)
```

## 字符串替换strings.Replacer
`strings.Replacer`用于替换字符串中的指定子串。它可以用于一次性替换多个子串, 内部使用线段树结构实现。
```golang
// 构建一个Replacer, 传入替换的规则，需要偶数字符串数组。例如传入x,y字符串。表示x字符串，替换为y字符串。
func NewReplacer(oldnew ...string) *Replacer

// 传入一个目标字符串s, 使用r替换器。把满足替换器规则的，进行替换, 把替换处理好的字符串返回。例如，把s的所有x字符子串，替换为y字符子串。
func (r *Replacer) Replace(s string) string

// 对s, 应用替换器规则。把替换后的结果字符串，写入到w写入流中去。
func (r *Replacer) WriteString(w io.Writer, s string) (n int, err error)
```

## 总结
这里介绍了`golang`中是如何看待拆解及存储字符串信息的，以及字节，字符的概念。顺便研究了一下`strings`包提供的一些处理字符串的能力。篇幅原因，下一个part，会集中认识`strings`包对字符串处理的核心文件`strings.go`。
