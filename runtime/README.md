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