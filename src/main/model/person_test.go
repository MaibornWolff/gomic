package model

import (
	"testing"
)

func TestStringConversion(t *testing.T) {
	person := Person{FirstName: "John", LastName: "Doe"}
	personString := person.String()
	expectedPersonString := "Doe, John"

	if personString != expectedPersonString {
		t.Errorf("Expected '%s' for Person's string representation, got '%s'", expectedPersonString, personString)
	}
}

func TestUpperCasing(t *testing.T) {
	person := Person{FirstName: "John", LastName: "Doe"}
	upperCasedPerson := person.WithUpperCase()
	expectedPerson := Person{FirstName: "JOHN", LastName: "DOE"}

	if upperCasedPerson != expectedPerson {
		t.Errorf("Expected '%s' for Person in upper case, got '%s'", expectedPerson, upperCasedPerson)
	}
}
