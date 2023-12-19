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

var People = []Person{
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
	parser := value.SetUpParser(value.New())
	// Create a list to be used containing the people.
	people := value.NewListConvert(func(p Person) value.Value { return PersonToMap.Create(p) }, People)
	{
		// Create a function that evaluates the list of people.
		// The argument 'people' is passed to the function.
		fu, err := parser.Generate(`

people.map(p->p.Name).reduce((a,b)->a+", "+b)

        `, "people")
		if err != nil {
			panic(err)
		}
		// Evaluate the function.
		result, err := fu.Eval(people)
		if err != nil {
			panic(err)
		}
		fmt.Println(result)
	}

	{
		fu, err := parser.Generate(`

people
  .accept(p->p.PlaceOfBirth="New York" & p.Age>21)
  .map(e->e.Name+": "+e.Age)
  .reduce((a,b)->a+", "+b)

        `, "people")
		if err != nil {
			panic(err)
		}
		// Evaluate the function.
		result, err := fu.Eval(people)
		if err != nil {
			panic(err)
		}
		fmt.Println(result)
	}

	{
		fu, err := parser.Generate(`

people
  .groupByString(p->p.Surname)
  .orderRev(e->e.values.size())
  .map(l->l.key+":"+l.values.size())
  .reduce((a,b)->a+", "+b)

        `, "people")
		if err != nil {
			panic(err)
		}
		// Evaluate the function.
		result, err := fu.Eval(people)
		if err != nil {
			panic(err)
		}
		fmt.Println(result)
	}

}
