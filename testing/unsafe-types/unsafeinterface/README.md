# Unsafe Interfaces
## Building
```
$ go run .
win 92 5319
exit status 1
```

## Underlying interface structures
```go
type itab struct {
	inter *interfacetype
	_type *_type
	hash  uint32 // copy of _type.hash. Used for type switches.
	_     [4]byte
	fun   [1]uintptr // variable sized. fun[0]==0 means _type does not implement inter.
}
```
_Snippet from [go/src/runtime/runtime2.go](https://github.com/btonic/go-research/blob/e2ab4a841a35cad07b35fee7d5ac193d910f43b4/go/src/runtime/runtime2.go#L807-L813)_

```go
type iface struct {
	tab  *itab
	data unsafe.Pointer
}
```
_Snippet from [go/src/runtime/runtime2.go](https://github.com/btonic/go-research/blob/e2ab4a841a35cad07b35fee7d5ac193d910f43b4/go/src/runtime/runtime2.go#L200-L203)_

## Conformity Race
An interface defines a set of type-bound methods which must be implemented for a type to conform to the interface. This is enforced during the assignment of a value to any variable defined as requiring a given interface (e.g. `var x MyInterface` can only be assigned to value types implement `MyInterface`-required type-bound methods). In order for this to work, the underlying type of an `interface{}` is represented as a structure containing pointers to the interface table (`tab`), and data (`data`).

Because the evaulation of `data` is dependent on the type information stored within `tab`, interfaces are prone to race conditions during reassignment, which could potentially leading to incorrect evaluation of `data`.

```go
type C interface {
    Exec()
}

type A struct {
    fintptr *int
}

func (a *A) Exec() {
    // some sort of work where `fintptr` can be reassigned by an attacker
}

type B struct {
    fptr func()
}
func (b B) Exec() {
    // some sort of work here, where `fptr` is executed (e.g. `fptr()`)
}
```

Given the definitions of `A`, `B`, `C` and their bound functions to implement `C`'s interface, the following is the premise of the dangerous race:

1. Variable `confused` (an arbitrary variable) is a variable of type `C`.
2. `confused.Exec()` is repeatedly executed in a loop within goroutine `A`.
3. `confused` is repeatedly reassigned in an arbitrary goroutine, which we will denote as `B`.
4. `confused.Exec()` is executed when `data` points to that of an arbitrary `A`, but the type information points to that of an arbitrary `B`.
5. The confused state results in the evaluation of `A.fintptr` as `B.fptr`, resulting in `B.Exec` executing the attacker-supplied `A.fintptr`.
6. Execution of the function at address `A.fintptr` begins in coroutine A.

In our [example](./main.go), we can observe this type of race in a pure form. We can define a `win` function, which is what we will be racing to execute, and an `address` function which will let us calculate an integer-formatted address for use during our race.

```go
// Utility functions

// i and j simply hold the loop counters for the `main` coroutine, and the spawned coroutine.
var i, j int

// address simply returns the integer address of a given parameter.
func address(i interface{}) int {
	addr, err := strconv.ParseUint(fmt.Sprintf("%p", i), 0, 0)
	if err != nil {
		panic(err)
	}
	return int(addr)
}

// For our example, we will be invoking `win` through the `StructIntPtr`'s 	`funcptr` field.
// Win is a simple target because it expects no parameters, so we don't have to groom the stack.
func win() {
	fmt.Println("win", i, j)
	os.Exit(1)
}
```

We then have two structures (equivalent to `A` and `B`), which implement the same interface (equivalent to `C`).

```go

// Interface defines the interface which both types in our race must comply with.
// In this case, each type must implement a function `X` with no return value.
type Interface interface {
	Exec()
}

// StructIntPtr contains a single field, `funcptr`, which in our case is some sort of user-controlled
// address. Effectively, this is the field pointing to the address we would like to execute during our
// race.
type StructIntPtr struct {
	funcintptr *int
}

// X is stubbed out in the StructIntPtr for our example, so we don't have to worry about other logic
// impeding our ability to hit the race condition.
func (s StructIntPtr) Exec() {}

// StructFuncPtr defines an interface with a normal `funcptr` (the typical way you would reference a function pointer).
// This structure is what will execute our stale (`funcintptr`) address, as `funcptr` will inherit the value of `StructIntPtr.funcintptr`.
type StructFuncPtr struct {
	funcptr func()
}

func (u StructFuncPtr) Exec() {
	// Because we are not initializing `u.`
	if u.funcptr != nil {
		u.funcptr()
	}
}
```
Within `main`, we then calculate the `win` address so we can use it to emulate the attacker-controlled value of our race. We then initialize `a` and `b` to each of the respective struct types, and assign `confused` to `a` for an initial value. At this point, we are then ready to begin our racing operations. The main goroutine will spawn a new goroutine, which will infinitely loop, reassigning the value of `confused` back and forth between `a` and `b`. After this coroutine is spawned, the main goroutine will then loop infinitely, executing `confused.Exec()`. When a race occurs, the spawned goroutine will continue reassigning `confused`, while the main goroutine will then transition to executing the `win` function.

```go
func main() {
	// A minimum of two CPUs are necessary, since our corouines must execute on different MPs.
	if runtime.NumCPU() < 2 {
		fmt.Println("need >= 2 CPUs")
		os.Exit(1)
	}

	// The winAddr is the address of the `win` function.
	winAddr := address(win)

	// We are defining the variables involved in the race:
	//
	// * `confused`: The confused variable is what will be used as a proxy to `a` and `b`
	// * `a`: The struct we will be using during the race to overwrite `StructFuncPtr.funcptr` in `b`.
	// * `b`: The struct we will be executing our inherited `funcintptr` value from `a` within.
	var confused, a, b Interface
	a = &StructIntPtr{funcintptr: &winAddr}
	b = &StructFuncPtr{}

	// Since confused was not initialized to a value, we are going to start with a value of a,
	// Which has a stubbed `X`. That way, we can let the coroutine spawn and start assigning values
	// to `confused` between `a` and `b`.
	confused = a

	// Spawn the racy coroutine.
	go func() {
		// We are going to infinitely loop, reassigning confused back and forth between `a` and `b`.
		// When the race condition is hit in the mainthread, this coroutine won't matter anymore,
		// As we will be able to begin controlling execution flow, regardless of `confused`.
		for {
			confused = b

			// We need to keep the assignment to `confused` from being inlined by the compiler.
			// Referencing `confused` from inside an inlined function will allow us to force
			// inlining to be too expensive.
			func() {
				if i >= 0 {
					return
				}
				fmt.Println(confused)
			}()

			confused = a
			// Bump i to show how many times (roughly) the spawned coroutine executed prior to
			// achieving control of execution flow.
			i++
		}
	}()

	// In the `main` coroutine, repeatedly executes `confused.Exec()` so we can attempt
	// to invoke the `funcintptr` value leftover from `a`, as `funcptr`.
	for {
		confused.Exec()
		j++
	}
}
```