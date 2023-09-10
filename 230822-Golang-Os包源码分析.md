`os`包是`golang`标准库中的一个核心包，用于操作操作系统功能，包括文件操作、环境变量、进程控制等。其屏蔽了平台的多样性，我们使用起来不会关心当前在何种操作系统平台上。`os`包内会存在比较多的系统调用，这些系统调用在`syscall`包内，做个了解。

## 文件和目录操作
### 获取文件的信息
```Go
// 获取文件信息, 包括文件的大小、权限、修改时间等。其中FileInfo是一个接口
// 如果指定的路径是一个符号链接，os.Stat 会返回链接所指向的文件的信息，而不是链接本身的信息。
func Stat(name string) (FileInfo, error)
// 获取文件信息, 包括文件的大小、权限、修改时间等。
// 它不会跟随符号链接，而是直接返回链接本身的信息。
func Lstat(name string) (FileInfo, error)

// 上面的FileInfo会返回文件的相关信息。
type FileInfo interface {
    Name() string       // 文件或目录的名称
    Size() int64        // 文件的大小（以字节为单位）
    Mode() FileMode     // 文件的模式和权限位
    ModTime() time.Time // 文件的修改时间
    IsDir() bool        // 是否是目录
    Sys() any           // 底层数据来源，依赖具体的文件系统
}
```
### 获取一个文件
```Go
// 只读模式打开文件, 文件不存在会报错。
func Open(name string) (*File, error)
// 创建文件。
func Create(name string) (*File, error)
// 用于打开文件以供读写操作，还可以指定打开文件的模式和权限。文件不存在会创建。
func OpenFile(name string, flag int, perm FileMode) (*File, error)
```
### 文件操作
```Go
// 读取文件，返回文件内容content，内容是[]byte类型
func ReadFile(name string) ([]byte, error)

// 写入内容data，到name文件中。
// 如果文件不存在，以perm权限创建该文件。
// 如果文件存在，则先清空该文件，再写入data
func WriteFile(name string, data []byte, perm FileMode) error

// 获取文件的状态FileInfo
func (f *File) Stat() (FileInfo, error)

// 关闭文件, 
func (f *File) Close() error

// 用于读取目录中的内容，并返回一个文件信息的切片 ([]FileInfo)，其中每个元素代表一个文件或子目录的信息。
// n参数表示要读取的最大条目数。如果n为负数，则读取所有条目。
func (f *File) Readdir(n int) ([]FileInfo, error)

// 类似于Readdir，但是它返回一个字符串切片([]string)，其中每个元素是文件或子目录的名称。
func (f *File) Readdirnames(n int) (names []string, err error)

// 推荐：ReadDir函数是Go1.16中新增的函数，用于读取目录中的内容，并返回一个DirEntry的切片 ([]DirEntry)，其中每个元素代表一个文件或子目录的信息。
// 如果n为负数，则读取所有条目
// DirEntry接口提供了文件名和文件信息等方法，允许你在不直接调用文件系统的情况下操作条目。
func (f *File) ReadDir(n int) ([]DirEntry, error)

// 用于修改文件的所有者 (uid) 和所属组 (gid)。
// 需要注意的是，使用这个函数需要有足够的权限，通常需要以管理员或超级用户身份运行程序。
func (f *File) Chown(uid, gid int) error

// 用于截断文件到指定大小。
// 如果文件的大小大于size，则文件会被截断为指定大小；如果文件的大小小于size，则文件内容不变。
// 这个函数常用于调整文件大小或清空文件内容。
func (f *File) Truncate(size int64) error

// 用于将文件的修改同步到存储介质上，确保文件内容的变化被持久化。
// 调用这个函数会将缓冲区中的数据刷新到磁盘上。
func (f *File) Sync() error

// 更改当前工作目录，到文件所在的目录。
func (f *File) Chdir() error

// 用于移动文件指针到指定位置。
// offset参数表示相对于whence的偏移量，whence参数指定基准位置，可以是io.SeekStart（文件起始位置）、io.SeekCurrent（当前位置）或io.SeekEnd（文件末尾位置）。
// 这个方法常用于读写文件时定位文件指针的位置。
func (f *File) Seek(offset int64, whence int) (ret int64, err error)

// 方法用于获取文件的底层文件描述符（file descriptor）。
// 文件描述符是一个非负整数，用于在操作系统层级标识打开的文件或资源。
// 可以使用文件描述符来执行底层的 I/O 操作，但这需要谨慎操作，避免破坏文件状态。
func (f *File) Fd() uintptr
```
### 读写文件
```Go
// 用于从文件中读取数据，将读取的内容存储在给定的字节切片b中。
// 方法会尽可能地读取b长度的数据，并返回实际读取的字节数n。
// 如果到达文件末尾，会返回io.EOF错误。
func (f *File) Read(b []byte) (n int, err error)

// 用于从文件的指定偏移量off处开始读取数据，将读取的内容存储在给定的字节切片b中。
// 方法会尽可能地读取b的长度的数据，并返回实际读取的字节数n。
// 注意，这里并不会改变文件指针的位置。
func (f *File) ReadAt(b []byte, off int64) (n int, err error)

// 用于从实现了io.Reader接口的对象r中读取数据，并将读取的内容写入文件。其实是一个写操作，从r中读，写入到f
// 尽可能地多次读取并写入，直到r返回io.EOF错误。
// 返回值n表示总共写入的字节数。
func (f *File) ReadFrom(r io.Reader) (n int64, err error)

// 用于向文件中写入字节切片b中的数据。
// 会将切片中的内容写入文件，并返回实际写入的字节数n。
// 如果写入时发生错误，会返回错误信息。
func (f *File) Write(b []byte) (n int, err error)

// 用于从文件的指定偏移量off处开始写入字节切片b中的数据。
// 会将切片中的内容写入文件，并返回实际写入的字节数n。
// // 如果写入时发生错误，会返回错误信息。
func (f *File) WriteAt(b []byte, off int64) (n int, err error)

// 用于向文件中写入字符串s中的数据。
// 会将字符串的内容写入文件，并返回实际写入的字节数n。
// 写入时发生错误，会返回错误信息。
func (f *File) WriteString(s string) (n int, err error)

// 用于设置文件的读写操作的超时截止时间。
// 参数t是一个 time.Time类型的值，表示超时截止时间。
// 如果在截止时间之前未完成读写操作，会返回超时错误。
func (f *File) SetDeadline(t time.Time) error

// 用于设置文件的读操作的超时截止时间。用法与SetDeadline类似
func (f *File) SetReadDeadline(t time.Time) error

// 用于设置文件的写操作的超时截止时间。用法与SetDeadline类似
func (f *File) SetWriteDeadline(t time.Time) error
```
### 目录操作
```Go
// 创建一个目录，perm用于指定目录的权限，例如0666
func Mkdir(name string, perm FileMode) error

// 切换当前工作目录workdir到dir目录
func Chdir(dir string) error

// Rename类似lunix的mv命令，使用newpath替代为oldpath，可以用来更改文件（目录）的路径，更改文件（目录）的名字。
func Rename(oldpath, newpath string) error

// 递归创建目录
func MkdirAll(path string, perm FileMode) error

// 递归删除目录
func RemoveAll(path string) error

// 获取当前的工作目录
func Getwd() (dir string, err error)

// 获取系统的默认临时目录的路径，临时目录通常用于存储临时文件，这些文件在程序运行结束后可能会被清理掉。
// linux及类linux中，是/tmp目录。
// windowns中，有可能是%TMP%, %TEMP%, %USERPROFILE%
func TempDir() string

// 返回当前用户的缓存目录的路径。缓存目录通常用于存储应用程序的缓存数据，这些数据可以在应用程序运行期间进行管理。
// linux及类Linux中，是$HOME/.cache
// windowns中，是%LocalAppData%
// macOs中，是$HOME/Library/Caches
func UserCacheDir() (string, error)

// 返回当前用户的配置目录的路径。配置目录通常用于存储应用程序的配置文件，用户可以在这里找到自定义的应用程序配置。
// linux及类Linux中，是$HOME/.config
// windowns中，是%AppData%
// macOs中，是$HOME/Library/Application Support
func UserConfigDir() (string, error)

// 返回当前用户的主目录的路径。主目录通常用于存储用户的个人文件和文件夹
// linux及类Linux中及macOs，是$HOME
// windowns中，是%USERPROFILE%
func UserHomeDir() (string, error)

// 类似linux中的chmod命令，用来修改文件或者目录的权限。
func Chmod(name string, mode FileMode) error

// 创建一个虚拟文件系统，以便在其中进行文件操作，而无需实际操作系统文件系统的支持。
// 创建一个虚拟文件系统，然后使用这个文件系统进行文件和目录的操作，就像在实际文件系统中一样
func DirFS(dir string) fs.FS
```
## 环境变量和进程信息
### 环境变量
```Go
// 设置环境变量key, value
func Setenv(key, value string) error

// 清除key为键的环境变量。设置变量的逆操作
func Unsetenv(key string) error

// 会清除当前进程的环境变量，包括系统自身的环境变量。它会将环境变量重置为空，并且只留下一些与程序运行有关的最基本的环境变量。
// 这个函数通常在需要一个干净的环境来运行程序时使用，以避免受到外部环境变量的影响。
// 在使用 os.Clearenv 之前，务必要确认是否需要清除环境变量，以及清除后是否会影响程序的预期功能。
func Clearenv()

// 通过key，查询环境变量的值信息。如果没设置，返回的是""
func Getenv(key string) string

// 通过key，查询环境变量的值信息。如果没设置，返回的是"", false
func LookupEnv(key string) (string, bool)

// 查询所有的环境变量信息，返回的元素格式形如"key=value"
func Environ() []string

// 用于展开字符串中的环境变量。
// 在字符串s中，可以使用 ${key} 或$key的形式来表示要被替换的环境变量。
// 函数会自动将环境变量的值替换到字符串中。
// eg: s=Hello, $VAR! 如果我们系统存在VAR这个环境变量的值为World，则最终返回"Hello World!"
func ExpandEnv(s string) string
```
### 进程及用户组信息
```Go
// 引用os包时，会把系统进程args赋值给Args, 直接使用Args即可拿到进程命令行参数。
// 注意参数的第一个元素，是进程名。
var Args []string

func init() {
    if runtime.GOOS == "windows" {
        // Initialized in exec_windows.go.
        return
    }
    Args = runtime_args()
}

// 用于获取调用进程的有效用户ID (UID)。
func Getuid() int

// 用于获取调用进程的有效用户ID (EUID)。
func Geteuid() int 

// 用于获取调用进程的有效组ID (GID)。
func Getgid() int

// 用于获取调用进程的有效组ID (EGID)。
func Getegid() int

// 返回主机名，例如mac上调用dairongpengdeMacBook-Pro.local
func Hostname() (name string, err error)

// 用于获取当前进程所属的所有附加组的组ID列表。
// 返回一个整数切片，包含了当前进程的所有附加组 ID。
// 如果获取失败，会返回一个错误。
func Getgroups() ([]int, error)

// 用于立即终止程序的执行并退出。
// 它接受一个退出状态码作为参数，通常用来指示程序在退出时的状态。
// 通常情况下，状态码为0表示程序正常退出，非零状态码表示程序在退出时遇到了错误或异常。
// 需要注意的是，在使用os.Exit之前，应该确保程序执行完所需的清理工作，
// 例如关闭文件、释放资源等。因为一旦调用了os.Exit，程序会立即终止，未执行的延迟函数（defer）也不会执行。
func Exit(code int)
```

