`golang`标准库中提供了用于控制并发安全的库`sync`。这一篇，主要是来了解下这个库为我们开发提供了哪些便利，以及相关的源码简单分析。

首先我们知道，标准库提供的基础变量如`Int32`, `Int64`, `Float64`等，在对其递增递减等操作的时候是并发不安全的。在并发调用的过程中，有几率出现`竞态条件（Race Condition）`等问题。

## 并发问题分析

先简单看一下这个程序的并发问题：
```golang
func main() {
    var count int64

    for i := 0; i < 10000; i++ {
        go func() {
            // 并发问题1: 每个协程中对count执行加1操作，使用的是count=count+1， 这一个步骤不是并发安全的
            count += 1
        }()
    }

    // 并发问题2：我们并没有等待所有协程都执行完了count=count+1，再退出主协程，导致可能有些协程没有来得及执行自己的代码
    fmt.Println(count)
    
    // Output:
    // 9644
}
```

解决这个并发问题，第一是要对并发不安全的操作`count=count+1`进行并发控制，第二是要确保我们的`10000`个协程都执行了自己的代码。

```golang
func main() {
    var count int64 = 0
    var m sync.Mutex

    for i := 0; i < 10000; i++ {
        go func() {
            // 可以优化为atomic原子操作
            m.Lock()
            count += 1
            m.Unlock()
        }()
    }

    // 等待10000个协程执行完成；可优化为waitGroup
    time.Sleep(2 * time.Second)
    fmt.Println(count)
    
    // Output:
    // 10000
}
```

如果某个协程执行特别耗时，我们通过主协程休眠的方式，就可能会出问题。处理协程同步时，`sync`库中，有`waitGroup`, 这个是推崇的方式，经过优化：
```golang
func main() {
    var count int64 = 0
    var wg sync.WaitGroup

    wg.Add(10000)
    for i := 0; i < 10000; i++ {
        go func() {
            defer wg.Done()

            // 优化为atomic.AddInt64，这个操作是原子的。
            // 无论如何，我们保证我们对变量count的增加操作整个是并发安全的，都可以。但是count = count + 1这个操作非并发安全
            atomic.AddInt64(&count, 1)
        }()
    }

    wg.Wait()
    
    fmt.Println(count)

    // Output:
    // 10000
}
```

或者使用`atomic`提供的一些并发安全的类型，改造如下：
```golang
func main() {
    var count atomic.Int64
    var wg sync.WaitGroup

    wg.Add(10000)
    for i := 0; i < 10000; i++ {
        go func() {
            defer wg.Done()
            // 原子类型atomic.Int64的一些方法，是并发安全的
            count.Add(1)
        }()
    }

    wg.Wait()

    fmt.Println(count.Load())

    // Output:
    // 10000
}
```

## 原子类型
`golang`提供了一些原子类型以及方法，方便我们进行并发的修改某个类型的变量，在上文中我们已经使用过`atomic.Int64`。

### atomic.Value类型介绍
`atomic.Value`是`Golang`标准库中的一个结构体，它提供了一种原子性的方式来存储和加载任意类型的值。它在多个`goroutine`之间提供了安全的读写操作，可以用于共享数据而不需要显式的锁机制。`atomic.Value`适用于那些需要在不同的`goroutine`之间共享数据，并且对性能要求较高的情况。

其他原子类型，比如`atomic.Int64`,`atomic.Int32`都是同理的，这里只分析比较通用的结构`atomic.Value`。

```golang
type Value struct {
    v any
}

// Load方法用于原子地加载存储在atomic.Value中的值，并返回值的当前内容。这个操作是原子的，可以在没有锁的情况下进行，适用于读取操作。
func (v *Value) Load() (val any)

// Store方法用于原子地存储一个值到atomic.Value中，替换之前的值。这个操作是原子的，适用于写入操作。
func (v *Value) Store(val any)

// Swap方法用于原子地交换atomic.Value中的值，并返回之前的值。可以用于实现某种形式的乐观锁。
func (v *Value) Swap(new any) (old any)

// CompareAndSwap方法用于原子地比较并交换atomic.Value中的值。如果当前值与给定的旧值相等，就将新值存储到atomic.Value中。它可以用于实现CAS (Compare-And-Swap) 操作。
func (v *Value) CompareAndSwap(old, new any) (swapped bool)
```

