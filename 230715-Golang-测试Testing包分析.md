Golang的标准库中包含了一个名为`testing`的包，它提供了一套用于编写单元测试和性能测试的工具和框架。testing包是Go语言中测试相关功能的核心部分，它使得编写和运行测试变得简单和高效。(ChatGPT)

下面是testing包的一些重要组件和功能：

1. **测试函数**：testing包要求测试函数以"Test"开头，并且函数签名必须是`func TestXxx(t *testing.T)`，其中"Xxx"可以是任意字符串。这些测试函数用于编写单元测试逻辑。
2. **测试报告**：测试函数中可以使用`testing.T`提供的方法来记录测试结果和错误。`testing.T`提供了`Fail、Error、FailNow`等方法用于标记测试失败，以及`Log、Logf`等方法用于输出日志信息。测试运行完成后，测试报告将显示每个测试函数的运行结果。
3. **子测试**：testing包支持在测试函数中使用子测试，这样可以将一个大的测试函数拆分为多个小的子测试函数，每个子测试函数都有自己的名称和独立的测试逻辑。
4. **示例函数**：除了测试函数，testing包还支持示例函数，用于展示某个函数或方法的使用方式和预期输出。示例函数以"Example"开头，并且函数签名必须是`func ExampleXxx()`。示例函数的执行结果将作为文档的一部分输出。
5. **基准测试**：testing包还提供了性能测试的功能，称为基准测试。基准测试用于评估代码的性能指标，如函数的执行时间。基准测试函数以`Benchmark`开头，并且函数签名必须是`func BenchmarkXxx(b *testing.B)`。基准测试的运行次数由testing.B提供，并且可以通过-bench参数设置。
6. **表格驱动测试**：testing包支持表格驱动测试，这种测试方式通过使用测试数据表格来覆盖多个测试用例。可以使用结构体、切片或者其他数据结构来表示测试数据表格，然后在测试函数中遍历表格进行测试。
7. **测试覆盖率**：testing包内置了对测试覆盖率的支持。可以通过在测试过程中收集代码覆盖率信息，并通过go test命令生成覆盖率报告。覆盖率报告可以显示每个文件和每个函数的覆盖率百分比，帮助开发者评估测试的完整性。

这些是testing包的一些主要特性和功能。使用testing包可以编写全面的单元测试和性能测试，帮助开发者保证代码的质量和性能。需要指出的事，`testing`包内包含大量的文件，比较细节且内容比较多，这里值分析少量源码，并且从主要功能的演示来认识`golang test`包。

## 单元测试
首先想要做单元测试，则必须对响应的文件名，方法名，参数都进行限制, 满足一下规则。
1. 测试文件名，满足`XXX_test.go`的形式。
2. 测试函数，满足`Test[XXX]`的形式。
3. 测试函数从参数需要携带`t *testing.T`参数。
4. 测试文件需要和被测试函数处在同一个包下。

如果我们使用集成开发环境，则有些工具默认可以一键生成，例如Goland可以`command + n`生成测试函数。选择方法后`Test for selection`为方法生成单侧。选择`Test for file`为当前文件所有方法生成单测。选择`Test for package`为当前包内的所有方法生成单测。

```golang
func Add(a, b int64) int64 {
	return a + b
}

// 生成的测试函数形如。在同包下的xxx_test.go文件中
func TestAdd(t *testing.T) {
	type args struct {
		a int64
		b int64
	}
    // 表格驱动测试。
	tests := []struct {
		name string
		args args
		want int64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Add(tt.args.a, tt.args.b); got != tt.want {
				t.Errorf("Add() = %v, want %v", got, tt.want)
			}
		})
	}
}
```

### 运行测试用例函数
运行单元测试函数命令，需要使用到`go tools`工具命令。

1. `go test`默认会执行当前包下的所有测试用例。
2. `-v`指定option，可以显示每个用例的测试结果。
3. `-run`指定option，可以运行某一个或某一些（基于通配符和部分正则`*`,`^`,`$`）指定的测试用例。
4. `-cover`指定option，观测用例的覆盖率。