### 操作系统进程信息
```Go
// 获取当前进程的PID
func Getpid() int

// 获取当前进程Pid的父亲进程Pid
func Getppid() int

// 通过进程pid，查找进程信息，返回go封装的Process结构。
func FindProcess(pid int) (*Process, error)

// 在新的进程中启动一个外部命令或程序。返回这个运行的进程信息
func StartProcess(name string, argv []string, attr *ProcAttr) (*Process, error)

// 用于释放与 Process 类型相关的资源。
// 在调用完其他方法并完成进程操作后，调用 Release 方法可以释放资源以防止资源泄漏。
func (p *Process) Release() error

// 用于向进程发送终止信号（通常是 SIGKILL）以终止进程的运行。
// 调用此方法会强制终止进程，不会等待进程正常退出。
func (p *Process) Kill() error

// 用于等待进程的退出并返回关于进程的状态信息。
// 此方法会阻塞当前 goroutine，直到进程退出为止。
func (p *Process) Wait() (*ProcessState, error)

// 用于向进程发送指定的操作系统信号。
// 参数sig是一个os.Signal类型，表示要发送的信号。
func (p *Process) Signal(sig Signal) error

// 用于获取进程在用户态（用户空间）运行的时间。
func (p *ProcessState) UserTime() time.Duration

// 用于获取进程在内核态（系统空间）运行的时间。
func (p *ProcessState) SystemTime() time.Duration

// 用于检查进程是否已经退出。
func (p *ProcessState) Exited() bool

// 用于检查进程是否成功退出。
func (p *ProcessState) Success() bool

// 返回与进程状态相关的系统特定信息。具体返回值的类型和内容与操作系统相关。
func (p *ProcessState) Sys() any

// 返回与进程资源使用情况相关的系统特定信息。具体返回值的类型和内容与操作系统相关。
func (p *ProcessState) SysUsage() any
```
## 用户和用户组
### 用户信息
```Go
type User struct {
    // 用户ID，eg: 501
    Uid string
    // 用户所在用户组ID，eg: 20
    Gid string
    // 用户登录名，eg: dairongpeng
    Username string
    // 用户名，eg: dairongpeng
    Name string
    // 用户家目录，eg: /Users/dairongpeng
    HomeDir string
}

// 获取当前用户信息
func Current() (*User, error)

// 指定用户名，获取用户
func Lookup(username string) (*User, error)

// 指定uid获取用户
func LookupId(uid string) (*User, error)
```
### 用户组信息
```Go
type Group struct {
    Gid  string // group ID
    Name string // group name
}

// 根据用户组名，获取用户组
func LookupGroup(name string) (*Group, error)

// 根据组ID, 获取用户组
func LookupGroupId(gid string) (*Group, error)

// 获取用户属于哪些用户组
func (u *User) GroupIds() ([]string, error)
```
## 操作系统信号
```Go
// 用于忽略一个或多个操作系统信号（os.Signal）
// 传递一个或多个信号作为参数，当程序接收到这些信号时，它们将被忽略，不会触发任何默认的处理机制。
// 例如，忽略syscall.SIGTERM信号，当按下Ctrl+C时，信号被忽略，当前程序也不会退出。
func Ignore(sig ...os.Signal)

// 用于检查指定的操作系统信号（os.Signal）是否被忽略
func Ignored(sig os.Signal) bool

// 用于向指定的信号通道（chan<- os.Signal）注册要接收的操作系统信号（os.Signal）
// 当sig内指定的信号被接收时，它将被发送到c通道。读取c通道的地方会接受到通知，需要注意，c是只读的。
func Notify(c chan<- os.Signal, sig ...os.Signal)

// 用于重置（恢复默认）一个或多个操作系统信号的处理方式。通常我们搭配Notify来使用，当c接受到信号后，我们重置Notify的信号后续为默认处理方式即可 。
func Reset(sig ...os.Signal)

// 用于停止接收信号。它接受一个单向通道参数，用于停止接收从该通道传入的信号。通常也是搭配Notify使用，当c接受到信号，处理完后，我们在设置停止从c中接受后续信号。
func Stop(c chan<- os.Signal)

// 用于在特定的操作系统信号到达时创建一个上下文（context）。
// parent是父上下文，用来衍生ctx。
// signals是用来监听的信号。当监听到这些信号后，实质上会调用stop，当然我们也可以在后续的代码中手动调用stop来取消上下文。
// stop, 用于手动停止上下文的传播和取消。
func NotifyContext(parent context.Context, signals ...os.Signal) (ctx context.Context, stop context.CancelFunc)
```
## 操作系统exec.Cmd
`exec`包提供了执行外部命令和进程的功能。它允许你启动并与外部命令进行交互，获取命令的输出，等待命令执行完成等操作。

