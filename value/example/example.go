package main

import (
	"fmt"
	"github.com/hneemann/parser2/value"
)

type Person struct {
	Name         string
	Surname      string
	PlaceOfBirth string
	Age          int
}

var Persons = []Person{
	{"John", "Doe", "London", 23},
	{"Jane", "Doe", "London", 25},
	{"Bob", "Smith", "New York", 21},
	{"Frank", "Muller", "New York", 22},
	{"Mary", "Green", "Seattle", 21},
	{"Jake", "Muller", "Washington", 22},
}

var PersonToMap = value.NewToMapReflection[Person]()

func main() {
	// Create a parser.
	parser := value.New(nil)
	// Create a list to be used containing the persons.
	persons := value.NewListOfMaps[Person](PersonToMap, Persons)
	{
		// Create a function that evaluates the list of persons.
		// The argument 'persons' is passed to the function.
		fu, err := parser.Generate(`

persons.map(p->p.Name).reduce((a,b)->a+", "+b)

        `, "persons")
		if err != nil {
			panic(err)
		}
		// Evaluate the function.
		result, err := fu.Eval(persons)
		if err != nil {
			panic(err)
		}
		fmt.Println(result)
	}

	{
		fu, err := parser.Generate(`

persons
  .accept(p->p.PlaceOfBirth="New York" & p.Age>21)
  .map(e->e.Name+": "+e.Age)
  .reduce((a,b)->a+", "+b)

        `, "persons")
		if err != nil {
			panic(err)
		}
		// Evaluate the function.
		result, err := fu.Eval(persons)
		if err != nil {
			panic(err)
		}
		fmt.Println(result)
	}

	{
		fu, err := parser.Generate(`

persons
  .groupByString(p->p.Surname)
  .orderRev(e->e.values.size())
  .map(l->l.key+":"+l.values.size())
  .reduce((a,b)->a+", "+b)

        `, "persons")
		if err != nil {
			panic(err)
		}
		// Evaluate the function.
		result, err := fu.Eval(persons)
		if err != nil {
			panic(err)
		}
		fmt.Println(result)
	}

}