```shell
# go test 默认执行包内的所有测试用例。按照CPU核心数并发。如果希望串行执行包内的所有测试用例，可以增加 -parallel 1 控制
➜  unit ls
unit.go      unit_test.go
➜  unit go test 
PASS
ok      alltest/unit    0.441s

# 显示每个用例的执行情况 -v
➜  unit go test -v
=== RUN   TestAdd
--- PASS: TestAdd (0.00s)
=== RUN   TestSub
--- PASS: TestSub (0.00s)
PASS
ok      alltest/unit    0.432s

# 指定运行某个测试用例 -run
➜  unit go test -v -run TestAdd
=== RUN   TestAdd
--- PASS: TestAdd (0.00s)
PASS
ok      alltest/unit    0.167s

# 报告单侧的覆盖率
➜  unit go test -v -run TestAdd -cover
=== RUN   TestAdd
# 用例1执行
=== RUN   TestAdd/case1
# 用例2执行
=== RUN   TestAdd/case2
--- PASS: TestAdd (0.00s)
    --- PASS: TestAdd/case1 (0.00s)
    --- PASS: TestAdd/case2 (0.00s)
PASS
        alltest/unit    coverage: 66.7% of statements # 覆盖了测试方法，66.7%的语句
ok      alltest/unit    0.419s
```

### testing.T
单元测试的核心结构`T`组合了`testing`包的`common`，从而实现了`日志输出`, `断言`, `报错`等相关控制。

#### T

```golang
type T struct {
	common // 组合了common
	isEnvSet bool
	context  *testContext // For running tests and subtests.
}


// 通过t，运行子测试。name指定子测试的名字。支持并发调用。每个子测试会报告自身的测试信息。
func (t *T) Run(name string, f func(t *T)) bool

// 运行单元测试时，可能会依赖环境变量信息，SetEnv会调用os.Setenv设置环境变量。
// 当t运行完毕后，会清理掉Cleanup设置到t中的环境变量。
// 需要注意：设置变量不能使用在并发子测试中。即，在当前t中设置了setenv, 就避免在当前t中再创建子测试。
// if the current test or any parent is parallel. break
func (t *T) Setenv(key, value string)

// 指定当前测试，可以被测试框架并行运行当前测试函数。
// 单元测试中，每个测试函数通常来说是相互独立的。意味着他们可以独立的执行而不会互相干扰。
// 不设置，按照指定的CPU核心数并发。设置后，在并发的基础上，再按照协程并发。
func (t *T) Parallel()

// eg:
func TestF(t *testing.T) {
    t.Parallel() // 声明该测试函数可以并行。
    // 测试逻辑...
}

// 获取t的截止时间。再根据ok的值确定获取截止时间是否成功。
// 设置测试的超时时间，可以通过运行测试时设置。例如：go test -timeout 10s。 再在具体的测试用例中检查，是否达到了截止时间。
func (t *T) Deadline() (deadline time.Time, ok bool)
```

#### common
`common`是一个基础结构，在`testing`包中, 很多public结构，都组合了common，从而实现一些公共能力。
```golang
// 获取当前测试的名字
func (c *common) Name() string

// 标识当前测试最终不成功，但是不中断测试。
func (c *common) Fail()

// 检查此时当前测试，是否已经被之前代码判断为不成功（Fail），如果是则返回true。
func (c *common) Failed() bool

// 标识当前测试失败，并且立刻退出当前测试。与Fail()方法不一样的点，是这个方法会立刻退出。
func (c *common) FailNow()

// 当前测试打印log, 和格式化打印log
func (c *common) Log(args ...any)
func (c *common) Logf(format string, args ...any)

// 当前测试打印错误Err，测试最终被标记为失败。不会终端测试当前。
func (c *common) Error(args ...any)
func (c *common) Errorf(format string, args ...any)

// 当前测试打印错误Fatal，测试立刻标记为失败。直接退出，与Err区别是会直接退出当前测试。
func (c *common) Fatal(args ...any)
func (c *common) Fatalf(format string, args ...any)

// 当前测试标记被SKIP状态，退出当前测试，但是不影响其他后续的测试用例。例如我Skip case2，最终效果形如下：
// --- PASS: TestAdd (0.00s)
//  --- PASS: TestAdd/case1 (0.00s)
//  --- SKIP: TestAdd/case2 (0.00s)
//  --- PASS: TestAdd/case3 (0.00s)
// PASS
// ok      alltest/unit    0.399s
func (c *common) SkipNow()

// 跳过当前测试，标记为SKIP状态，并打印个日志。
func (c *common) Skip(args ...any)
func (c *common) Skipf(format string, args ...any)

// 获取当前测试用例，是否是被标记为SKIP状态
func (c *common) Skipped() bool

// 标记当前测试函数，为辅助函数。辅助函数通常是为了支持测试函数的执行，而不是直接进行测试断言的函数。
// 通过标记为辅助函数，测试框架可以提供更好的错误报告和跟踪，以帮助你定位问题。
// 测试框架会在出现测试失败时，优先报告调用辅助函数的位置，而不是辅助函数内部的具体代码位置。
func (c *common) Helper()

// 为当前测试用例，增加一个收尾函数。该函数在当前用例，以及其所有子用例运行完毕时，调用。
// 如果一个用例定义了多个Cleanup，则后定义的先执行。类似于defer
func (c *common) Cleanup(f func())

// 当前测试用例，用于创建一个临时目录，供测试使用。
// 返回一个字符串，表示创建的临时目录的路径。该临时目录在测试函数执行完毕后会自动被清理。
func (c *common) TempDir() string

// 为当前测试用例，设置环境变量。
func (c *common) Setenv(key, value string)
```