`exec`包中最重要的结构是`exec.Cmd`，它表示一个正在准备执行的命令。通过创建`Cmd`结构的实例，你可以设置命令的参数、环境变量、工作目录以及其他相关属性。然后，你可以使用Cmd结构的方法来启动命令，并与其进行交互。

通过使用`exec`包和`exec.Cmd`结构，你可以在Go程序中方便地执行外部命令，并与其进行交互、处理输出结果和错误。这使得你可以在Go语言中与外部系统和工具进行集成，提供更强大的功能和灵活性。

```Go
type Cmd struct {
    // 要执行的外部命令的路径。
    // 可以是可执行文件的绝对路径，也可以是环境变量$PATH中可执行文件的名称。
    Path string

    // 命令的参数列表。
    // 第一个参数通常是命令本身，后面的参数是命令的选项和参数。
    Args []string

    // 命令执行时的环境变量
    // 是一个字符串切片，每个元素都是形如"key=value"的字符串。
    Env []string

    // 命令执行时的工作目录
    // 如果未指定，则使用当前工作目录。
    Dir string

    // 命令的标准输入
    Stdin io.Reader

    // 命令的标准输出
    Stdout io.Writer
    // 命令的标准错误输出
    Stderr io.Writer

    // 用于指定要传递给子进程的额外文件描述符。
    // 这些文件描述符可以在子进程中使用，以便与父进程共享文件句柄。
    // 添加到ExtraFiles切片中的文件句柄，传递给子进程，达到子进程和父进程共享文件资源的目的。
    ExtraFiles []*os.File

    // 于设置系统级进程属性，例如进程的信号处理、进程组ID等。
    // 通过创建syscall.SysProcAttr实例并设置其字段，可以控制子进程的行为。
    // 例如，你可以指定子进程的进程组ID，使其成为某个特定进程组的一部分。
    SysProcAttr *syscall.SysProcAttr

    // 表示由Cmd启动的子进程。它允许你与子进程进行交互，如向其发送信号、等待其完成等。
    // 在Cmd执行后，可以通过访问Process字段来获取与子进程相关的信息和控制
    // 例如，你可以使用Process的Signal方法向子进程发送信号，使用Wait方法等待子进程完成等。
    Process *os.Process

    // ProcessState字段包含有关已完成的进程的信息，例如进程的退出状态、运行时间等。
    // 当子进程完成后，可以通过访问ProcessState字段来获取与已完成的进程相关的信息
    ProcessState *os.ProcessState

    // 让cmd通过CommandContext，来携带ctx
    ctx context.Context

    // 通常是cmd中path不对应或找不到执行的命令的错误
    Err error

    // 当携带ctx后，ctx可能会cancel，这里提供cancel的回调函数，用来注册"收尾"函数。
    Cancel func() error
}
```

