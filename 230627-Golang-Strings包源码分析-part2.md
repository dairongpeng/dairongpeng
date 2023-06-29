`strings`包中`strings.go`文件中的方法分析，包含字符串统计，字符串切割，字符串转换，字符串处理等。

## 字符串统计
```golang
// 统计s中含有多少个非重叠子串substr。非重叠子串意味着，s=eeeee, substr=ee, 则统计为2
// 如果substr是空串，则统计s中含有的（字符数+1）, 比如s=five，可以理解为'' + 'f' + '' + 'i' + '' + 'v' + '' + 'e' + '', 字母都是字符，则''为字符数加1，5个。
func Count(s, substr string) int

// 检查统计，substr在字符串s中是否存在。空串'', 被认为存在任何字符串中
func Contains(s, substr string) bool

// 检查统计，s是否包含chars字符集合的任何一个字符。
// 常被用来做字符的校验。比如检查一个字符串s是否含有特殊字符。就可以让检查的字符串s，去ContainsAny一个特殊字符的集合，例如";:!?+*/=[]{}_^°&§~%#@<\">\\", 只要s包含集合内的任意一个字符，就返回true
func ContainsAny(s, chars string) bool

// 检查统计，s是否包含Unicode码点r。
func ContainsRune(s string, r rune) bool

// 检查s的各个字符，经过回调函数f的处理，能否为true。如果其中一个字符，被f处理为真，整体返回真。用来扩展字符串的自定义存在规则的查找
func ContainsFunc(s string, f func(rune) bool) bool
```

## 字符串搜索
```golang
// 查找substr在s中是否存在，如果存在，返回最后出现的位置。不存在返回-1位置
func LastIndex(s, substr string) int

// 查找字符串s是否包含字节c，如果存在，返回字节c第一次出现的位置。不存在则返回-1
func IndexByte(s string, c byte) int

// 查找字符串s是否包含Unicode码值r，如果存在，返回r第一次出现的位置（注意是该字符，占底层数组的起始位置）。不存在返回-1
// eg: strings.IndexRune("世界第一等", '界') 返回 3 而不是 1
func IndexRune(s string, r rune) int

// 查找s中是否含有chars字符集合中的任意一个字符，找到chars中第一个存在的字符，则返回该字符占用的底层字节数组的起始位置。所有字符都找不到返回-1
func IndexAny(s, chars string) int

// 同IndexAny，查找s中是否含有chars字符集合中的任意一个字符，找到s满足chars集合元素的最后一个位置，不存在返回-1
func LastIndexAny(s, chars string) int

// 查找字节c在s中是否存在，存在则返回最后出现的位置。
func LastIndexByte(s string, c byte) int

// 查找字符串s中满足f函数的字符，找到的第一个满足f的位置返回
func IndexFunc(s string, f func(rune) bool) int

// 查找字符串s中满足f函数的字符，找到的最后一个满足f的位置返回
func LastIndexFunc(s string, f func(rune) bool) int

// 返回substr在s中是否存在，返回第一次数出现的位置
func Index(s, substr string) int
```

## 字符串切割
```golang
// 按指定分隔符，分割为子字符串数组。s是源字符串，sep是分隔符，n指定拆分次数。
// n > 0的时候，从头开始，按分隔符切出来最多n个片段。第n个片段，是字符串剩余的部分，可能仍然包含sep。满足的片段达不到n的，有多少满足的返回多少。
// n = 0时，返回空切片。
// n < 0时，返回所有满足sep切分的子字符串数组。等价于Splic
func SplitN(s, sep string, n int) []string

// 返回所有的满足sep切割的，子字符串列表、如果没有满足sep的切割，则返回源字符串s作为切割的唯一元素。
func Split(s, sep string) []string

// SplitAfterN和SplitN的区别，就是结果是否保留源字符串的分隔符。 SplitAfterN结果会保留sep。
// 如果不保留，就类似["one" "two" "three" "four" "five"]; 如果保留则["one," "two," "three," "four," "five"]
func SplitAfterN(s, sep string, n int) []string

// 同Split， SplitAfter结果中会保留sep
func SplitAfter(s, sep string) []string

// 按照空白符（空格，制表符，换行符等）, 将原字符串切分为子字符串数组。
func Fields(s string) []string

// 自定义拆分规则，切割源字符串s。
// 自定义拆分函数的入参是字符，表示字符串中的每个字符，返回bool字符串中的某个字符满足拆分的分隔符, 即如果返回true，则表示当前字符被视为分隔符。
// 返回的切片中不包含，空字符串。
// 例如：func(r rune) bool { return r == ' ' || r == '.' }
func FieldsFunc(s string, f func(rune) bool) []string
```

