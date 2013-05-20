package mapstructure

import (
	"fmt"
)

type Person struct {
	Name   string
	Age    int
	Emails []string
	Extra  map[string]string
}

func ExampleDecode() {
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

	fmt.Printf("%#v", result)
	// Output:
	// mapstructure.Person{Name:"Mitchell", Age:91, Emails:[]string{"one", "two", "three"}, Extra:map[string]string{"twitter":"mitchellh"}}
}

func ExampleDecode_errors() {
	// This input can come from anywhere, but typically comes from
	// something like decoding JSON where we're not quite sure of the
	// struct initially.
	input := map[string]interface{}{
		"name":   123,
		"age":    "bad value",
		"emails": []int{1, 2, 3},
	}

	var result Person
	err := Decode(input, &result)
	if err == nil {
		panic("should have an error")
	}

	fmt.Println(err.Error())
	// Output:
	// 5 error(s) decoding:
	//
	// * 'root.Name' expected type 'string', got 'int'
	// * 'root.Age' expected type 'int', got 'string'
	// * 'root.Emails[0]' expected type 'string', got 'int'
	// * 'root.Emails[1]' expected type 'string', got 'int'
	// * 'root.Emails[2]' expected type 'string', got 'int'
}