`CAS`（Compare-And-Swap，比较并交换）是一种并发编程中的原子操作，用于实现多线程环境下的数据同步和同步。它是一种乐观锁的实现方式，通过比较内存中的值与预期值是否相等，如果相等则将新值写入，如果不相等则不进行写入，并且返回比较结果。`CAS`操作通常由硬件提供支持，可以在无锁的情况下实现线程安全。`CAS`有几个关键参数，其中，`destination`是要更新的内存位置，`oldValue`是预期的旧值，`newValue`是要写入的新值。`CompareAndSwap`操作会比较`destination`的当前值是否等于`oldValue`，如果相等，则将`newValue`写入`destination`，并返回true，表示写入成功；如果不相等，则不进行写入，返回 false。`CAS`的比较和交换两个步骤通过CPU的一个汇编指令执行，具有原子性，通常CAS可以看做是一种乐观锁的机制。

```golang
bool CompareAndSwap(T* destination, T oldValue, T newValue);
```

一些使用CAS的场景：
1. **实现无锁数据结构**：CAS可以用于实现无锁的数据结构，例如无锁队列、无锁哈希表等。在这些数据结构中，使用CAS可以避免锁的开销，提高并发性能。
2. **并发计数器**：CAS可以用于实现并发计数器，比如记录访问次数或计数统计等。多个线程可以在不阻塞的情况下对计数器进行递增操作。
3. **乐观锁**：CAS是乐观锁的基础，它可以用于在更新数据之前检查数据是否被其他线程修改过。如果没有被修改，则更新数据，否则进行回滚或重试。
4. **资源管理**：CAS可以用于管理共享资源的访问，例如线程池中的任务分配，资源的申请与释放等。
5. **并发队列**：CAS可以用于实现线程安全的并发队列，例如在生产者-消费者模型中。

```golang
func main() {
    var v atomic.Value

    var wg sync.WaitGroup
    for i := 0; i < 1000; i++ {
        index := i
        wg.Add(1)
        go func() {
            defer wg.Done()
            // 锁的入口
            if index == 0 {
                v.Store(0)
                return
            }

            // 自旋的场景
            b := v.CompareAndSwap(index-1, index)
            for !b {
                time.Sleep(3 * time.Millisecond)
                b = v.CompareAndSwap(index-1, index)
            }
        }()
    }

    wg.Wait()
    fmt.Println(v.Load())
    // Output:
    // 999
}
```

`CAS`存在`ABA`问题，例如我有其他协程把共享变量，从`A`修改为`B`再修改为`A`, 这个时候依赖`A`这个`oldValue`的协程都会得到机会执行，但不一定是当前得到执行线程严格依赖的`A`对应的`olcValue`，因为这个`A`是经过`B`状态修改回来的`A`。为了解决这个问题，可以对变量的值增加版本号，使用引用更新等技术。在使用版本号的方法里，每个协程标识自己需要依赖的`oldValue`是哪个版本的，即可。

顺便提一下，在`Value`的`CompareAndSwap`函数内部，可以看到调用了一个函数`CompareAndSwapPointer`，这个函数在`doc.go`中，但是是未实现的。

```golang
// CompareAndSwapPointer executes the compare-and-swap operation for a unsafe.Pointer value.
// Consider using the more ergonomic and less error-prone [Pointer.CompareAndSwap] instead.
func CompareAndSwapPointer(addr *unsafe.Pointer, old, new unsafe.Pointer) (swapped bool)
```

对于`CompareAndSwapPointer`这个函数，它是`Go`语言运行时库中的一个函数，用于执行原子的指针比较与交换操作。具体的实现在`Go`语言的运行时系统中（底层编译器或运行时库中进行实现），而不是在标准库的源码中。可以找到，在运行时库的源码中，`src/runtime/atomic_pointer.go`文件中包含了`CompareAndSwapPointer`等原子操作的实现代码。这些代码主要使用汇编语言来实现，因为原子操作通常需要底层硬件支持来保证操作的原子性。运行时库的源码是非常底层的，包含了与硬件相关的汇编代码，因此在正常的应用开发中，你不需要直接操作这些底层函数。标准库中提供的高级原子操作函数（如`sync/atomic`包中的函数）已经足够满足大多数需求，而且更易于使用和理解。

