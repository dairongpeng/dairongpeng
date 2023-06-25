Fmt包是Golang语言格式化输入输出相关的包。包内主要文件为`format.go`, `print.go`和`scan.go`，`print.go`的主要功能是格式化输出到Stdout流，而`scan.go`的主要功能是从Stdin流格式化输入到程序。`format.go`提供给标准库内部调用，都是一些私有方法，含有一个主要的结构`fmt`，用来处理格式的内部格式的转换。

一般操作系统标准输入输出相关的流有`Stdin`，`Stdout`和`Stderr`。`Stdin`是程序的标准输入流，`Stdout`是程序的标准输出流，而`Stderr`是输出流的一种，常常用来输出错误，警告等异常信息。一般来说`Stdout`可以通过重定向到文件，而`Stderr`通常直接限制在终端或控制台，不会被重定向，这样可以确保错误信息能够及时显示给用户或开发者。

# 格式化控制台输出
`print.go`包含13个公共方法。这里以熟悉提供给上层应用调用的这13个方法，来熟悉Golang格式化输出的能力。公共方法内部，标准库会使用一些自己内部的结构，例如`pp`, `bufer`, `fmt`结构及其结构的方法，用来缓存需要输出的信息，格式化类型等一些内部逻辑，这里不再关心。

## 获取格式化状态机标识
```golang
// State是接口类型，传入实现State接口的结构体， 比如stat, 用来构建打印状态。
//      1. stat在FormatString方法内需要调用Width，获取打印机状态的宽度
//      2. stat在FormatString方法内需要调用Precision，获取打印机状态的精度
//      3. stat在FormatString方法内需要调用Flag，获取打印机状态的标识，例如 ' ', '+', '-', '#', '0'
// verb表示格式化动作，例如'x', 'd'等，用来构建打印动作。
// 例如：宽度为0，精度为0，flag为'',构建stat，verb为'x'。 ==> FormatString输出"%x"
// 例如：宽度为0，精度为3，flag为'',构建stat，verb为'x'。 ==> FormatString输出"%.3x"
// 例如：宽度为7，精度为3，flag为'+',构建stat，verb为'x'。 ==> FormatString输出"%+7.3x"
// 参考：state_test.go#74
func FormatString(state State, verb rune) string

type State interface {
	Write(b []byte) (n int, err error)
	Width() (wid int, ok bool)
	Precision() (prec int, ok bool)
	Flag(c int) bool
}
```

## 格式化类型
Golang格式化输出的格式化类型，包含几个组成部分。`宽度`, `精度`, `动词Verb`, `宽度`, `精度`用来控制输出的格式, `动词Verb`负责按什么类型转化传入的变量。

### 宽度

跟在`%`后的一个正整数，用来对输出结果填充位宽。

- `fmt.Printf("Name: %10s\n", "Alice")`: 输出为：`Name:      Alice`， 其中`Alice`前填充了5个空格，确保该字段的宽度为10。

### 精度

跟在`%`后，使用点号`.`紧跟精度信息。用于控制浮点数的小数点后的位数或字符串的截断长度。

- `fmt.Printf("Pi: %.2f\n", 3.14159265359)`: 输出为：`Pi: 3.14`， 保留了小数点后的2位精度。

### 动词Verb

跟在`%`后的一个字符表示。表示特定的格式化类型。下面列举一些常见的动词组合。

### Flag

跟在`%`后，动词前的标识，用来指定格式化行为和输出的格式。这些标识可以在格式化动词前，执行默认输出的行为。

- `+`: 在数值类型的输出中，如整数，浮点数。默认情况下不会显示正负号。使用`+`会强制显示正负号。比如输出为`+1`， `-1`
- `-`: 左对齐输出。默认情况下，使用空格进行右对齐输出，使用`-`标识可以改为左对齐输出。
- `#`: 用于改变输出的格式或添加附加信息。它的具体效果取决于具体的格式化动词。比如对于 `%#x` 动词，它会在输出十六进制时添加 `0x` 前缀；对于 `%#o` 动词，它会在输出八进制时添加 `0` 前缀。
- ` `: （空格）在正数前面添加空格，而负数前面添加负号。如果使用 + 标识，则空格标识无效。
- `0`: 在数值类型的输出中，使用 `0` 标识可以在数值前面填充零，以达到指定的输出宽度。默认情况下，使用空格进行填充。

