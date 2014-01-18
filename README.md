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

## Usage & Example

For usage and examples see the [Godoc](http://godoc.org/github.com/mitchellh/mapstructure).

The `Decode` and `DecodePath` functions have examples associated with it there.

## But Why?!

Go offers fantastic standard libraries for decoding formats such as JSON.
The standard method is to have a struct pre-created, and populate that struct
from the bytes of the encoded format. This is great, but the problem is if
you have configuration or an encoding that changes slightly depending on
specific fields. For example, consider this JSON:

```json
{
  "type": "person",
  "name": "Mitchell"
}
```

Perhaps we can't populate a specific structure without first reading
the "type" field from the JSON. We could always do two passes over the
decoding of the JSON (reading the "type" first, and the rest later).
However, it is much simpler to just decode this into a `map[string]interface{}`
structure, read the "type" key, then use something like this library
to decode it into the proper structure.

## DecodePath

Sometimes you have a large and complex JSON document where you only need to decode
a small part.

```
"userContext": {
    "conversationCredentials": {
        "sessionToken": "06142010_1:75bf6a413327dd71ebe8f3f30c5a4210a9b11e93c028d6e11abfca7ff"
    },
	"valid": true,
    "isPasswordExpired": false,
    "cobrandId": 10000004,
    "channelId": -1,
    "locale": "en_US",
    "tncVersion": 2,
    "cobrandConversationCredentials": {
        "sessionToken": "06142010_1:b8d011fefbab8bf1753391b074ffedf9578612d676ed2b7f073b5785b"
    },
	"loginName": "sptest1",
	"userType": {
        "userTypeId": 1,
        "userTypeName": "normal_user"
    }
}
```
It is nice to be able to define and pull the documents and fields you need without
having to map the entire JSON structure.

```
type UserType struct {
	UserTypeId   int
	UserTypeName string
}
	
type User struct {
		Session   string   `xpath:"userContext.cobrandConversationCredentials.sessionToken"`
		CobrandId int      `xpath:"userContext.cobrandId"`
		UserType  UserType `xpath:"userType"`
		LoginName string   `xpath:"loginName"`
}

user := User{}
mapstructure.DecodePath(docMap, &user)
```
The DecodePath function give you this functionality.