```golang
//go:linkname sync_atomic_CompareAndSwapPointer sync/atomic.CompareAndSwapPointer
//go:nosplit
func sync_atomic_CompareAndSwapPointer(ptr *unsafe.Pointer, old, new unsafe.Pointer) bool {
    if writeBarrier.enabled {
        atomicwb(ptr, new)
    }
    if goexperiment.CgoCheck2 {
        cgoCheckPtrWrite(ptr, new)
    }
    return sync_atomic_CompareAndSwapUintptr((*uintptr)(noescape(unsafe.Pointer(ptr))), uintptr(old), uintptr(new))
}
```

`go:linkname`是`Go`语言的一个编译器标记（Compiler Directive），用于在不同的包之间绕过`Go`语言的可见性规则，从而允许你在一个包中使用另一个包中私有（未导出）的标识符（变量、函数等）。正常情况下，`Go`语言的导出规则要求标识符只能在其定义的包内可见，其他包无法访问。但是在某些情况下，我们可能希望在不同的包中访问私有标识符，这时可以使用`go:linkname`来实现。如下，其中`localname`是当前包中的标识符名，importpath.name 是目标包中的标识符名。通过这种方式，你可以在当前包中使用`localname`来引用目标包中的`importpath.name`。

```golang
//go:linkname localname importpath.name
```

`go:linkname`的使用应该谨慎，并且只在必要的情况下使用，以避免代码的可维护性和可理解性下降。应用代码编写的过程中，要杜绝使用，这里只作为了解即可，以防标准库代码看不懂。

## Mutex与RWMutex
`Mutex`和`RWMutex`的相同点是都是增加临界区，保护临界区的并发安全性。不同点是，`Mutex`是互斥锁（排它锁），`RWMutex`是读写锁（共享锁）。

提示和建议：
- 初始`Mutex`的state是0，表示未加锁状态
- 不要拷贝锁变量
- `RWMutex`是封装了`Mutex`, 并且增加了一些原子变量实现的
- 对一个未加锁的`Mutex`, 解锁，会`panic`
- 可以使用`Trylock`函数，来让当前`goroutine`尝试获取锁，不会阻塞
- 多协程加锁时，未获取到锁的协程，会进入阻塞，底层进入到自旋状态。

`golang`中，当一个`goroutine`调用`Lock`操作时，如果锁是可用的（没有被其他`goroutine`占用），那么这个`goroutine`就会立即获取锁并继续执行。如果锁已经被其他`goroutine`占用，那么当前`goroutine`就会进入自旋状态(`spin`)，不断尝试获取锁。自旋是一种快速的等待方式，它适用于锁的占用时间短暂的情况。

如果自旋一段时间后仍未能获取到锁，当前`goroutine`就会放弃自旋，将自己休眠(`sleep`)一段时间，然后再次尝试获取锁。这种休眠的方式可以防止锁饥饿现象的发生，即避免某个`goroutine`永远无法获取到锁。

为了防止锁饥饿，`sync.Mutex`在实现中会使用一种随机的策略来控制休眠时间，从而避免多个`goroutine`在同一时间唤醒。这样可以减少竞争，使得等待的`goroutine`在一段时间后有机会获得锁。

虽然`sync.Mutex`使用了混合的策略来平衡自旋和休眠，但并不能完全避免锁饥饿。在某些极端情况下，仍然可能出现某个`goroutine`无法获取锁的情况。因此，在一些特殊的应用场景中，需要考虑使用其他类型的锁，如`sync.RWMutex`或使用信号量等机制，以进一步降低锁饥饿的风险。

```golang
// A Locker represents an object that can be locked and unlocked.
type Locker interface {
    Lock()
    Unlock()
}

// 互斥锁
type Mutex struct {
    state int32
    sema  uint32
}

// 读写锁
type RWMutex struct {
    w           Mutex        // held if there are pending writers
    writerSem   uint32       // semaphore for writers to wait for completing readers
    readerSem   uint32       // semaphore for readers to wait for completing writers
    readerCount atomic.Int32 // number of pending readers
    readerWait  atomic.Int32 // number of departing readers
}

// 加锁，增加一个临界区
func (m *Mutex) Lock()
// 尝试加锁
func (m *Mutex) TryLock() bool
// 解锁
func (m *Mutex) Unlock()


// 加读锁
func (rw *RWMutex) RLock()
// 尝试加读锁
func (rw *RWMutex) TryRLock() bool
// 释放读锁
func (rw *RWMutex) RUnlock()
// 加写锁
func (rw *RWMutex) Lock()
// 尝试加写锁
func (rw *RWMutex) TryLock() bool
// 释放写锁
func (rw *RWMutex) Unlock()
// 将当前读写锁，转化为Locker结构类型，接口类型对应的Lock和Unlock，对应当前读写锁的RLock和RUnlock
func (rw *RWMutex) RLocker() Locker
```