## 类型解释

### 通用类型
- `%v`: 根据变量的类型选择合适的格式。
- `%+v`: 以扩展格式输出结构体，包括字段名。
- `%#v`: 以`Go`语法表示的值的格式输出，包括类型和值的详细信息。
- `%T`: 输出变量的类型。

通用类型接受其他任何类型都可以，比如可以接受布尔类型，整数类型，字符串类型，结构体类型等。接受其他类型时会存在默认接受类型。

1. 当变量是布尔`bool`类型时，`%v`默认使用`%t`来接收。
2. 当变量是整数类型时`int`, `int8`, `int32`等。`%v`默认使用`%d`来接收。
3. 当变量是无符号整数类型时`uint`, `uint8`, `uint32`等。`%v`默认使用`%d`来接收, 如果`%#v`则默认使用`%#x`。
4. 当变量是浮点类型时`float32`, `float64`等。`%v`默认使用`%g`来接收。
5. 当变量是字符串类型时，`%v`默认使用`%s`来接收。
6. 当变量是通道类型`chan`，指针类型时，`%v`默认使用`%p`来接收。
7. 当变量是结构体类型，则`%v`默认使用`{field0 field1 ...}`
8. 当变量是`array`, `slice`类型时时，则`%v`默认使用`[elem0 elem1 ...]`
9. 当变量是`map`类型时，则`%v`默认使用`map[key1:value1 key2:value2 ...]`
10. 当变量是指针数组，指针切片，指针map时，`%v`默认使用`&{}, &[], &map[]`, 元素为指针的地址。

### 布尔类型
- `%t`: 接受输出变量的布尔类型。`true`或者`false`。

### 整数类型
- `%b`: 输出整数的二进制表示。
- `%c`: 输出整数Unicode码对应的字符。
- `%d`: 输出整数的十进制表示。
- `%o`及`%O`: 输出整数的八进制表示。
- `%q`: golang安全输出字面量，需要转义的会转义。
- `%x`及`%X`: 输出整数的十六进制表示。

### 浮点数
- `%b`: 输出浮点数二进制表示。
- `e`及`%E`: 输出浮点数为科学计数法表示。科学计数法表示将浮点数表示为尾数乘以 10 的指数次幂的形式。在科学计数法中，浮点数被表示为 `m x 10^n`，其中`m`是尾数，`n`是指数。
- `%f`及`%F`: 输出浮点数的十进制表示。
- `%g`及`%G`: 输出浮点数合适的且较短的方式表示, 避免冗长。小指数则使用`%e`或`%E`表示，否则用`%f`或`%F`表示。
- `%x`及`%X`: 输出浮点数的十六进制表示。

### 字符串
- `%s`: 输出字符串的原始内容，不尽兴任何转义或引号添加。
- `%q`: 输出字符串的带引号标识，并对字符串中的特殊字符进行转义。

### 切片
- `%p`: 输出切片的起始地址，即[0]元素的地址十六进制表示。

### 指针
- `%p`: 输出指针类型的地址十六进制表示，同样的`%b`, `%d`, `%o`, `%x`等也可以接受指针类型，分别是指针地址的不同进制的表示。

## 格式化输出源码方法
```golang
// 格式话输出写入到w。其中w是一个接口，即可以写入到任何实现io.Writer接口的实现
func Fprintf(w io.Writer, format string, a ...any) (n int, err error)

// 格式化输出字符串到标准输出设备Stdout, 
func Printf(format string, a ...any) (n int, err error)

// 格式化字符串，并把格式化后的结果字符串返回给调用方
func Sprintf(format string, a ...any) string

// 追加字节数组，把a格式化后的结果，追加到传入的字节数组中，把最终字节数组结果返回
func Appendf(b []byte, format string, a ...any) []byte

// 把传入的内容，使用默认格式化，写入到io.Writer设备, 返回io.Writer写入到的位置
func Fprint(w io.Writer, a ...any) (n int, err error)

// 标准输出设备，默认格式化a, 输出a, 其中a可以是多个值
func Print(a ...any) (n int, err error)

// 采用默认格式化， 格式化a, 格式化后的结果返回调用方 
func Sprint(a ...any) string

// 追加字节数组，把a默认格式化后的结果，追加到传入的字节数组中，把最终字节数组结果返回
func Append(b []byte, a ...any) []byte

// 把a默认格式化且换行，追加到io.Writer实现的结构中，返回io.Writer写入的位置
func Fprintln(w io.Writer, a ...any) (n int, err error)

// 把a默认格式化，输出到标准输出设备Stdout中,返回标准输出的写入位置
func Println(a ...any) (n int, err error)

// 把a默认格式化，返回格式化后的字符串
func Sprintln(a ...any) string

// 把a默认格式化，输出到输入的b中，把最终结果返回
func Appendln(b []byte, a ...any) []byte
```

