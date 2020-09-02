# What is this?
This repository is an attempt to detail interesting semantics of the Go language. While it is not directly targetted at any one particular area, the overall goal is to highlight how language semantics interact with eachother, and areas which may be problematic due to these interactions.

# Why?
Research. Bugs. Nihilistic angst.

---

# Areas of research
1. [Compiler](compiler/) - Reserach related to how Go programs are constructed (IR/assembler/codegen/etc).
2. [Runtime](runtime/) - Research related to how coroutines, scheduling, and process management are performed.
3. [Type system](type_system/) - Research into how the type system works and affects program behavior.

# Local Environment Configuration
To have the same environment we are testing against, you can follow the following steps for setup.

1. Ensure no existing install of Go is present on your machine (seriously, this causes issues if things are scattered amongst your path).
2. Clone the Go `1.4` release source into `~/go1.4`.
3. Build the `1.4` release (this is your bootstrap toolchain).
4. Clone the target branch (e.g. `1.14.4`).
5. Link the cloned target branch to a location in your `PATH` (e.g. `/usr/local/bin`).
6. Build the cloned target branch.
7. Test your installation (easiest is to `go get` a package w/ a built bin, and executing the bin).