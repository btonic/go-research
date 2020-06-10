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

- [Language Specification]()
- [Memory Model]()
- [Garbage Collection (talk)]()
- [Scheduler]()
- [Contracts Proposal]()

## IR

## Assembler

# Scheduler

## Entrypoint
A Go program is typilcally built into an arbitrary executable (typically static, can be dynamic). This executable is built based on a `package main`, which contains a `func main()`. However, because Go compiles a runtime into the binary to handle things such as garbage collection, coroutine scheduling, OS threads, and similar services, the developer-defined `main` is managed through the `runtime.main` function. After these initialization tasks have completed, the `runtime.main` invokes the `main.main`. Upon `main.main` 

## 

## Coroutine scheduling
In Go you can spawn "coroutines", or "threads" in most other languages. Unlike interpreted languages such as Python/Ruby (in some variants), there is no Global Interpreter like functionality, but there is also no system-thread like functionality. Go handles all of this for you through the concepts of:

```go
// The main concepts are:
// G - goroutine.
// M - worker thread, or machine.
// P - processor, a resource that is required to execute Go code.
//     M must have an associated P to execute Go code, however it can be
//     blocked or in a syscall w/o an associated P.
```

A given set of N-coroutines (G) execute on a process (P), and a process is a one-to-one mapping to an OS thread (M). Gs are managed in a queue-like structure, where any MP, if available, can execute a given set of Gs. If no Gs are available, a rebalancing of Gs from other MPs is performed through attempts to steal them.

## Garbage Collection
The GC uses locking and mutex semantics through this concept of "stopping the world," where in order to identify values available for GC it will attempt to park every coroutine, perform it's cleanup, and "resume the world". It does this on a set timer, asynchronously. Now, something that is also interesting is that because it must obey the G parking status, the program execution can affect GC execution indirectly, as it is possible to prevent a coroutine from successfully parking.