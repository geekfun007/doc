# Go GMP 调度模型详解

本文档详细介绍 Go 语言的 GMP 调度模型，包括 Goroutine、M（Machine）、P（Processor）的工作原理和调度机制。

## 目录

1. [GMP 模型概述](#gmp-模型概述)
2. [G - Goroutine](#g---goroutine)
3. [M - Machine](#m---machine)
4. [P - Processor](#p---processor)
5. [调度流程](#调度流程)
6. [调度策略](#调度策略)
7. [抢占机制](#抢占机制)
8. [网络轮询器](#网络轮询器)
9. [调度器源码分析](#调度器源码分析)
10. [性能调优](#性能调优)

---

## GMP 模型概述

### 什么是 GMP？

GMP 是 Go 运行时调度器的核心模型，由三个主要组件组成：

- **G (Goroutine)**：Go 协程，轻量级线程
- **M (Machine)**：操作系统线程，真正执行计算的资源
- **P (Processor)**：逻辑处理器，调度上下文

### 模型架构图

```
                    ┌─────────────────────────────────────────────────┐
                    │                  Go Runtime                      │
                    └─────────────────────────────────────────────────┘
                                          │
          ┌───────────────────────────────┼───────────────────────────────┐
          │                               │                               │
          ▼                               ▼                               ▼
    ┌──────────┐                    ┌──────────┐                    ┌──────────┐
    │    P0    │                    │    P1    │                    │    P2    │
    │ ┌──────┐ │                    │ ┌──────┐ │                    │ ┌──────┐ │
    │ │ LRQ  │ │                    │ │ LRQ  │ │                    │ │ LRQ  │ │
    │ │G G G │ │                    │ │G G G │ │                    │ │G G G │ │
    │ └──────┘ │                    │ └──────┘ │                    │ └──────┘ │
    └────┬─────┘                    └────┬─────┘                    └────┬─────┘
         │                               │                               │
         ▼                               ▼                               ▼
    ┌──────────┐                    ┌──────────┐                    ┌──────────┐
    │    M0    │                    │    M1    │                    │    M2    │
    │ (Thread) │                    │ (Thread) │                    │ (Thread) │
    └────┬─────┘                    └────┬─────┘                    └────┬─────┘
         │                               │                               │
         └───────────────────────────────┼───────────────────────────────┘
                                         │
                    ┌────────────────────┼────────────────────┐
                    │                    │                    │
                    ▼                    ▼                    ▼
              ┌──────────┐        ┌──────────────┐      ┌──────────┐
              │   CPU    │        │   Global     │      │  Network │
              │  Core 0  │        │ Run Queue    │      │  Poller  │
              │          │        │   (GRQ)      │      │          │
              └──────────┘        │  G G G G G   │      │  G G G   │
                                  └──────────────┘      └──────────┘
```

### 核心关系

```
┌─────────────────────────────────────────────────────────────────────────┐
│                           关系说明                                       │
├─────────────────────────────────────────────────────────────────────────┤
│  G : M : P  =  N : M : P                                                │
│                                                                         │
│  • G 的数量：可以有成千上万个                                             │
│  • M 的数量：默认最大 10000 个（可调整）                                   │
│  • P 的数量：默认等于 CPU 核心数（GOMAXPROCS）                             │
│                                                                         │
│  • 一个 G 需要绑定一个 P 才能被 M 执行                                    │
│  • 一个 P 同时只能绑定一个 M                                              │
│  • 一个 M 同时只能执行一个 G                                              │
│  • M 的数量 >= P 的数量（有些 M 可能在阻塞中）                             │
└─────────────────────────────────────────────────────────────────────────┘
```

### 演进历史

| 版本 | 调度模型 | 特点 |
|------|---------|------|
| Go 1.0 | GM | 单一全局队列，锁竞争严重 |
| Go 1.1 | GMP | 引入 P，本地队列，大幅提升性能 |
| Go 1.2 | GMP + 抢占 | 基于协作的抢占 |
| Go 1.14 | GMP + 信号抢占 | 基于信号的异步抢占 |

---

## G - Goroutine

### G 的定义

```go
// runtime/runtime2.go
type g struct {
    // 栈相关
    stack       stack   // 栈内存范围 [stack.lo, stack.hi)
    stackguard0 uintptr // 栈溢出检查
    stackguard1 uintptr // 被 C 代码使用

    // 当前 G 的状态
    _panic    *_panic // 最内层的 panic
    _defer    *_defer // 最内层的 defer
    
    // 绑定的 M
    m         *m      // 当前关联的 M
    
    // 调度相关
    sched     gobuf   // 调度信息（保存上下文）
    
    // 状态
    atomicstatus atomic.Uint32 // G 的状态
    goid         uint64        // G 的唯一 ID
    
    // 抢占相关
    preempt       bool // 抢占标记
    preemptStop   bool // 抢占时是否要停止
    preemptShrink bool // 抢占时是否要收缩栈
    
    // 其他字段...
}

// 保存 G 的调度上下文
type gobuf struct {
    sp   uintptr // 栈指针
    pc   uintptr // 程序计数器
    g    guintptr // 关联的 G
    ctxt unsafe.Pointer
    ret  uintptr
    lr   uintptr
    bp   uintptr // 基址指针
}
```

### G 的状态

```go
// G 的状态常量
const (
    _Gidle = iota // 0 - 刚分配，未初始化
    _Grunnable    // 1 - 在运行队列中，等待被调度
    _Grunning     // 2 - 正在 M 上运行
    _Gsyscall     // 3 - 正在执行系统调用
    _Gwaiting     // 4 - 被阻塞，不在运行队列
    _Gdead        // 6 - 已退出，可被复用
    _Gcopystack   // 8 - 栈正在被复制
    _Gpreempted   // 9 - 被抢占，等待恢复
)
```

### G 的状态转换图

```
                                    ┌──────────────┐
                                    │   _Gidle     │
                                    │  (刚分配)     │
                                    └──────┬───────┘
                                           │ newproc
                                           ▼
    ┌───────────────────────────────────────────────────────────────┐
    │                                                               │
    │  ┌─────────────┐    schedule    ┌─────────────┐              │
    │  │ _Grunnable  │ ─────────────► │  _Grunning  │              │
    │  │ (可运行)     │ ◄───────────── │  (运行中)    │              │
    │  └─────────────┘    preempt/    └──────┬──────┘              │
    │        ▲            yield              │                      │
    │        │                               │                      │
    │        │                    ┌──────────┴──────────┐          │
    │        │                    │                     │          │
    │        │              syscall                 chan/          │
    │        │                    │               lock/wait        │
    │        │                    ▼                     ▼          │
    │        │           ┌─────────────┐        ┌─────────────┐    │
    │        │           │  _Gsyscall  │        │  _Gwaiting  │    │
    │        │           │ (系统调用)   │        │  (等待中)    │    │
    │        │           └──────┬──────┘        └──────┬──────┘    │
    │        │                  │                      │           │
    │        │            exitsyscall            ready/notify      │
    │        │                  │                      │           │
    │        └──────────────────┴──────────────────────┘           │
    │                                                               │
    └───────────────────────────────────────────────────────────────┘
                                           │
                                           │ goexit
                                           ▼
                                    ┌──────────────┐
                                    │   _Gdead     │
                                    │  (已退出)     │
                                    └──────────────┘
```

### G 的创建过程

```go
// 创建 Goroutine
go func() {
    // 函数体
}()

// 编译器转换为
// runtime.newproc(funcval)
```

```go
// runtime/proc.go
func newproc(fn *funcval) {
    // 获取调用者的 PC（用于追踪）
    pc := getcallerpc()
    
    // 切换到 g0 栈执行
    systemstack(func() {
        newg := newproc1(fn, pc)
        
        // 获取当前 P
        pp := getg().m.p.ptr()
        
        // 将新 G 放入 P 的本地队列
        runqput(pp, newg, true)
        
        // 如果有空闲的 P 和 M，唤醒它们
        if mainStarted {
            wakep()
        }
    })
}

func newproc1(fn *funcval, callerpc uintptr) *g {
    // 获取当前 G
    gp := getg()
    
    // 尝试从 P 的空闲 G 列表获取
    pp := gp.m.p.ptr()
    newg := gfget(pp)
    
    if newg == nil {
        // 创建新的 G，初始栈大小 2KB
        newg = malg(stackMin)
        casgstatus(newg, _Gidle, _Gdead)
        allgadd(newg) // 加入全局 G 列表
    }
    
    // 初始化 G 的栈和调度信息
    // ...
    
    // 设置状态为可运行
    casgstatus(newg, _Gdead, _Grunnable)
    
    return newg
}
```

### G 的栈管理

```go
// 初始栈大小
const (
    stackMin = 2048  // 2KB 最小栈
    stackMax = 1 << 30 // 1GB 最大栈（64位系统）
)

// 栈增长
// 当栈空间不足时，会触发栈增长（复制到更大的空间）
func morestack() {
    // 分配更大的栈（通常是 2 倍）
    // 复制旧栈内容到新栈
    // 更新所有栈上的指针
}

// 栈收缩
// GC 时可能会收缩使用率低的栈
func shrinkstack(gp *g) {
    // 如果栈使用率 < 25%，收缩为 1/2
}
```

---

## M - Machine

### M 的定义

```go
// runtime/runtime2.go
type m struct {
    // G0：每个 M 都有的调度协程
    g0      *g     // 调度栈
    
    // 当前运行的 G
    curg    *g     // 当前 G
    
    // 关联的 P
    p       puintptr // 当前关联的 P
    nextp   puintptr // 暂存的 P（唤醒时使用）
    oldp    puintptr // 系统调用之前的 P
    
    // 线程相关
    id      int64
    spinning bool // 是否在自旋寻找 G
    blocked  bool // 是否被阻塞
    
    // 调度相关
    schedlink muintptr // 空闲 M 链表
    
    // 锁相关
    locks     int32
    
    // 信号相关
    gsignal   *g // 信号处理 G
    
    // 其他字段...
}
```

### M 的状态

```
┌─────────────────────────────────────────────────────────────────────┐
│                          M 的状态                                    │
├─────────────────────────────────────────────────────────────────────┤
│                                                                     │
│  ┌──────────┐                                                       │
│  │  创建    │                                                       │
│  └────┬─────┘                                                       │
│       │                                                             │
│       ▼                                                             │
│  ┌──────────┐      获取 P      ┌──────────┐                        │
│  │  自旋    │ ───────────────► │  运行    │                        │
│  │ spinning │ ◄─────────────── │ running  │                        │
│  └────┬─────┘    释放 P/      └────┬─────┘                        │
│       │          无 G              │                               │
│       │                            │ 系统调用                       │
│       │                            ▼                               │
│       │                       ┌──────────┐                         │
│       │                       │  阻塞    │                         │
│       │                       │ blocked  │                         │
│       │                       └────┬─────┘                         │
│       │                            │                               │
│       │              长时间空闲     │ 返回                          │
│       │                            │                               │
│       ▼                            ▼                               │
│  ┌──────────┐                ┌──────────┐                         │
│  │  休眠    │                │ 被抢占P  │                         │
│  │ sleeping │                │(syscall) │                         │
│  └──────────┘                └──────────┘                         │
│                                                                     │
└─────────────────────────────────────────────────────────────────────┘
```

### M0 和 G0

```go
// M0：主线程对应的 M
// - 程序启动时创建
// - 负责初始化和启动第一个 G（main goroutine）

// G0：每个 M 都有的调度 G
// - 不执行用户代码
// - 负责调度其他 G
// - 执行栈增长、GC 等运行时任务
// - 使用系统栈（比普通 G 的栈大得多）

func schedule() {
    // 在 G0 上执行
    gp := getg() // 获取当前 G（此时是 G0）
    
    // 寻找可运行的 G
    // ...
    
    // 切换到用户 G 执行
    execute(gp, inheritTime)
}
```

### M 的创建

```go
// 创建新的 M
func newm(fn func(), pp *p, id int64) {
    // 分配 M 结构
    mp := allocm(pp, fn, id)
    
    // 设置要关联的 P
    mp.nextp.set(pp)
    
    // 创建系统线程
    newm1(mp)
}

func newm1(mp *m) {
    // 创建操作系统线程
    newosproc(mp)
}

// 不同系统的线程创建
// Linux: clone
// macOS: pthread_create
// Windows: CreateThread
```

---

## P - Processor

### P 的定义

```go
// runtime/runtime2.go
type p struct {
    id          int32
    status      uint32 // P 的状态
    
    // 关联的 M
    m           muintptr // 当前关联的 M
    
    // 本地运行队列
    runqhead uint32        // 队列头
    runqtail uint32        // 队列尾
    runq     [256]guintptr // 环形队列，固定 256 大小
    
    // 下一个要运行的 G（优先级最高）
    runnext guintptr
    
    // 空闲 G 列表（用于复用）
    gFree struct {
        gList
        n int32
    }
    
    // 内存分配缓存
    mcache *mcache
    
    // defer 池
    deferpool []*_defer
    
    // 调度统计
    schedtick   uint32 // 调度次数
    syscalltick uint32 // 系统调用次数
    
    // GC 相关
    gcBgMarkWorker guintptr // GC 后台标记协程
    
    // 其他字段...
}
```

### P 的状态

```go
const (
    _Pidle    = iota // 空闲，在空闲列表中
    _Prunning        // 正在被 M 使用
    _Psyscall        // M 正在系统调用中
    _Pgcstop         // 被 GC 停止
    _Pdead           // 不再使用（GOMAXPROCS 减少时）
)
```

### P 的状态转换图

```
                         ┌───────────────┐
                         │    _Pidle     │
                         │   (空闲)      │◄────────────────┐
                         └───────┬───────┘                 │
                                 │                         │
                          acquirep                   releasep
                                 │                         │
                                 ▼                         │
                         ┌───────────────┐                 │
          ┌─────────────►│  _Prunning    │─────────────────┤
          │              │   (运行)      │                 │
          │              └───────┬───────┘                 │
          │                      │                         │
          │               entersyscall                     │
          │                      │                         │
          │                      ▼                         │
          │              ┌───────────────┐                 │
          │              │  _Psyscall    │                 │
     exitsyscall         │ (系统调用)    │                 │
          │              └───────┬───────┘                 │
          │                      │                         │
          │           ┌──────────┴──────────┐             │
          │           │                     │             │
          │     exitsyscall            handoffp           │
          │     (快速路径)            (被抢占)            │
          │           │                     │             │
          └───────────┘                     └─────────────┘
                         
                         ┌───────────────┐
                         │   _Pgcstop    │
                         │  (GC 停止)    │
                         └───────────────┘
                                 ▲
                                 │
                              stopTheWorld
```

### P 的本地队列

```go
// 本地队列操作

// 放入本地队列
func runqput(pp *p, gp *g, next bool) {
    if next {
        // 放入 runnext（最高优先级）
        oldnext := pp.runnext
        pp.runnext.set(gp)
        if oldnext != 0 {
            // 原来的 runnext 放入队列尾部
            runqputslow(pp, oldnext.ptr(), h, t)
        }
        return
    }
    
    // 放入队列尾部
    h := atomic.LoadAcq(&pp.runqhead)
    t := pp.runqtail
    if t-h < uint32(len(pp.runq)) {
        pp.runq[t%uint32(len(pp.runq))].set(gp)
        atomic.StoreRel(&pp.runqtail, t+1)
        return
    }
    
    // 本地队列满了，放入全局队列
    runqputslow(pp, gp, h, t)
}

// 从本地队列获取
func runqget(pp *p) (gp *g, inheritTime bool) {
    // 先检查 runnext
    next := pp.runnext
    if next != 0 && pp.runnext.cas(next, 0) {
        return next.ptr(), true
    }
    
    // 从队列头获取
    h := atomic.LoadAcq(&pp.runqhead)
    t := pp.runqtail
    if t == h {
        return nil, false
    }
    gp = pp.runq[h%uint32(len(pp.runq))].ptr()
    atomic.StoreRel(&pp.runqhead, h+1)
    return gp, false
}
```

### GOMAXPROCS

```go
// 设置 P 的数量
runtime.GOMAXPROCS(n)

// 默认值
// Go 1.5+：等于 CPU 核心数
// Go 1.5 之前：1

// 最佳实践
// CPU 密集型：GOMAXPROCS = CPU 核心数
// IO 密集型：GOMAXPROCS 可以适当增大
```

---

## 调度流程

### 完整调度循环

```go
// runtime/proc.go

// 调度循环入口
func schedule() {
    mp := getg().m
    
    // 检查 GC 等待
    if gp.m.p.ptr().gcBgMarkWorker != 0 {
        // 执行 GC 标记工作
    }
    
    var gp *g
    var inheritTime bool
    
    // 1. 为了公平，每 61 次调度从全局队列获取
    if gp == nil {
        if mp.p.ptr().schedtick%61 == 0 && sched.runqsize > 0 {
            gp = globrunqget(mp.p.ptr(), 1)
        }
    }
    
    // 2. 从本地队列获取
    if gp == nil {
        gp, inheritTime = runqget(mp.p.ptr())
    }
    
    // 3. 调用 findrunnable 获取（会阻塞）
    if gp == nil {
        gp, inheritTime = findrunnable()
    }
    
    // 执行 G
    execute(gp, inheritTime)
}

// 查找可运行的 G
func findrunnable() (gp *g, inheritTime bool) {
    mp := getg().m
    pp := mp.p.ptr()
    
top:
    // 1. 检查本地队列
    if gp, inheritTime := runqget(pp); gp != nil {
        return gp, inheritTime
    }
    
    // 2. 检查全局队列
    if sched.runqsize != 0 {
        gp := globrunqget(pp, 0)
        if gp != nil {
            return gp, false
        }
    }
    
    // 3. 检查网络轮询器
    if netpollinited() && netpollWaiters.Load() > 0 {
        if list := netpoll(0); !list.empty() {
            gp := list.pop()
            injectglist(&list) // 其他放入全局队列
            return gp, false
        }
    }
    
    // 4. 工作窃取：从其他 P 偷取
    if gp := runqsteal(pp, stealOrder.start(fastrandn(uint32(gomaxprocs)))); gp != nil {
        return gp, false
    }
    
    // 5. 再次检查全局队列
    if sched.runqsize != 0 {
        gp := globrunqget(pp, 0)
        if gp != nil {
            return gp, false
        }
    }
    
    // 6. 再次检查网络轮询器（阻塞）
    if netpollinited() && netpollWaiters.Load() > 0 {
        // 阻塞等待网络事件
    }
    
    // 7. 没有工作，停止当前 M
    stopm()
    goto top
}
```

### 调度时机

```go
// 1. go 关键字：创建新 G
go func() {}()

// 2. GC：STW 和并发标记
runtime.GC()

// 3. 系统调用
syscall.Read(fd, buf)

// 4. 同步原语
<-ch        // channel 接收
ch <- v     // channel 发送
lock.Lock() // 互斥锁

// 5. time.Sleep
time.Sleep(time.Second)

// 6. runtime.Gosched：主动让出
runtime.Gosched()

// 7. 栈增长
// 函数调用时检查栈空间

// 8. 抢占
// 长时间运行的 G 会被抢占
```

### 调度流程图

```
┌─────────────────────────────────────────────────────────────────────────┐
│                            调度流程                                      │
└─────────────────────────────────────────────────────────────────────────┘
                                    │
                                    ▼
                        ┌───────────────────┐
                        │    schedule()     │
                        └─────────┬─────────┘
                                  │
              ┌───────────────────┼───────────────────┐
              │                   │                   │
              ▼                   ▼                   ▼
     ┌─────────────────┐ ┌─────────────────┐ ┌─────────────────┐
     │ 每61次从全局队列 │ │  从本地队列获取  │ │ findrunnable() │
     │    获取一个G    │ │    获取 G       │ │   查找可运行G   │
     └─────────────────┘ └─────────────────┘ └────────┬────────┘
                                                      │
                    ┌─────────────────────────────────┤
                    │                                 │
                    ▼                                 ▼
           ┌─────────────────┐               ┌─────────────────┐
           │   检查本地队列   │               │   检查全局队列   │
           └─────────────────┘               └─────────────────┘
                    │                                 │
                    ▼                                 ▼
           ┌─────────────────┐               ┌─────────────────┐
           │  检查网络轮询器  │               │    工作窃取      │
           └─────────────────┘               │  (从其他P偷取)   │
                    │                         └─────────────────┘
                    ▼                                 │
           ┌─────────────────┐                       │
           │   stopm()       │◄──────────────────────┘
           │  (停止M，休眠)   │        没有可运行的 G
           └─────────────────┘
                    │
                    ▼ 找到 G
           ┌─────────────────┐
           │   execute(gp)   │
           │   执行 G        │
           └────────┬────────┘
                    │
                    ▼
           ┌─────────────────┐
           │    gogo()       │
           │  切换到 G 栈执行 │
           └────────┬────────┘
                    │
                    ▼
           ┌─────────────────┐
           │   用户代码执行   │
           └────────┬────────┘
                    │
        ┌───────────┼───────────┐
        │           │           │
        ▼           ▼           ▼
   ┌─────────┐ ┌─────────┐ ┌─────────┐
   │ 主动让出 │ │ 被抢占  │ │  退出   │
   │ Gosched │ │ preempt │ │ goexit │
   └────┬────┘ └────┬────┘ └────┬────┘
        │           │           │
        └───────────┴───────────┘
                    │
                    ▼
           ┌─────────────────┐
           │   mcall(fn)     │
           │  切换到 G0 栈   │
           └────────┬────────┘
                    │
                    ▼
           ┌─────────────────┐
           │   schedule()    │
           │   重新调度      │
           └─────────────────┘
```

---

## 调度策略

### 1. 工作窃取 (Work Stealing)

```go
// 从其他 P 窃取 G
func runqsteal(pp *p, p2 *p, stealRunNextG bool) *g {
    t := pp.runqtail
    n := runqgrab(p2, &pp.runq, t, stealRunNextG)
    if n == 0 {
        return nil
    }
    n--
    gp := pp.runq[(t+n)%uint32(len(pp.runq))].ptr()
    if n == 0 {
        return gp
    }
    // 更新队列
    atomic.StoreRel(&pp.runqtail, t+n)
    return gp
}

// 窃取一半的 G
func runqgrab(pp *p, batch *[256]guintptr, batchHead uint32, stealRunNextG bool) uint32 {
    h := atomic.LoadAcq(&pp.runqhead)
    t := atomic.LoadAcq(&pp.runqtail)
    n := t - h
    n = n - n/2 // 窃取一半
    
    // 复制到 batch
    for i := uint32(0); i < n; i++ {
        g := pp.runq[(h+i)%uint32(len(pp.runq))]
        batch[(batchHead+i)%uint32(len(batch))] = g
    }
    
    return n
}
```

### 2. Hand Off (交接)

```go
// 当 M 进入系统调用时，P 会被交接给其他 M
func handoffp(pp *p) {
    // 如果本地队列有 G，或者全局队列有 G
    if !runqempty(pp) || sched.runqsize != 0 {
        startm(pp, false) // 启动新的 M 来接管 P
        return
    }
    
    // 如果有 GC 工作
    if gcBlackenEnabled != 0 {
        startm(pp, false)
        return
    }
    
    // 没有工作，将 P 放入空闲列表
    pidleput(pp)
}
```

### 3. 自旋 (Spinning)

```go
// 自旋的 M 会积极寻找 G，而不是立即休眠
// 这样可以减少唤醒延迟

func findrunnable() (gp *g, inheritTime bool) {
    mp := getg().m
    
    // 如果已经有足够的自旋 M，则不自旋
    if !mp.spinning && 2*sched.nmspinning.Load() >= gomaxprocs-sched.npidle.Load() {
        goto stop
    }
    
    // 标记为自旋
    if !mp.spinning {
        mp.spinning = true
        sched.nmspinning.Add(1)
    }
    
    // 自旋寻找工作...
}
```

---

## 抢占机制

### 协作式抢占 (Go 1.13 及之前)

```go
// 在函数调用时检查抢占标记
// 编译器会在函数入口插入检查代码

func someFunction() {
    // 编译器插入的检查
    if getg().stackguard0 == stackPreempt {
        // 执行调度
        runtime.morestack()
    }
    
    // 函数体...
}

// 问题：没有函数调用的循环无法被抢占
for {
    // 死循环，无法抢占
}
```

### 异步抢占 (Go 1.14+)

```go
// 使用信号实现异步抢占
// Linux: SIGURG
// macOS: SIGURG
// Windows: SuspendThread

func preemptM(mp *m) {
    // 向 M 发送信号
    signalM(mp, sigPreempt)
}

// 信号处理器
func sighandler(sig uint32, info *siginfo, ctxt unsafe.Pointer, gp *g) {
    if sig == sigPreempt {
        // 执行抢占
        doSigPreempt(gp, ctxt)
    }
}

func doSigPreempt(gp *g, ctxt *sigctxt) {
    // 保存当前上下文
    // 设置 PC 指向 asyncPreempt
    ctxt.pushCall(abi.FuncPCABI0(asyncPreempt), ctxt.rip())
}

// 异步抢占入口
func asyncPreempt() {
    // 保存寄存器
    // 调用 asyncPreempt2
}

func asyncPreempt2() {
    gp := getg()
    gp.asyncSafePoint = true
    
    if gp.preemptStop {
        mcall(preemptPark) // 停止 G
    } else {
        mcall(gopreempt_m) // 重新调度
    }
}
```

### 抢占的安全点

```go
// 不能在任意位置抢占，需要在安全点

// 安全点：
// 1. 函数调用
// 2. 阻塞操作（channel、锁等）
// 3. 循环回边（通过插入抢占检查）

// 不安全的位置：
// 1. 持有运行时锁
// 2. 正在分配内存
// 3. 写屏障期间
// 4. 某些关键区域
```

### sysmon - 系统监控

```go
// sysmon 是一个特殊的 M，不需要 P
// 周期性执行以下任务：

func sysmon() {
    for {
        // 1. 检查死锁
        checkdead()
        
        // 2. 轮询网络
        if netpollinited() {
            list := netpoll(0)
            if !list.empty() {
                injectglist(&list)
            }
        }
        
        // 3. 抢占长时间运行的 G
        retake(now)
        
        // 4. 强制 GC（如果 2 分钟没有 GC）
        if t := (gcTrigger{kind: gcTriggerTime, now: now}); t.test() {
            forcegc.g.schedlink = 0
            injectglist(&forcegc.g)
        }
        
        // 休眠一段时间
        usleep(delay)
    }
}

// 抢占检查
func retake(now int64) uint32 {
    for i := 0; i < len(allp); i++ {
        pp := allp[i]
        
        // 如果 P 在系统调用中超过 10ms
        if pp.status == _Psyscall {
            if now-pp.syscalltick > 10*1000 {
                handoffp(pp) // 交接 P
            }
        }
        
        // 如果 G 运行超过 10ms
        if pp.status == _Prunning {
            if now-pp.schedtick > forcePreemptNS {
                preemptone(pp) // 抢占
            }
        }
    }
}
```

---

## 网络轮询器

### 网络轮询器架构

```
┌─────────────────────────────────────────────────────────────────────┐
│                         用户代码                                     │
│  conn.Read() / conn.Write() / net.Dial()                           │
└────────────────────────────────┬────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      net 包 / internal/poll                         │
└────────────────────────────────┬────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                      runtime/netpoll                                │
│  ┌─────────────────────────────────────────────────────────────┐   │
│  │                    netpoll 实现                               │   │
│  │  Linux: epoll                                                │   │
│  │  macOS: kqueue                                               │   │
│  │  Windows: IOCP                                               │   │
│  └─────────────────────────────────────────────────────────────┘   │
└────────────────────────────────┬────────────────────────────────────┘
                                 │
                                 ▼
┌─────────────────────────────────────────────────────────────────────┐
│                         操作系统内核                                 │
└─────────────────────────────────────────────────────────────────────┘
```

### netpoll 实现 (Linux epoll)

```go
// runtime/netpoll_epoll.go

// 初始化 epoll
func netpollinit() {
    epfd = epollcreate1(_EPOLL_CLOEXEC)
    // 创建用于唤醒的管道
    nonblockingPipe(&netpollBreakRd, &netpollBreakWr)
    // 注册唤醒管道
    epollctl(epfd, _EPOLL_CTL_ADD, netpollBreakRd, &ev)
}

// 注册 fd
func netpollopen(fd uintptr, pd *pollDesc) int32 {
    var ev epollevent
    ev.events = _EPOLLIN | _EPOLLOUT | _EPOLLRDHUP | _EPOLLET
    ev.data = (*epollevent)(unsafe.Pointer(&pd))
    return epollctl(epfd, _EPOLL_CTL_ADD, int32(fd), &ev)
}

// 等待事件
func netpoll(delay int64) gList {
    var events [128]epollevent
    n := epollwait(epfd, &events[0], int32(len(events)), waitms)
    
    var toRun gList
    for i := int32(0); i < n; i++ {
        ev := &events[i]
        pd := *(**pollDesc)(unsafe.Pointer(&ev.data))
        
        // 将等待的 G 加入就绪列表
        netpollready(&toRun, pd, mode)
    }
    return toRun
}
```

### 网络 IO 流程

```go
// 用户调用 Read
func (fd *netFD) Read(p []byte) (n int, err error) {
    // 尝试非阻塞读取
    n, err = syscall.Read(fd.pfd.Sysfd, p)
    if err == syscall.EAGAIN {
        // 没有数据，等待
        if err = fd.pd.waitRead(fd.isFile); err != nil {
            return 0, err
        }
        // 重试
        return syscall.Read(fd.pfd.Sysfd, p)
    }
    return
}

// 等待可读
func (pd *pollDesc) waitRead(isFile bool) error {
    return pd.wait('r', isFile)
}

func (pd *pollDesc) wait(mode int, isFile bool) error {
    // 将当前 G 设置为等待状态
    gpp := &pd.rg // 或 pd.wg
    for {
        old := gpp.Load()
        if old == pdReady {
            gpp.Store(0)
            return nil
        }
        if gpp.CompareAndSwap(old, pdWait) {
            break
        }
    }
    
    // 挂起当前 G
    gopark(netpollblockcommit, unsafe.Pointer(gpp), ...)
    return nil
}
```

---

## 调度器源码分析

### 关键函数

```go
// ============ 调度相关 ============

// 调度入口
schedule()

// 查找可运行的 G
findrunnable() (gp *g, inheritTime bool)

// 执行 G
execute(gp *g, inheritTime bool)

// 切换到 G 执行（汇编实现）
gogo(buf *gobuf)

// 切换回 g0（汇编实现）
mcall(fn func(*g))

// ============ G 相关 ============

// 创建 G
newproc(fn *funcval)
newproc1(fn *funcval, callergp *g, callerpc uintptr) *g

// G 退出
goexit0(gp *g)

// G 休眠
gopark(unlockf func(*g, unsafe.Pointer) bool, lock unsafe.Pointer, reason waitReason, traceReason traceBlockReason, traceskip int)

// G 唤醒
goready(gp *g, traceskip int)

// ============ M 相关 ============

// 创建 M
newm(fn func(), pp *p, id int64)

// 停止 M
stopm()

// 启动 M
startm(pp *p, spinning bool)

// ============ P 相关 ============

// 获取 P
acquirep(pp *p)

// 释放 P
releasep() *p

// P 交接
handoffp(pp *p)
```

### 启动流程

```go
// 程序启动流程

// 1. 入口（汇编）
// runtime·rt0_go

// 2. 初始化
runtime.schedinit()
    // 初始化内存分配器
    // 初始化调度器
    // 创建 P
    // 设置 GOMAXPROCS

// 3. 创建 main goroutine
runtime.newproc(runtime.main)

// 4. 启动调度
runtime.mstart()
    // 调用 schedule()
    // 开始调度循环

// 5. 执行 main goroutine
runtime.main()
    // 初始化
    // 调用 main.main()
    // 退出
```

### 调度器数据结构

```go
// 全局调度器
var sched schedt

type schedt struct {
    // 互斥锁
    lock mutex
    
    // 空闲 M 列表
    midle        muintptr // 空闲 M 链表头
    nmidle       int32    // 空闲 M 数量
    nmidlelocked int32    // 等待锁的空闲 M 数量
    mnext        int64    // 下一个 M 的 ID
    maxmcount    int32    // M 的最大数量
    
    // 空闲 P 列表
    pidle      puintptr // 空闲 P 链表头
    npidle     int32    // 空闲 P 数量
    
    // 全局运行队列
    runq     gQueue // 全局 G 队列
    runqsize int32  // 全局队列大小
    
    // 空闲 G 列表
    gFree struct {
        lock    mutex
        stack   gList // 有栈的 G
        noStack gList // 无栈的 G
        n       int32
    }
    
    // 自旋 M 数量
    nmspinning atomic.Int32
    
    // 调度统计
    // ...
}
```

---

## 性能调优

### GOMAXPROCS 设置

```go
import "runtime"

// 获取当前值
n := runtime.GOMAXPROCS(0)

// 设置新值
runtime.GOMAXPROCS(runtime.NumCPU())

// 最佳实践
// CPU 密集型：NumCPU()
// IO 密集型：NumCPU() * 2 或更高
// 混合型：根据测试调整
```

### 调度器追踪

```bash
# 开启调度器追踪
GODEBUG=schedtrace=1000 ./myapp
# 每 1000ms 输出一次调度信息

# 输出示例：
# SCHED 0ms: gomaxprocs=8 idleprocs=6 threads=5 spinningthreads=1 
#            idlethreads=0 runqueue=0 [0 0 0 0 0 0 0 0]

# 详细追踪
GODEBUG=schedtrace=1000,scheddetail=1 ./myapp
```

### 调度延迟分析

```go
import (
    "runtime"
    "runtime/trace"
)

// 使用 trace 分析
func main() {
    f, _ := os.Create("trace.out")
    trace.Start(f)
    defer trace.Stop()
    
    // 你的代码...
}

// 分析 trace
// go tool trace trace.out
```

### 常见性能问题

```go
// 1. Goroutine 泄漏
// 问题：goroutine 没有正确退出
// 解决：使用 context 控制生命周期
func worker(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            // 工作
        }
    }
}

// 2. 过多的 Goroutine
// 问题：创建了太多 goroutine
// 解决：使用 worker pool
type Pool struct {
    tasks chan func()
}

func (p *Pool) Submit(task func()) {
    p.tasks <- task
}

// 3. 锁竞争
// 问题：过多的锁竞争导致调度开销
// 解决：减少锁的范围，使用无锁数据结构

// 4. 频繁的系统调用
// 问题：频繁的系统调用导致 M 阻塞
// 解决：批量处理，使用异步 IO
```

### 监控指标

```go
import "runtime"

// Goroutine 数量
runtime.NumGoroutine()

// 内存统计
var m runtime.MemStats
runtime.ReadMemStats(&m)

// 调度器统计
runtime.NumCPU()      // CPU 核心数
runtime.GOMAXPROCS(0) // P 的数量

// pprof 分析
import _ "net/http/pprof"
go http.ListenAndServe(":6060", nil)
// 访问 http://localhost:6060/debug/pprof/
```

---

## 总结

### GMP 核心要点

| 组件 | 作用 | 数量 |
|------|------|------|
| G | 协程，执行用户代码 | 数以万计 |
| M | 系统线程，执行 G | 默认最大 10000 |
| P | 处理器，调度上下文 | 默认 = CPU 核心数 |

### 调度特点

| 特性 | 说明 |
|------|------|
| 工作窃取 | 空闲 P 从其他 P 窃取 G |
| 抢占 | Go 1.14+ 支持异步抢占 |
| 网络轮询 | 集成 epoll/kqueue/IOCP |
| Hand Off | M 阻塞时交接 P |

### 最佳实践

1. 合理设置 `GOMAXPROCS`
2. 控制 Goroutine 数量
3. 避免长时间阻塞
4. 使用 `context` 管理生命周期
5. 使用 `pprof` 和 `trace` 分析性能
