# bookdata-api

## sync.RWMutex

There is a race condition on CREATE and DELETE unless you use a RWMutex to handle locking.

I added a race_test.go to demonstrate the problem (though it will pass now because I fixed it). You can check for race conditions during test or runtime by adding the -race flag.

## Deleting

I chose to do a "re-slice" to preserve order...

```go
s = append(s[:index], s[index+1:]...)
```

...but if this were a big list you might consider one of:

- using a linked list

- not preserving the order

- marking an item as deleted rather than actually deleting it

## Custom Error

Here is an example of throwing a custom error...

```go
type InsertBookConflictError struct {
	Msg string
}

func (e *InsertBookConflictError) Error() string {
	return e.Msg
}

func (b *Books) InsertBook(book *loader.BookData) error {

	// check for ISBN conflict
	if b.GetByISBN(book.ISBN) != nil {
		return &InsertBookConflictError{fmt.Sprintf("InsertBook() ISBN conflict on %v", book.ISBN)}
	}

	return nil
}
```

And catching it...

```go
err := books.InsertBook(&book)
var conflict *datastore.InsertBookConflictError
if errors.As(err, &conflict) {
	w.WriteHeader(http.StatusConflict)
	resp := fmt.Sprintf(`{"error": "%v"}`, err.Error())
	w.Write([]byte(resp))
	return
}
if err != nil {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(`{"error": "unknown error inserting the book"}`))
	return
}
```

## Inheritance

Here are some learnings for interfaces, inheritance, and composition.

```go
package main

import (
	"fmt"
)

type iNamePrinter interface {
	print()
}

type base struct {
	name string
}

func (b *base) print() {
	fmt.Println(b.name)
}

type derived struct {
	base
	name string
}

func (d *derived) print() {
    // NOTE: base is called an "anonymouse field", it is embedded in this struct
	fmt.Printf("%v %v\n", d.name, d.base.name)
}

type derivedWithoutPointer struct {
	base
	name string
}

func (d derivedWithoutPointer) print() {
	fmt.Printf("%v %v\n", d.name, d.base.name)
}

func main() {

    // NOTE: defining the interface type here does throw compilation errors if not implemented
    var b iNamePrinter
    // NOTE: because print() is defined on the pointer type for base, a pointer to the struct must be passed
	b = &base{name: "Alice"}
    b.print()
    
    // NOTE: here is the syntax to get the base (embedded) struct
    // NOTE: here is also some alternate syntax for assigning to the pointer
	var d iNamePrinter = (&derived{name: "Brad", base: base{name: "Carl"}})
    d.print()
    
    // NOTE: if your methods aren't on the pointer, you can assign directly
	var n iNamePrinter = derivedWithoutPointer{name: "Darla", base: base{name: "Erik"}}
    n.print()
    
}
```