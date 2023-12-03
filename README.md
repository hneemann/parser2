# Parser #

A simple parser that is able to parse expressions. It is configurable 
to many use cases. It supports a generic value type. This allows a dynamic 
type system, which requires runtime type checking.

# Examples #

The folder _examples_ contains some simple examples of a float based expression 
parser and a bool base one. The package _value_  contains a more advanced 
example parser which also supports lists, maps and closures.

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