## WaitGroup
并发的产生常常是多个角色（`goroutine`），共同操作同一个临界区的代码导致，我们通过对临界区代码做同步，可以有效的解决临界区被`乱入`的问题。但是除了多个`goroutine`并发的控制，也需要合理编排`goroutine`使其可以按照我们的期望，进行运行和退出，这个时候我们可以使用`WaitGroup`。在golang中，`WaitGroup`是等待一组并发任务完成的同步原语，例如我们可以使用协程（`goroutine`）来实现并发任务，但在主线程结束前，我们需要确保所有协程都已经完成。这时就可以使用 sync.WaitGroup 来管理协程的执行，使主线程等待所有协程完成后再继续执行。。

`WaitGroup`包含三个方法：
```Go
// 往WaitGroup计数器添加增量
func (wg *WaitGroup) Add(delta int)

// 从WaitGroup计数器中减去一个增量
func (wg *WaitGroup) Done() 

// wg进行阻塞，等待计数器变为0为止
func (wg *WaitGroup) Wait()
```

一个简单的`WaitGroup`使用的案例：

```Go
func worker(id int, wg *sync.WaitGroup) {
    defer wg.Done() // 任务完成后通知WaitGroup减少一个计数

    fmt.Printf("Worker %d starting\n", id)
    time.Sleep(time.Second) // 模拟耗时操作
    fmt.Printf("Worker %d done\n", id)
}

func main() {
    var wg sync.WaitGroup

    for i := 1; i <= 3; i++ {
        wg.Add(1) // 增加等待的任务的数量
        go worker(i, &wg)
    }

    wg.Wait() // 阻塞，直到所有任务完成, 即计数器变为0

    fmt.Println("All workers are done")
}
```

## Once
`sync.Once`是标准库的一个并发原语，用来控制程序只需执行一次的操作。内部仅仅包含一个方法。

```Go
type Once struct {
    done uint32
    m    Mutex
}

// Do方法传入一个函数，并执行这个函数。
// 通过该once传入的函数，只会被执行一次
// 需要注意的是：是否会执行f的控制是在o结构内，即如果一个once执行了函数f1，下文再执行f2也是得不到执行的。
func (o *Once) Do(f func())
```

查看`Do`方法的细节，可以看到，如果一个`once`执行过函数，其内部的`done`会被标记，在程序的其他地方再调用该`once`来执行相同函数，或者其他函数，都会得不到执行。
```Go
// Do方法调用doSlow方法来执行f
func (o *Once) doSlow(f func()) {
    o.m.Lock()
    defer o.m.Unlock()
    // 如果该once已经执行过f，则状态为1，后传入的函数f都得不到执行。
    if o.done == 0 {
        defer atomic.StoreUint32(&o.done, 1)
        f()
    }
}
```

## sync.Map
`sync.Map`是标准库提供的并发安全的`Map`，相较于标准库`Map`一个注重并发安全，一个注重读写性能。

