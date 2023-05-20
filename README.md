# allfields

allfields is a Go linter. It checks that all fields in the struct literal are set.

allfields checks only the struct literals with the `//allfields` comments just inside them. In the following example the linter will throw error because the `Age` field is not set while creating `userAlice`, but `userBob` will successfully pass the checks. To run the allfields linter use the command like `go run github.com/subtle-byte/allfields/cmd/go-allfields path/to/packages` (`path/to/packages` can be `./...` for example).

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

Also, you can use the `//allfields:lint` comment instead of `//allfields` if you think it looks better or your IDE handles it better.

### Use cases

Developing backend services in Go we frequently meet the situation when we need to copy the data between structs. Let's imagine we write grpc server:

```go
func (s *grpcServer) GetUser(ctx context.Context, req *api.GetUserRequest) (*api.GetUserResponse, error) {
	user := s.Service.GetUser(ctx, req.Id)
	return &api.GetUserResponse{
		Id: user.ID,		
		Name: user.Name,
		Age: user.Age,
		//allfields
	}, nil
}
```

In this example `//allfields` guarantees that if you extend API by adding fields to `User` (for example adding field `CreatedAt`) you will not forget to set this new field in `GetUserResponse`.

### Run from tests

You can run allfields from tests. To do this you need to add the following code:

```go
import (
	"testing"
	"github.com/subtle-byte/allfields"
)

func TestAllFields(t *testing.T) {
	allfields.Analyze(allfields.AnalyzeConfig{
		PackagesPattern: "./...",
		ReportErr: func(message string) {
			t.Error(message)
		},
	})
}
```