## 格式化输入源码方法
```golang
// 传入a, 等待标准输入Stdin的键入内容，绑定到该变量。返回绑定变量的数量。
// 从标准输入读取输入，并按空白字符分隔赋值给传递的参数。
func Scan(a ...any) (n int, err error)

// 传入a, 等待标准输入Stdin的键入内容，绑定到该变量。返回绑定变量的数量。
//  从标准输入读取一行输入，并按空白字符分隔赋值给传递的参数。
func Scanln(a ...any) (n int, err error)

// 传入a, 等待标准输入Stdin的键入内容，绑定到该变量。返回绑定变量的数量。按照指定format进行绑定。
// 从标准输入读取输入，按format顺序绑定。例如：format='%s %d %f', 输入'zhangsan 18 172.5'
func Scanf(format string, a ...any) (n int, err error)

// 把字符串作为输入，读取字符串中的内容，按空白符分割，绑定到a。返回绑定成功几个元素
func Sscan(str string, a ...any) (n int, err error)

// 把字符串作为输入，读取字符串内换行前的的内容，即读取一行，把该行按空格分割，绑定到a。返回绑定成功几个元素
func Sscanln(str string, a ...any) (n int, err error)

// 把字符串作为输入，读取字符串中的内容，按指定的format依次读取，绑定到a。返回绑定成功几个元素
func Sscanf(str string, format string, a ...any) (n int, err error)

// 从指定的io.Reader中按空白字符分隔读取输入。它将读取并解析输入，直到遇到空白字符（空格、制表符、换行符）为止。该函数返回成功读取的项目数量。
func Fscan(r io.Reader, a ...any) (n int, err error)

// Fscanln函数与Fscan函数类似，不同之处在于它在处理换行符时会将换行符视为输入项的分隔符。它会读取整行输入，只读取一行，并将输入按空白字符分隔为多个项。与 Fscan 类似，它也返回成功读取的项目数量。
func Fscanln(r io.Reader, a ...any) (n int, err error)

// Fscanf函数使用指定的格式字符串来读取输入。它根据格式字符串的规则解析输入，并将解析的结果存储在指定的变量中。格式字符串中的特殊字符指定了要读取的值的类型和顺序。例如: format为 %s %d %f， 那么按照一个字符串，一个整数，一个浮点数依次读取标准输入。
func Fscanf(r io.Reader, format string, a ...any) (n int, err error)
```

### 一些案例
- 标准输入按空白符分割读取绑定

```golang
func main() {
    var name1, name2 string
    fmt.Print("Enter your name: ")
    n, _ := fmt.Scan(&name1, &name2)
    fmt.Println(n)
    fmt.Println("Name1:", name1, " Name2:", name2)

    // Output:
    // Enter your name: zhangsan lisi
    // 2
    // Name1: zhangsan  Name2: lisi
}
```

- 字符串输入，按行读取，再按空白符分割，读取绑定

```golang
func main() {
	str := "zhangsan 18 172.5\n读取不到我"

	var name string
	var age int
	var height float64
	n, _ := fmt.Sscanln(str, &name, &age, &height)

	fmt.Println(n)
	fmt.Println("name:", name, " age:", age, " height:", height)

	// Output:
	// 3
	// name: zhangsan  age: 18  height: 172.5
}
```

- 从标准输入读取，按照空格分割读取，并按照给定的format顺序，绑定

```golang
func main() {
	var name string
	var age int
	var height float64
	fmt.Print("Enter: ")
	n, _ := fmt.Scanf("%s %d %f", &name, &age, &height)

	fmt.Println(n)
	fmt.Println("name:", name, " age:", age, " height:", height)

	// Output:
	// Enter: zhangsan 18 172.5
	// 3
	// name: zhangsan  age: 18  height: 172.5
}
```