# mapstructure

mapstructure is a Go library for decoding generic map values to structures
and vice versa, while providing helpful error handling.

This library is most useful when decoding values from some data stream (JSON,
Gob, etc.) where you don't _quite_ know the structure of the underlying data
until you read a part of it. You can therefore read a `map[string]interface{}`
and use this library to decode it into the proper underlying native Go
structure.

## Example

```go
import "mapstructure"

type Person struct {
	name   string
	age    uint
	emails []string
}

// You can imagine that the "input" comes from some external source
// such as decoding JSON or something.
input := map[string]interface{}{
	"name": "Mitchell",
	"age": 91,
	"emails": []string{"foo@bar.com", "bar@baz.com"},
}

var result Person
err := mapstructure.Decode(input, &result)
if err != nil {
	panic(err)
}

// The value of "result" now contains what you would expect. The decoding
// process is properly type-checked and human-friendly errors are returned,
// if any.
```