如何使用：
```Go
// 获取Cmd, name可以是命令的绝对路径，也可以是在path中的命令名。arg是命令需要的参数
func Command(name string, arg ...string) *Cmd

// 类似Command, 携带ctx。当ctx被取消时，会cmd.Process.Kill()来停掉cmd
func CommandContext(ctx context.Context, name string, arg ...string) *Cmd

// 获取命令Cmd的信息。包括path和args等。
func (c *Cmd) String() string

// 用于执行外部命令并等待其完成。会阻塞当前的 goroutine，直到命令执行完成后返回。
// 如果命令执行失败（退出状态非零），则会返回一个非空的错误，否则返回 nil。
func (c *Cmd) Run() error

// 用于启动外部命令，但不会等待命令的完成。
// 它会返回一个错误，如果启动命令失败则返回非空的错误，否则返回 nil。
// 使用 Start 后，你可以使用 Wait 方法来等待命令的完成。
func (c *Cmd) Start() error

// 调用Wait，等待Cmd完成。一般搭配Start使用，先Start再Wait
func (c *Cmd) Wait() error

// 用于执行外部命令并捕获其标准输出。
// 会阻塞当前的 goroutine，直到命令执行完成后返回。
// 字节切片表示命令的标准输出，如果命令执行失败（退出状态非零），则会返回一个非空的错误。
func (c *Cmd) Output() ([]byte, error)

// 用于执行外部命令并捕获其标准输出和标准错误输出。
// 会阻塞当前的 goroutine，直到命令执行完成后返回。
// 字节切片表示命令的标准输出和标准错误输出的合并，如果命令执行失败（退出状态非零），则会返回一个非空的错误。
func (c *Cmd) CombinedOutput() ([]byte, error)

// 用于创建一个用于向外部命令的标准输入写入数据的管道。
// 返回一个实现了 io.WriteCloser 接口的对象，使用这个io.WriteCloser，我们可以在cmd外部，往cmd进程标准输入中，写入数据。
func (c *Cmd) StdinPipe() (io.WriteCloser, error)

// 用于创建一个用于从外部命令的标准输出读取数据的管道。
// 实现了 io.ReadCloser 接口的对象，可以用于从命令的标准输出读取数据。
// 和StdinPipe相反，这里可以通过io.ReadCloser，读取到cmd的标准输出
func (c *Cmd) StdoutPipe() (io.ReadCloser, error)

// 用于创建一个用于从外部命令的标准错误输出读取数据的管道。
// 返回一个实现了 io.ReadCloser 接口的对象，可以用于从命令的标准错误输出读取数据。
// 与StdoutPip类似，差别是StderrPipe只读取标错误输出。
func (c *Cmd) StderrPipe() (io.ReadCloser, error)

// 返回当前cmd环境中，存在的环境变量信息。
func (c *Cmd) Environ() []string
```