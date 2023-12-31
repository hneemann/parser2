# Parser #

A simple but still fast parser that is able to parse expressions. It is configurable 
to many use cases. It supports a generic value type. This allows a dynamic 
type system, which requires runtime type checking.

# Usage #

The folder _examples_ contains some simple examples of a float based expression 
parser and a bool based one. 

# In Memory Queries #

The package _value_  contains a more advanced example parser which also 
supports lists, maps and closures. 

Let there be a list of people as an example. This list should be stored in 
the server's memory:

``` Go
type Person struct {
	Name         string
	Surname      string
	PlaceOfBirth string
	Age          int
}
```

now we create some data to play with:

``` Go
var Persons = []Person{
	{"John", "Doe", "London", 23},
	{"Jane", "Doe", "London", 25},
	{"Bob", "Smith", "New York", 21},
	{"Frank", "Muller", "New York", 22},
	{"Mary", "Green", "Seattle", 21},
	{"Jake", "Muller", "Washington", 22},
}
```

The parser must somehow access the data. This could be done using reflection, but this way is more flexible and also 
faster, but a bit more cumbersome. We create a wrapper for the data:

``` Go
var PersonToMap = value.NewToMap[Person]().
	Attr("Name",         func(p Person) value.Value { return value.String(p.Name) }).
	Attr("Surname",      func(p Person) value.Value { return value.String(p.Surname) }).
	Attr("PlaceOfBirth", func(p Person) value.Value { return value.String(p.PlaceOfBirth) }).
	Attr("Age",          func(p Person) value.Value { return value.Int(p.Age) })
```

If you don't mind a small loss of performance or flexibility, you can simply write:

``` Go
var PersonToMap = value.NewToMapReflection[Person]()
```

Now let's do some operations on the data. At first we have to create the parser, and the list of persons:

``` Go
func main() {
	// Create a parser.
	parser := value.New()
	// Create a list to be used containing the persons.
	persons := value.NewListOfMaps[Person](PersonToMap, Persons)
```
Now we can make some queries on the data. Let's create a list of all names:

``` Go
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
```
The result will be a string containing all names separated by a comma and a space:

```
John, Jane, Bob, Frank, Mary, Jake
```

A more sophisticated example would be to create a list of all people that are older than 21 and live in New York:

``` Go
	fu, err := parser.Generate(`

persons
  .accept(p->p.PlaceOfBirth="New York" & p.Age>21)
  .map(e->e.Name+": "+e.Age)
  .reduce((a,b)->a+", "+b)

    `, "persons")
```

Results in

```
Frank: 22
```

Or find out, which surnames are used and how often, ordered by the number of people with that surname:

``` Go
	fu, err := parser.Generate(`

persons
  .groupByString(p->p.Surname)
  .orderRev(e->e.values.size())
  .map(l->l.key+":"+l.values.size())
  .reduce((a,b)->a+", "+b)

    `, "persons")
```
This results in:

``` 
Doe:2, Muller:2, Smith:1, Green:1
```


# Structure #

The parser first creates an abstract syntax tree (AST) which is than 
used to performe some optimizations, like evaluation of constants and 
so on. After that a function is created which can be used to evaluate 
the expression.

All these steps are highly customizable. 

The main idea is to create a function and then evaluate it multiple 
times to offset the cost of going through the process of creating an 
AST, optimizing it and creating a function, instead of simply calculate 
the result of the expression.   
