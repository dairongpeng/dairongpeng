# 介绍
`time`包是golang标准库中关于时间的包, 提供了程序需要操作时间的相关方法。时间在Golang中以纳秒为精度，表示某一个时间瞬间。

几点说明：

1. 在标准库的文档中，明确指出。时间类型应该使用值类型传递`time.Time`，而不是指针类型传递`*time.Time`。
2. 一般情况下，`time.Time`类型是并发安全的，特例是使用`GobDecode`, `UnmarshalBinary`, `UnmarshalJSON`和`UnmarshalText`等方法用于将数据从二进制、JSON或文本格式解码为Go语言中的结构体对象.这些特殊操作可能会修改内部的数据结构和状态，并且没有做并发的控制。当多个goroutine同时访问这些方法时，需要注意。
3. 两个时间瞬间`time.Time`可以使用`Before`、`After`和`Equal`进行比较。两个时间瞬间的差值，可以产生一个`Duration`类型，表示时间间隔。以此类推，一个时间瞬间`time.Time`加上一个时间间隔`Duration`,会产生另一个时间瞬间`time.Time`。
4. `time.Time`的默认零值是UTC时间`00:00:00.000000000`, 可以使用`IsZero`判断时间是不是没有被初始化过（是否是零值）。
5. 每个时间瞬间`time.Time`都与一个时区相关联。可视化时间时，需要考虑所处的时区。需要注意，一个时间瞬间，不管所处哪个时区，其对应的时间戳都是相等的。
6. 在Go语言中，`==`操作符比较两个时间瞬间时，不仅会比较时间瞬间的值，还会比较时区。这就会造成两个时间瞬间实际是相等的，但是由于时区不同，比较也返回`false`。不区分时区比较两个时间瞬间，可以使用`t1.Equal(t2)`。

## time.Time结构及方法

### 结构说明
```golang
// 一个完整的时间信息, 由wall，ext，loc确定。wall字段表示日期和时间本身，ext字段表示在wall字段基础上可能出现的闰秒偏移量，loc字段表示该时间所处的时区信息。
type Time struct {
    // wall 代表从纪元开始到当前时间经过的纳秒数，它是一个绝对时间的表示。
	wall uint64
    // ext 闰秒的秒数，也就是LeapSecond(闰秒)的数量，闰秒是为了使UTC时间与地球自转的实际时间保持同步。
	ext  int64

    // loc 时区信息，它包含了与地球上所有可能使用的时区的信息，用于将绝对时间（wall）转换为当地的时间。
	loc *Location
}
```

### 结构体方法

#### 格式化输出时间
```golang
// 其中layout指定按什么格式输出，返回一个字符串表示
// 在time.format.go文件中，内置了一些layout常量。
// 例如：
//      const RFC3339     = "2006-01-02T15:04:05Z07:00";
//      const DateTime   = "2006-01-02 15:04:05"
func (t Time) Format(layout string) string

// 按照layout格式化输出，前缀增加b标识。
// 例如：
//      text := []byte("Time: ")
//      text = t.AppendFormat(text, time.Kitchen)
//      fmt.Println(string(text)) // 输出为：Time: 11:00AM
func (t Time) AppendFormat(b []byte, layout string) []byte
```

#### 时间比较
```golang
// t时刻是否在u时刻之后
func (t Time) After(u Time) bool

// t时刻是否在u时刻之前
func (t Time) Before(u Time) bool

// 比较两个时刻，返回-1， 0， 1
func (t Time) Compare(u Time) int

// 不区分时区比较两个时刻
func (t Time) Equal(u Time) bool

// t是否是零时刻，time.Time的默认零值
func (t Time) IsZero() bool
```

#### 时间精度
```golang
// 获取t时刻的年月日信息，其中其中Month是从1开始的, 到12
func (t Time) Date() (year int, month Month, day int)

// 获取t的年信息
func (t Time) Year() int

// 获取t的月信息
// 其中Month是从1开始的, 到12
func (t Time) Month() Month

// 获取t的日信息
func (t Time) Day() int

// 获取当前t是一年中的第几天，如果是非闰年，范围为[1, 365]，如果是闰年范围为[1, 366]
func (t Time) YearDay() int

// 获取t的周信息。其中Weekday是从周日（0）开始的, 到6
func (t Time) Weekday() Weekday

// 获取t时间瞬间，对应的ISO 8601标准格式的周年份和周数.ISOWeek()函数比较适用于需要跨年计算周数的场景，比如某些企业的年度报告中需要使用ISO 8601标准格式计算周数的情况。
// 需要注意的是：ISOWeek()函数只适用于公历时间，并且仅适用于处理UTC时区的时间。在其他时区下的时间可能会造成结果的偏差。
func (t Time) ISOWeek() (year, week int)

// 获取t时间瞬间的时钟信息。当天的小时，分钟和秒。其中hour取值是[0, 23]
func (t Time) Clock() (hour, min, sec int)

// 获取t对应的小时 [0, 23]
func (t Time) Hour() int

// 获取t对应的分钟数
func (t Time) Minute() int

// 获取t对应的秒数
func (t Time) Second() int

// 获取t对应的纳秒数
func (t Time) Nanosecond() int
```

#### 时间区间
```golang
// t时间加上一个时间区间，获取一个新的时间瞬间
func (t Time) Add(d Duration) Time

// 两个时间瞬间相减，获取一个时间区间
func (t Time) Sub(u Time) Duration

// t时间加上多少年，多少月，多少天后，获得一个新的时间瞬间
func (t Time) AddDate(years int, months int, days int) Time
```

