package fixtures

import "github.com/subtle-byte/allfields_fixtures/fixtures2"

type Details struct {
	Detail1 string
	Detail2 string
}

type User struct {
	Name    string
	Age     int
	Details Details
}

func F() {
	_ = User{
		Name:    "John",
		Details: Details{ // want `fields Detail1, Detail2 are not set`
			//allfields
		},
		//allfields ignore=Age,Name // want `field Name is marked as ignored but is present in the struct literal`
	}

	_ = User{
		Name: "John",
	}

	_ = User{ // want `field Details is not set`
		Name: "John",
		Age:  25,
		//allfields ignore=Abc // want `field Abc is not present in the struct but ignored`
	}

	_ = fixtures2.WithPrivate{ // want `field Name is not set`
		//allfields ignore=somePrivate // want `unexported field somePrivate is not available in this package, so the field should not be ignored`
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
