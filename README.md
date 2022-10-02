# allfields

allfields is linter for Go programming language. It checks that all fields in the struct literal are set.

allfields checks only the struct literals with the `//allfields` comments just inside them. In the following example the linter will throw error because the `Age` field is not set while creating `userAlice`, but `userBob` will successfully pass the checks. To run the allfields linter use the command like `go run github.com/subtle-byte/allfields/cmd/go-allfields path/to/packages`.

```go
type User struct {
    Name string
    Age int
}

func main() {
    userAlice := User{
        Name: "Alice",
        //allfields
    }
    userBob := User{
        Name: "Bob",
        Age:  20,
        //allfields
    }
}
```
