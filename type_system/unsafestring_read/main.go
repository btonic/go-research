/*
Inspiration from https://github.com/StalkR/misc
*/

package main

import (
	"fmt"
	"strings"
)

func main() {
	// Necessary for preventing inlining of the nested func.
	var i int = 0

	// Define `a`, `b`, and `confused`, which will allow our race to leak data.
	var a, b, confused string

	// The length of `a-1` is the amount of data we will be able to leak relative to the location
	// of the variable in the heap.
	a = strings.Repeat("a", 1000)
	// An arbitrary value provided, since we just need another value to assign to.
	b = "b"

	// Our coroutine which re-assigns `confused` to `a` and `b` repeatedly.
	go func() {
		for {
			confused = a

			// Avoid the compiler from inlining & optimizing out our reassignments.
			func() {
				if i >= 0 {
					return
				}
				fmt.Println(confused)
			}()

			confused = b
		}
	}()

	// Repeatedly attempt to print elements 1 and on. This will panic when `confused` is assigned
	// to `the length of `b`, but not when assigned to the length of `a`.
	for {
		func() {
			defer func() { recover() }()
			fmt.Println(confused[1:])
		}()
	}
}
