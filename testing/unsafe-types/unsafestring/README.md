# Unsafe Strings
## Building
```
$ go run .
...
aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa

hms + @ P [) )()
, ->..25: > CcCfCoCsLlLmLoLtLuMcMeMnNdNlNoPcPdPePfPiPoPsScSkSmSoTZYiZlZpZs"

 ][]
i)msnss us}
  G  M  P *( -  <  >  m=%: ../125625???EOFHanLaoMayMroNaNNkoPC=PWDUTCVai]:
adxaesavxdupendfmagc gp intmapnilobjpc=ptrµsμs <== at  fp= in  is  lr: of  on  pc= sp: sp=) = ) m=+Inf, n -Inf3125: p=AhomChamDashGOGCJulyJuneLisuMiaoModiNewaThai
        m=] = ] n=avx2basebmi1bmi2boolcallcas1cas2cas3cas4cas5cas6chandeadermsfilefuncidleint8kindopensbrksse2sse3stattrueuint ...
 H_T= H_a= H_g= MB,  W_a= and  cnt= h_a= h_g= h_t= max= ptr  siz= tab= top= u_a= u_g=, ..., fp:/etc/1562578125<nil>AdlamAprilBamumBatakBuhidDograErrorGreekKhmerLatinLimbuLocalMarchNushuOghamOriyaOsageRunicSTermTakriTamil] = (argp=arraychmodclosefalsefaultfcntlgcinggetwdint16int32int64lstatpanicscav sleepslicesse41sse42ssse3uint8write Value addr= base  code= ctxt: curg= goid  jobs= list= m->p= next= p->m= prev= span= varp=% util(...)
, i = , not 390625<-chanArabicAugustBrahmiCarianChakmaCommonCopticFormatFridayGOROOTGot
```
## Underlying string structures
```go
type stringStruct struct {
	str unsafe.Pointer
	len int
}
```
_Snippet from [go/src/runtime/string.go](https://github.com/btonic/go-research/blob/e2ab4a841a35cad07b35fee7d5ac193d910f43b4/go/src/runtime/string.go#L218-L221)_

## String Length Race
The `string` type in Go is represented by an underlying structure composed of two fields: a pointer to the string (`str`), and the length of `str`'s contents. As a result, there is an opportunity for string operations to be raced to produce unintended behavior. In this case, we are able to read the contents of the heap or stack (depending on if the variable escapes the stack) around a given `string` variable.

Given the definition of `a`, `b`, and `confused` as strings, the following is the premise of the dangerous race:

1. `a` is assigned a value with a greater length than `b`.
2. `b` is assigned a value with a lesser length than `a`.
3. The value of `confused` is repeatedly printed from index `[length(b):]` in goroutine `A`.
4. The `confused` variable is repeatedly reassigned between `a` and `b` in goroutine `B`.
5. A race is encountered in goroutine `A`, where the underlying structure representing `confused` has a length of `a`, but points to `b`, and thus leaks `length(a) - length(b)` bytes of surrounding memory through indexing.

In our [example](./main.go), we can observe this type of race in a pure form. We first begin by defining our variables, `a`, `b`, and `confused`. We initialize both `a` and `b` to values, where `a` is our long string which will allow us to read `length(a) - length(b)` bytes adjacent to `b`.

```go
	// Necessary for preventing inlining of the nested func.
	var i int = 0

	// Define `a`, `b`, and `confused`, which will allow our race to leak data.
	var a, b, confused string

	// The length of `a-1` is the amount of data we will be able to leak relative to the location
	// of the variable in the heap.
	a = strings.Repeat("a", 1000)
	// An arbitrary value provided, since we just need another value to assign to.
    b = "b"
```

We then spawn our goroutine to reassign `confused` back and forth between the value of `a` and `b` repeatedly.

```go
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
```

In the main goroutine, we then infinitely loop, printing the value of `confused[1:]`. When the race is encountered and `confused` points to the `ptr` of `b`, but the `len` of `a`, the `fmt.Println` will print `length(a) - length(b)` of memory adjacent to `b`. However, until we hit this race, we must `recover()` errors, since it is possible we will attempt to index `b`, which will yield a `slice bounds out of range` error.

```go
	// Repeatedly attempt to print elements 1 and on. This will panic when `confused` is assigned
	// to `the length of `b`, but not when assigned to the length of `a`.
	for {
		func() {
			defer func() { recover() }()
			fmt.Println(confused[1:])
		}()
	}
```