### testing.M

与`testing.T`的区别：
- `testing.M`结构体表示一个测试的运行环境，用于执行测试程序的入口点。
通过调用`testing.Main`函数创建`testing.M`结构体，并调用其`Run()`方法来运行测试。
- `testing.M`结构体包含用于管理和控制测试执行过程的字段和方法，如 `Name`、`Run()`、`Fail()`、`Failed()`、`ImportPath`等。
- `testing.M`结构体主要用于控制测试的执行、处理测试失败和执行其他与测试相关的操作。
- `testing.T`结构体表示单个测试函数的运行环境，用于执行单个测试函数。
- `testing.T`结构体是在测试函数中作为参数传递的，通常用`t`来表示。
- `testing.T` 结构体提供了许多用于断言和报告测试结果的方法，如 `Error()`、`Fatal()`、`Skip()`、`Helper()`、`Parallel()` 等。
- `testing.T` 结构体用于编写测试函数，通过断言方法判断测试结果是否符合预期，并报告测试的成功或失败。

需要注意的是，`TestMain`函数也一定要在`xxx_text.go`中定义, 且需要让`TestMain`运行包下的所有用例，执行`go test -v .`， 如果希望仅仅运行TestMain所在的文件内的用例，可以`go test -v main_test.go`。总结起来，`testing.M` 结构体用于控制和管理整个测试程序的执行过程，而 `testing.T` 结构体则用于在单个测试函数中进行断言和报告测试结果。两者在测试中扮演不同的角色，分别用于控制测试的执行和实际测试的断言和报告。

```golang
package main_test

import (
	"testing"
)

func TestMain(m *testing.M) {
	// before test的时机。可以用来设置环境变量，做一些初始化
	fmt.Println("I am Before")

	// 运行测试。这里会调用TestA, TestB
	exitCode := m.Run()

	fmt.Println("I am After")
	// after test的时机，可以用来关闭链接，清理一些结果数据等。

	// 根据测试结果返回适当的退出码
	// 可以根据需要返回不同的退出码，例如根据测试失败的数量来返回相应的值
	// 0 表示测试成功，非零值表示测试失败
	// 这里只是简单地返回测试的退出码
	os.Exit(exitCode)
}

func TestA(t *testing.T) {
	// 测试逻辑A...
	t.Log("I am A")
}

func TestB(t *testing.T) {
	// 测试逻辑B...
	t.Log("I am B")
}

// TestC 在包内的其他测试文件内
func TestC(t *testing.T) {
	// 测试逻辑C...
	t.Log("I am C")
}
```

```shell
➜  t go test -v .
I am Before
=== RUN   TestC
    hello_test.go:7: I am C
--- PASS: TestC (0.00s)
=== RUN   TestA
    main_test.go:28: I am A
--- PASS: TestA (0.00s)
=== RUN   TestB
    main_test.go:33: I am B
--- PASS: TestB (0.00s)
PASS
I am After
ok      alltest/t       0.418s
```

在上面的示例中，我们创建了一个名为 TestMain 的测试函数，并将其参数命名为 m，类型为 *testing.M。这个函数是特殊的，它允许我们在测试程序开始之前进行一些初始化操作或设置全局环境，并在所有测试运行完毕后进行清理操作。

在 TestMain 函数中，我们可以进行一些初始化工作，例如设置全局变量、建立数据库连接等。然后，我们调用 m.Run() 方法来运行测试。该方法会执行所有注册的测试函数，如果有测试失败，会返回相应的退出码。我们可以根据测试结果返回适当的退出码。

在示例中，我们定义了两个普通的测试函数 TestSomething 和 TestSomethingElse，它们会在 m.Run() 被调用时执行。

通过使用 testing.M 结构体和 TestMain 函数，我们可以在测试程序的开始和结束时执行一些操作，以及在所有测试运行完毕后对环境进行清理。这样，我们可以更好地控制和管理整个测试过程。(来自GPT)

## 基准测试
基准测试，与单元测试类似，首先也需要在`xxx_test.go`文件内。与单元测试不通，单元测试用例函数以`Test`开头定义，而基准测试用例函数，需要已`Benchmark`开头来定义。`go test`默认不运行基准测试, 运行基准测试，使用`-bench`指定需要运行的函数名，文件名，或当前包`.`。`TestMain`函数，仍然适用于基准测试。

