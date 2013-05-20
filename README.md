# mapstructure

mapstructure is a Go library for decoding generic map values to structures
and vice versa, while providing helpful error handling.

This library is most useful when decoding values from some data stream (JSON,
Gob, etc.) where you don't _quite_ know the structure of the underlying data
until you read a part of it. You can therefore read a `map[string]interface{}`
and use this library to decode it into the proper underlying native Go
structure.

## Installation

Standard `go get`:

```
$ go get github.com/mitchellh/mapstructure
```

## Example

```go
import "mapstructure"

type Person struct {
	Name   string
	Age    int
	Emails []string
	Extra  map[string]string
}

// This input can come from anywhere, but typically comes from
// something like decoding JSON where we're not quite sure of the
// struct initially.
input := map[string]interface{}{
	"name":   "Mitchell",
	"age":    91,
	"emails": []string{"one", "two", "three"},
	"extra": map[string]string{
		"twitter": "mitchellh",
	},
}

var result Person
err := Decode(input, &result)
if err != nil {
	panic(err)
}

// The value of "result" now contains what you would expect. The decoding
// process is properly type-checked and human-friendly errors are returned,
// if any.
fmt.Printf("%#v", result)
```
