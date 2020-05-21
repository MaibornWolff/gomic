package model

import (
	"fmt"
	"strings"
)

type Person struct {
	FirstName string `bson:"firstName" json:"firstName"`
	LastName  string `bson:"lastName" json:"lastName"`
}

func (person Person) String() string {
	return fmt.Sprintf("%s, %s", person.LastName, person.FirstName)
}

func (person Person) WithUpperCase() Person {
	return Person{FirstName: strings.ToUpper(person.FirstName), LastName: strings.ToUpper(person.LastName)}
}