```Go
// 读取Key
func (m *Map) Load(key any) (value any, ok bool)

// 写入key, value
func (m *Map) Store(key, value any)

// 用于尝试加载指定的键key对应的值，如果键不存在则存储这个key的值为value
// actual: 这是存储在键key下的实际值。如果之前已经有一个值与指定的键关联，那么actual将返回已经存在的值；否则，它将返回你提供的新值value。
// loaded: 表示是否加载了一个已存在的值。如果loaded为true，则说明之前已经存在一个与指定键关联的值；如果为false，则说明之前没有与指定键关联的值，是新存储的。
func (m *Map) LoadOrStore(key, value any) (actual any, loaded bool)

// 原子性地加载并删除指定键（key）对应的值。
// key: 要加载和删除的key
// value: 加载的值。如果之前有一个与指定键关联的value, 将返回这个已存在的值；如果该key没有value，返回nil。
// loaded: 表示是否加载了一个已存在的值。如果loaded为true，则说明之前已经存在一个与指定键关联的值；如果为false，则说明之前没有与指定键关联的值。
func (m *Map) LoadAndDelete(key any) (value any, loaded bool)

// 删除给定的key对应的key-value
func (m *Map) Delete(key any)

// Swap用于原子性地交换给定键的值。
// 1. key：需要交换value对应的键
// 2. value: 要替换的新值
// 3. previous: 交换前的旧值。如果之前已经有一个与指定键关联的值，previous将返回这个已存在的值；如果没有previous将返回nil
// 4. loaded表示key是否在map中存在。
func (m *Map) Swap(key, value any) (previous any, loaded bool)

// 比较交换key对应的value
// 1. key: 要比较value的key
// 2. old: 要比较的旧值
// 3. new: 要替换的新值
// 返回是否成功交换值。具体逻辑为：如果旧值old与指定key的value匹配，表示交换成功结果为true；如果不匹配，结果为false
func (m *Map) CompareAndSwap(key, old, new any) bool

// 比较删除给定的key。
// 如果key对应的value等于给定的old，则把该Key删掉，否则不删。是否删除返回deleted标识。
func (m *Map) CompareAndDelete(key, old any) (deleted bool)

// Range用于并发安全的遍历所有键值对, 并对每个键值对执行指定的f函数。Range方法在遍历时，内部会锁住sync.Map，以确保在遍历过程中不会出现并发问题。
// 1. f函数的key，value会在map中每一组key-value键值对上调用。
// 2. 如果希望继续遍历，f返回true。
// 3. 如果希望终止遍历，f返回false。
func (m *Map) Range(f func(key, value any) bool)
```

## Pool
`sync.Pool`是一个用于管理临时`对象池`的类型，用于提高性能并减少内存分配和垃圾回收的开销。`sync.Pool`的主要目的是在需要时重用临时对象，而不是频繁地创建和销毁它们, `Pool`可以被多个`goroutine`同时安全使用。主要适用场景为：
1. **短暂的临时对象**：如果你需要在短时间内创建和使用一些临时对象，然后丢弃它们，`sync.Pool`可以帮助你在这些对象之间进行重用，从而避免不必要的内存分配。
2. **频繁的小内存分配**：频繁地分配小的内存块会增加垃圾回收的压力。通过使用`sync.Pool`，可以减少内存分配的频率，从而减少垃圾回收的开销。
3. **性能优化**：在一些性能关键的应用场景中，避免过多的内存分配和垃圾回收是提高性能的一个重要策略。`sync.Pool`可以帮助你在一定程度上优化程序的性能。

需要注意的是，`sync.Pool`并不保证池中的对象一定会被重用，也不保证对象在池中的存在时间。系统会根据实际情况来决定是否从池中获取对象或创建新的对象。因此，在使用`sync.Pool`时，不应该假设池中的对象会一直存在。它主要用于优化短暂的临时对象的创建和销毁过程, 其次与其他并发包原语类似，`Pool`在首次使用后不能被复制。

```Go
type Pool struct {
    // ... 省略 ...

    // New可选地指定一个函数来生成一个值，否则Get将返回nil。它不能与调用Get同时更改。
    New func() any
}

// Get从池中获取一个对象，相应地，池中移出获得到的该对象。
// 1. 调用者不应该假定传递给Put的值与Get返回的值之间存在任何关系。
// 2. 如果Get函数返回nil并且p.New为非nil，则Get函数返回调用p.New获得到的对象结果。
func (p *Pool) Get() any

// 往对象池中放入一个对象。
func (p *Pool) Put(x any)
```

下面通过标准库`fmt`包中对于`sync.Pool`的一处使用, `buffer`池，来看一下如何使用`sync.Pool`：
```Go
// 定义一个并发安全的对象池, buffer池，用来生成buffer对象
var fmtBufferPool = sync.Pool{

    // 某人构建对象函数New, 当Get从对象池中获取对象，但对象池中没有盈余的对象时，使用New函数生成一个
    New: func() interface{} {
        return new(bytes.Buffer)
    },
}

// fmt.go 270L
func tconv(t *Type, verb rune, mode fmtMode) string {
    // 从buffer对象池中获取一个Buffer对象，响应的Buffer池中弹出这个对象，如果池中没有可用对象，使用New函数生成一个
    buf := fmtBufferPool.Get().(*bytes.Buffer)
    // 由于是复用的，可能slice存在内容残留，这里把获取到的对象的slice清空
    buf.Reset()
    // 在使用完这个对象后，把这个对象放回到Buffer池中。
    defer fmtBufferPool.Put(buf)

    tconv2(buf, t, verb, mode, nil)
    return InternString(buf.Bytes())
}
```