需要说明，`go test`命令默认会执行单元测试用例，如果仅仅希望运行基准测试，跳过单元测试，可以使用`-run=^$`。完整命令如下`go test -v -bench BenchmarkFib -run=^a`

```golang
// 例如在上文的单元测试文件main_test.go内，新增一个基准测试用例BenchmarkFib。
func Fib(n int) int {
        if n < 2 {
                return n
        }
        return Fib(n-1) + Fib(n-2)
}

func BenchmarkFib(b *testing.B) {
	// run the Fib function b.N times
	for n := 0; n < b.N; n++ {
		Fib(10)
	}
}
```

```shell
# 运行当前包下的TestMain，指定TestMain只运行当前指定的基准测试，不运行单元测试。
➜  t go test -v -bench BenchmarkFib -run=^a       
I am Before
goos: darwin
goarch: arm64
pkg: alltest/t
BenchmarkFib
# BenchmarkFib-8 表明当前基准测试运用的系统CPU核心数GOMAXPROCS, 为8。调用了内部函数6395314次，平均每次调用花费187.1ns
BenchmarkFib-8           6395314               187.1 ns/op
PASS
I am After
ok      alltest/t       1.588s
```

1. 可以使用`-cpu`指定基准测试核心数。例如`-cpu=2,4`，可以给出2核心的基准测试情况，4核心的基准测试情况，用作对比。
2. 可以使用benchtime指定基准测试的时长，默认每个基准测试执行1s。例如，可以`-benchtime=5s`, 指定当前基准测试执行时间为5s。实际报告中，时长可能会大于5s。是调用目标测试函数外的开销。另外`-benchtime=50x`可以指定基准测试时长需要满足调用目标函数50次的时长。
3. 可以使用`-count`指定，执行基准测试的轮次数。每一轮次会提供一个报告条目，例如三个`BenchmarkFib-8`条目。

基准测试`B`接口内组合了上文介绍过的`common`结构， `common`接口含有的日志输出，错误中断等功能，在这里仍可以使用。

## 覆盖率测试
在`go test`命令中增加`cover`选项，覆盖率测试一般统计的是行数。

- `go test xxx -v -covermode=count`: `xxx`是包名，该命令会显示包下的所有测试的覆盖率，但是无法判断覆盖了哪些，没覆盖哪些。
- `go test xxx -v -coverprofile=count.out`: `xxx`是包名，该命令展示测试覆盖率,并生成覆盖统计文件到 `count.out`, `count.out`文件中详细展示了每个文件测试时某一行,执行的次数及其他信息(暂时只能用到次数)
- `go tool cover -func=count.out`: 用来分析第二点中生成的`count.out`文件。
	- `-func`: 生成每个函数的覆盖率
	- `-html`: 生成 html 文件,已图形形式展示每个函数,每一行代码的覆盖率
- `go tool cover -html=count.out`: 直接分析`count.out`文件，通过打开默认浏览器的方式。

```shell
➜  common git:(master) ✗ go test . -v -coverprofile=count.out
=== RUN   TestCheckPhoneNumber
--- PASS: TestCheckPhoneNumber (0.00s)
=== RUN   TestStringLength
--- PASS: TestStringLength (0.00s)
PASS
        github.com/dairongpeng/dev-test/common  coverage: 100.0% of statements
ok      github.com/dairongpeng/dev-test/common  0.638s  coverage: 100.0% of statements
➜  common git:(master) ✗ ls
common.go      common_test.go count.out
➜  common git:(master) ✗ go tool cover -func=count.out
github.com/dairongpeng/dev-test/common/common.go:9:     CheckPhoneNumber        100.0%
github.com/dairongpeng/dev-test/common/common.go:19:    StringLength            100.0%
total:                                                  (statements)            100.0%
```
## 测试断言
在`golang`官方的回答中，建议所有异常处理，都显示的判断且给出处理的策略。即`if err != nil`的形式。但是为了测试的方便，社区仍然提供了测试的各种异常情况的断言，省去一些代码量。常见的库是`github.com/stretchr/testify`

直接来自该库README的例子：

```golang
package yours

import (
  "testing"
  "github.com/stretchr/testify/assert"
)

func TestSomething(t *testing.T) {

  // assert equality
  assert.Equal(t, 123, 123, "they should be equal")

  // assert inequality
  assert.NotEqual(t, 123, 456, "they should not be equal")

  // assert for nil (good for errors)
  assert.Nil(t, object)

  // assert for not nil (good when you expect something)
  if assert.NotNil(t, object) {

    // now we know that object isn't nil, we are safe to make
    // further assertions without causing any errors
    assert.Equal(t, "Something", object.Value)

  }

}
```