## 字符串拼接
```golang
// 通过sep拼接符，把elems中的所有元素，拼接成一个字符串，返回
func Join(elems []string, sep string) string
```

## 校验字符串
```golang
// 检查字符串s,是否以prefix开头
func HasPrefix(s, prefix string) bool

// 检查字符串s,是否以suffix结尾
func HasSuffix(s, suffix string) bool
```

## 字符串变换
```golang
// 字符串字母转大写。非字母字符不作处理。
func ToUpper(s string) string

// 字符串字母转小写
func ToLower(s string) string

// 与ToUpper类似，字母字符也转大写。区别在于，非字符字母，也会处理成标题形式，例如一些其他符号。在标准库中也会定义需要转换标题的字符。
func ToTitle(s string) string

// 将给定的字符串转换为有效的 UTF-8 编码的新字符串。该函数，会修复或删除无效的 UTF-8 字节序列，使字符串符合 UTF-8 编码规范。
// 如果 replacement 为空字符串，则无效的字节序列会被删除。
func ToValidUTF8(s, replacement string) string
```

## 字符串处理
```golang
// 根据映射规则mapping, 对字符串中的每个字符进行转换或删除. 返回经过处理的字符串
func Map(mapping func(rune) rune, s string) string

// s重复count份，返回重复后的结果。例如：strings.Repeat("A", 3) // AAA
// 内部通过strings.Builder实现。
// 需要考虑字符串最终长度, 是否整数越界问题。
func Repeat(s string, count int) string

// 从字符串左侧开始剔除满足f条件的字符
func TrimLeftFunc(s string, f func(rune) bool) string

// 从字符串右侧开始剔除满足f条件的字符
func TrimRightFunc(s string, f func(rune) bool) string

// 字符串左侧和右侧同时剔除满足f条件的字符
func TrimFunc(s string, f func(rune) bool) string

// 从字符串s的左右两侧，剔除满足cutset内的字符。eg: cutset="0123456789"，那么字符串s的左右两侧是数字都会被剔除掉
func Trim(s, cutset string) string

// 与Trim类似，区别是只从左边开始剔除
func TrimLeft(s, cutset string) string

// 与Trim类似，区别是只从右边开始剔除
func TrimRight(s, cutset string) string

// 剔除s的左右两端的空白符（空格，制表符，回车符号）
func TrimSpace(s string) string

// 剔除s左边的prefix字符串的前缀
func TrimPrefix(s, prefix string) string

// 剔除s右边suffix字符串的后缀
func TrimSuffix(s, suffix string) string
```

## 字符串替换
```golang
// 替换字符串。源字符串s，需要替换的子串old，替换为字符串new。
// n < 0时，所有满足新旧替换的子串都会被替换
// n == 0时，替换个数为0.原字符串直接返回
// n > 0时，返回替换n个满足条件子串后的结果，生成一个新的字符串。
// eg: strings.Replace("99999", "9", "1", 3) // 11199
func Replace(s, old, new string, n int) string

// 与Replace类似，相当于n无限大的情况，不控制替换的个数，全部替换
func ReplaceAll(s, old, new string) string
```

## 字符串比较
```golang
// 比较两个字符串，大小写不敏感，内部是都转换为小写，进行字典序比较相等。
// 注意：区分大小写比较字典序，直接使用 s == t，s > t， s < t即可。也就是之前的strings.Compare(a, b string) int的实现
func EqualFold(s, t string) bool
```

## 字符串裁剪
```golang
// 通过sep间隔符，找到s第一个匹配的sep, 对s进行截取。
// 找到关键字，进行截取，返回截取的前一段和后一段，以及true
// 找不到关键字，返回s和""， 以及false
// eg: 
// Cut("Gopher", "Go") = "", "pher", true
// Cut("Gopher", "ph") = "Go", "er", true
// Cut("Gopher", "er") = "Goph", "", true
// Cut("Gopher", "Badger") = "Gopher", "", false
// Cut("Gophpher", "ph") = "Go", "pher", true
func Cut(s, sep string) (before, after string, found bool)

// 是Cut的特例，只从s的头部查找是否匹配prefix，进行截取，找不到匹配的关键字，返回false
// 如果找到匹配，截取后，只会留下后一段
func CutPrefix(s, prefix string) (after string, found bool)

// 是Cut的特例，只从s的尾部查找是否匹配suffix，进行截取，找不到匹配的关键字，返回false
// 如果找到匹配，截取后，只会留下前一段
func CutSuffix(s, suffix string) (before string, found bool)
```