## Cond
`sync.Cond`是标准库中提供的条件变量`Condition Variable`类型，用于在多个协程之间进行等待和通信。条件变量在需要在某个条件满足时通知或唤醒等待的协程时非常有用。条件变量通常适合以下场景：
1. **协程间的同步和通信**：当你需要协调多个协程的执行顺序，让某个协程等待某个条件达成时，可以使用`sync.Cond`。
2. **生产者-消费者问题**：当生产者和消费者之间需要同步，确保生产者在队列非满时添加数据，消费者在队列非空时取出数据，可以使用条件变量来实现。
3. **任务调度**：在某些情况下，你可能希望让某些协程等待一些特定的事件或条件，然后再继续执行，这时可以使用条件变量。

```Go
// 创建一个新的条件变量，需要传入一个实现了sync.Locker接口的互斥锁作为参数, 通常为sync.Mutex和sync.RWMutex
// 这个互斥锁将与条件变量一起使用来保护条件的等待和通知操作。
func NewCond(l Locker) *Cond

// 进入等待状态，等待条件的通知。在调用这个方法前，协程应该持有互斥锁。Wait方法会自动释放互斥锁，并在条件满足时等待通知。
// 当接收到通知时，会重新获取互斥锁并继续执行。
func (c *Cond) Wait()

// 发送一个通知给一个等待的协程。通知将会唤醒等待队列中的一个协程，如果有等待的协程存在的话。在调用Signal方法前，应该已经获取了互斥锁。
// 唤醒一个等待中的goroutine
func (c *Cond) Signal()

// 发送一个通知给所有等待的协程。与Signal不同，Broadcast会唤醒等待队列中的所有协程。同样，在调用Broadcast方法前，应该已经获取了互斥锁。
// 唤醒所有等待goroutine。
func (c *Cond) Broadcast()
```

`sync.Cond`是标准库中最容易出错的并发结构，一般不推荐使用。这里是什么场景需要考虑使用`sync.Cond`的[讨论](https://stackoverflow.com/questions/36857167/how-to-correctly-use-sync-cond)。
1. 如果每次写入和读取都有一个`goroutine`，建议使用`sync.Mutex`而不是`sync.Cond`。
2. 在多个消费者等待共享资源可用的情况下，sync.Cond 可能很有用
3. 即时如此，如果情况允许，使用通道代替`sync.Cond`仍然是推荐的传递数据的方式。

```Go
// 多消费者共享资源
var sharedRsc = make(map[string]interface{})

func main() {
    var wg sync.WaitGroup
    wg.Add(2)
    m := sync.Mutex{}
    c := sync.NewCond(&m)
    
    go func() {
        defer wg.Done()

        c.L.Lock()
        // 如果资源没有准备就绪，就等待
        for len(sharedRsc) == 0 {
            // wiat会先Unlock, 再阻塞等待通知信号，获取到通知信号后，再Lock。
            c.Wait()
        }
        fmt.Println("goroutine1", sharedRsc["rsc1"])
        c.L.Unlock()
    }()

    go func() {
        defer wg.Done()

        c.L.Lock()
        // 如果资源没准备就绪，就等待
        for len(sharedRsc) == 0 {
            // wiat会先Unlock, 再阻塞等待通知信号，获取到通知信号后，再Lock。
            c.Wait()
        }
        fmt.Println("goroutine2", sharedRsc["rsc2"])
        c.L.Unlock()
    }()

    c.L.Lock()
    // 准备共享资源数据
    sharedRsc["rsc1"] = "foo"
    sharedRsc["rsc2"] = "bar"
    // 共享资源准备就绪，打开栅栏，让所有消费者去获取资源
    c.Broadcast()
    c.L.Unlock()
    
    // 仅仅用来等待两个消费者协程消费完毕
    wg.Wait()
}
```

注意：如果确实需要选用`sync.Cond`，牢记两个点，以免出现并发死锁问题：
1. Wait方法的调用需要加锁。
2. 被通知唤醒后的协程，一定要检查条件是否真的已经满足。