#### 时区信息
```golang
// 获取t对应的UTC时区的时间
func (t Time) UTC() Time

// 获取t对应的本机时区的时间
func (t Time) Local() Time

// 获取t对应的指定时区的时间。如果loc为nil，会panic
func (t Time) In(loc *Location) Time

// 获取t的时区信息
func (t Time) Location() *Location

// 获取t的时区简称，及偏移量
func (t Time) Zone() (name string, offset int)

// 判断给定的时间t是不是处在夏令时
// 美国一直实行夏令时。而中国考虑掉弊大于利，已经取消，改用夏季作息表。
func (t Time) IsDST() bool
```

#### 时间戳信息
```golang
// 返回自1970.1.1时间以来的秒数，与时区无关
func (t Time) Unix() int64

// 返回自1970.1.1时间以来的毫秒数，与时区无关
func (t Time) UnixMilli() int64

// 返回自1970.1.1时间以来的微秒数，与时区无关
func (t Time) UnixMicro() int64

// 返回自1970.1.1时间以来的纳秒数，与时区无关
func (t Time) UnixNano() int64
```

#### 序列化和反序列化
注意在`UnmarshalBinary`, `GobDecode`, `UnmarshalJSON`, `UnmarshalText`需要使用指针接受者，这是因为在这些函数内部，需要更改Time内的变量值。这些特殊操作可能会修改内部的数据结构和状态，并且没有做并发的控制。当多个goroutine同时访问这些方法时，需要注意。

```golang
// MarshalBinary 用于将时间对象序列化为二进制格式
func (t Time) MarshalBinary() ([]byte, error)
// UnmarshalBinary 用于将二进制格式反序列化为时间对象
func (t *Time) UnmarshalBinary(data []byte) error

// 与MarshalBinary类似
func (t Time) GobEncode() ([]byte, error)
// 与UnmarshalBinary类似
func (t *Time) GobDecode(data []byte) error

// json序列化时间
func (t Time) MarshalJSON() ([]byte, error)
// json反序列化时间
func (t *Time) UnmarshalJSON(data []byte) error

// text序列化
func (t Time) MarshalText() ([]byte, error)
// text反序列化
func (t *Time) UnmarshalText(data []byte) error
```

#### 其他补充
```golang
// Truncate按d截断时间，比如d为Hour，则会截断t的秒和纳秒信息只保留到小时粒度。
func (t Time) Truncate(d Duration) Time

// Round按d舍入，比如d为time.Minute，则将时间舍入到分钟精度。(与Truncate类似，一个是截断，一个是舍入)
func (t Time) Round(d Duration) Time
```

## time.Duration结构及方法
### 格式化输出时间区间
```golang
// 按{xx}h{xx}m{xx}.{xxx}s格式输出时间区间。如果不含有纳秒时间，则{xx}h{xx}m{xx}s
// 例如：
//      纳秒级别：1h15m30.918273645s 
//      微秒级别：1h15m30.918274s
//      毫秒级别：1h15m30.918s
//      秒级别： 1h15m31s
//      分钟级别：1h16m0s
//      小时级别：1h0m0s
func (d Duration) String() string
```

### 时间戳精度
```golang
// 返回时间区间对应的纳秒数
func (d Duration) Nanoseconds() int64

// 返回时间区间对应的微秒数
func (d Duration) Microseconds() int64

// 返回时间区间对应的毫秒数
func (d Duration) Milliseconds() int64

// 返回时间区间对应的秒数
func (d Duration) Seconds() float64

// 返回时间区间对应的分钟数
func (d Duration) Minutes() float64

// 返回时间区间对应的小时数
func (d Duration) Hours() float64
```

### 其他补充
```golang
// 按m粒度，截断d。与时间的截断类似
func (d Duration) Truncate(m Duration) Duration

// 按m粒度，舍入d。与时间的舍入类似
func (d Duration) Round(m Duration) Duration

// 返回时间区间d的绝对值。特例：math.MinInt64会返回math.MaxInt64
func (d Duration) Abs() Duration
```

## time包函数
```golang
// 返回给定时间t距离当前时间的时间间隔
func Since(t Time) Duration

// Until返回当前时间距离给定时间t, 存在多少时间间隔
func Until(t Time) Duration

// 获取当前时间，对应的本地时区的时间瞬间
func Now() Time

// 给定秒时间戳信息，含秒和纳秒信息。构建本地时区的时间瞬间，从UTC 1970.1.1开始
func Unix(sec int64, nsec int64) Time

// 给定毫秒时间戳, 构建本地时区的时间瞬间，从UTC 1970.1.1开始
func UnixMilli(msec int64) Time

// 给定微秒时间戳，构建本地时区的时间瞬间，从UTC 1970.1.1开始
func UnixMicro(usec int64) Time

// 给定年，月，日，时，分，秒，纳秒信息。以及时区信息，构建该时区对应的时间瞬间Time
func Date(year int, month Month, day, hour, min, sec, nsec int, loc *Location) Time
```

# 总结

Golang中的time包，整体还是比较简单易懂的。一般来说，我们使用time都不涉及更改内部的`wall`，`ext`和`loc`信息，直接传递值time.Time类型即可。

需要注意的是，在时间类型的序列化和反序列化中，使用的一些函数例如`GobDecode`, `UnmarshalBinary`, `UnmarshalJSON`和`UnmarshalText`, 存在赋值time.Time类型内的内部字段。此时要确保指针传递，且这些函数内未处理并发，我们还需额外考虑是否存在并发问题。

在梳理Time包时，先理清楚结构，比如`time.Time`结构，`time.Duration`结构，再理解结构提供了哪些能力给调用者调用。整体脉络就清晰了。