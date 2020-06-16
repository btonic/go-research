# What is this?
This repository is an attempt to detail interesting semantics of the Go language. While it is not directly targetted at any one particular area, the overall goal is to highlight how language semantics interact with eachother, and areas which may be problematic due to these interactions.

# Why?
Research. Bugs. Nihilistic angst.

---
# Local Environment Configuration
To have the same environment we are testing against, you can follow the following steps for setup.

1. Ensure no existing install of Go is present on your machine (seriously, this causes issues if things are scattered amongst your path).
2. Clone the Go `1.4` release source into `~/go1.4`.
3. Build the `1.4` release (this is your bootstrap toolchain).
4. Clone the target branch (e.g. `1.14.4`).
5. Link the cloned target branch to a location in your `PATH` (e.g. `/usr/local/bin`).
6. Build the cloned target branch.
7. Test your installation (easiest is to `go get` a package w/ a built bin, and executing the bin).

# Compiler
## Spec
The Go specification details the grammar and expected semantics during construction of Go programs. However, there are additional specifications scattered across multiple places. For convenience (and aggregation), these documents have been detailed below.

- [Language Specification](https://golang.org/ref/spec)
- [Memory Model](https://golang.org/ref/mem)
- [Garbage Collection Semantics](https://www.ardanlabs.com/blog/2018/12/garbage-collection-in-go-part1-semantics.html)
- [Garbage Collection Tracing](https://www.ardanlabs.com/blog/2019/05/garbage-collection-in-go-part2-gctraces.html)
- [Garbage Collection Pacing](https://www.ardanlabs.com/blog/2019/07/garbage-collection-in-go-part3-gcpacing.html)
- [Summary of OS Thread Scheduling (non-go-related)](https://www.ardanlabs.com/blog/2018/08/scheduling-in-go-part1.html)
- [Go's Scheduler](https://www.ardanlabs.com/blog/2018/08/scheduling-in-go-part2.html)
- [Go Escape analysis](http://www.agardner.me/golang/garbage/collection/gc/escape/analysis/2015/10/18/go-escape-analysis.html)
- [Go 1.14 compiler changes](https://tip.golang.org/doc/go1.14#compiler)

## IR

## Assembler

# Scheduler

## Entrypoint
A Go program is typilcally built into an arbitrary executable (typically static, can be dynamic). This executable is built based on a `package main`, which contains a `func main()`. However, because Go compiles a runtime into the binary to handle things such as garbage collection, coroutine scheduling, OS threads, and similar services, the developer-defined `main` is managed through the `runtime.main` function. After these initialization tasks have completed, the `runtime.main` invokes the `main.main`. Upon `main.main` 

## 

## Coroutine scheduling
In Go you can spawn coroutines to perform asynchronous tasks. Unlike interpreted languages such as Python/Ruby (in some variants), there is no Global Interpreter like functionality. Go handles all of this for you through the concepts of:

```go
// The main concepts are:
// G - goroutine.
// M - worker thread, or machine.
// P - processor, a resource that is required to execute Go code.
//     M must have an associated P to execute Go code, however it can be
//     blocked or in a syscall w/o an associated P.
```
_Extracted from [go/src/runtime/proc.go](https://github.com/btonic/go-research/blob/e2ab4a841a35cad07b35fee7d5ac193d910f43b4/go/src/runtime/proc.go#L19-L29)._

In breif, given set of N-coroutines (G) execute on a process (P), and a process is a one-to-one mapping to an OS thread (M). Gs are managed in a queue-like structure, where any MP, if available, can execute a given set of Gs. If no Gs are available from the queue, a rebalancing of Gs from other MPs is performed through attempts to steal them.

## Garbage Collection
Given the fact that the scheduler handles execution of program-requested coroutines, the garbage collector (GC) is able to perform it's job through similar semantics as the program. The GC uses locking and mutex semantics by "stopping the world," where in order to identify values available for GC it will attempt to park coroutines, perform it's cleanup, and "resume the world". It does this on a set timer, asynchronously. 

Interestingly, because the GC must obey the G parking status, the program execution can affect GC execution indirectly, as it is possible to prevent a coroutine from successfully parking.

## Safety of types
In Go, there is a general understanding that all operations are memory safe. However, this only holds when race conditions are not present. Certain types in Go are represented by underlying structures. These operations do not always enforce atomic properties. As a result, we can manipulate memory in a way that is unsafe. These types have been [documented within this repository](testing/unsafe-types/README.md).