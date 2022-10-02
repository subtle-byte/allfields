package fixtures

import "github.com/subtle-byte/allfields_fixtures/fixtures2"

type Details struct {
	Detail1 string
	Detail2 string
}

type User struct {
	Name string
	Age  int
	Details
}

func F() {
	_ = User{ // want `field Age is not set`
		Name:    "John",
		Details: Details{ // want `fields Detail1, Detail2 are not set`
			//allfields
		},
		//allfields
	}

	_ = User{
		Name: "John",
	}

	_ = User{
		Name:    "John",
		Age:     25,
		Details: Details{},
		//allfields
	}

	_ = fixtures2.WithPrivate{ // want `field Name is not set`
		//allfields
	}

	_ = Details{
		"one", "two",
		//allfields // want `allfields comment is placed in a non-keyed composite struct literal`
	}

	//allfields // want `allfields comment is not used`

	//allfields lkjlkjklj // want `invalid allfields comment`

	_ = []string{
		//allfields // want `allfields comment is not used`
	}
	_ = Details{}
}
