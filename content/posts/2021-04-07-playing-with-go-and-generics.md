---
title: "Playing with Go and Generics"
date: 2021-04-07
draft: false
---

Yes, it's happening!
The [proposal to add generics](https://go.googlesource.com/proposal/+/refs/heads/master/design/43651-type-parameters.md) to Golang was [accepted](https://github.com/golang/go/issues/43651#issuecomment-776944155).
It means that in some versions in the future, we'll be able to code generic solutions without having to use complex workarounds with the empty `interface{}`.
But, how it works?

**TLDR**: Now there's a _type parameter_ described with square brackets to define generics types for structs, functions, etc.
For instance:
```golang
func Print[T any](s []T) {
	for _, v := range s {
		fmt.Print(v)
	}
}
```
_The identifier `any` means that the type `T` can be any type_

With the previous notation we can use a slice of any type in an explicit way:

```golang
func main() {
	Print([]string{"Hello, ", "World\n"})
	Print([]int{4, 2})
}
```
And will work like a charm:
```
Hello, World
42
```
Let's take a look at some classics data structures implemented in go with generics.

## Queue
We can start with a simple _queue_ implementation:

```golang
type Queue[T any] []T

func (q *Queue[T]) enqueue(v T) {
	*q = append(*q, v)
}

func (q *Queue[T]) dequeue() (T, bool) {
	if len(*q) == 0 {
		var v T
		return v, false
	}
	v := (*q)[0]
	*q = (*q)[1:]
	return v, true
}
```

When we testing the _generic queue_ with `string` type:
```golang
func main() {
	q := new(Queue[string])
	q.enqueue("item-1")
	q.enqueue("item-2")
	q.enqueue("item-3")
	fmt.Println(q)
	fmt.Println(q.dequeue())
	fmt.Println(q.dequeue())
	fmt.Println(q)
}
```
We got the following output:
```
&[reg-1 reg-2 reg-3]
reg-1 true
reg-2 true
&[reg-3]
```

## Stack
If we try a _Stack_?:
```golang
type Stack[T any] []T

func (s *Stack[T]) push(v T) {
	*s = append([]T{v}, (*s)...)
}

func (s *Stack[T]) pop() (T, bool) {
	if len(*s) == 0 {
		var v T
		return v, false
	}
	v := (*s)[0]
	*s = (*s)[1:]
	return v, true
}
```
What about test using `int`?:
```golang
func main() {
	s := new(Stack[int])
	s.push(1)
	s.push(2)
	s.push(3)
	fmt.Println(s)
	fmt.Println(s.pop())
	fmt.Println(s.pop())
	fmt.Println(s)
}
```
No big deal:
```
&[3 2 1]
3 true
2 true
&[1]
```

## Linked List
Maybe a _Linked List_:
```golang
type Node[T any] struct {
	data T
	next *Node[T]
}

type LinkedList[T any] struct {
	head *Node[T]
}

func (l *LinkedList[T]) add(data T) {
	node := &Node[T]{data: data}
	if l.head == nil {
		l.head = node
		return
	}
	current := l.head
	for current.next != nil {
		current = current.next
	}
	current.next = node
}

func (l *LinkedList[T]) print() {
	current := l.head
	for current != nil {
		fmt.Println(current.data)
		current = current.next
	}
}
```
A bit more of code I agree but there's no secrets:
```golang
func main() {
	ll := new(LinkedList[float32])
	ll.add(1.01)
	ll.add(2.02)
	ll.add(3.03)
	ll.print()
}
```
Everything works as expected:
```
1.01
2.02
3.03
```

It looks a little strange on the first try, but I can get used to it.

## A more restrictive example

We can define a more restrictive set of types that a generic type can use.

```golang
type Number interface {
	type int, float32
}

func sum[T Number](a, b T) T {
	return a + b
}
```
For `int` and `float32` that function works fine:
```golang
func main() {
	fmt.Println(sum(10, 30))
	fmt.Println(sum[float32](1.5, 4.2))
}
```
But if we try with `strings` for example, the code won't compile:

```
string does not satisfy Number (string or string not found in int, float32)
```

## Try It Yourself!
You can play with generics in go by yourself by [build from the source](https://go.googlesource.com/go/+/refs/heads/dev.go2go/README.go2go.md) or in [The Go2Go Playground](https://go2goplay.golang.org/).
