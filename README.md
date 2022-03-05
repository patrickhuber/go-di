# di
a go runtime dependency injection framework

## features

* Supports lifetimes of static and per request
* Constructors injection with dependency resolution of parameters
* Constructor injection supports error return types with 
* Constructor injection supports multiple instances of same interface type
* Constructor injection supports array and multi-variate parameters 
* Constructor injection supports map[string]type resolution for registrations WithName

## getting started

```bash
go get github.com/patrickhuber/go-di@latest
```

## usage

Define a type and interface that will be used in registration

```golang
import(
  "log"
  "fmt"
  "github.com/patrickhuber/go-di"
)

// Namer
type Namer interface{
  Name() string
}

// the type of the namer interface. Defining the type like this makes using the container look much cleaner.
var NamerType = reflect.TypeOf((*Namer)(nil)).Elem()

// Person represents an implementation of the Namer interface
type Person struct{
  name string
}

// NewPerson returns a person with the given name
func NewPerson(name string) Namer{
  return &Person{
    name: name,
  }
}

// Name implements the Namer interface
func (p *Person) Name() string{
  return p.name
}
```

Create the container and register the type. A variable will be used to hold the type information making it easier to use

```golang
// create the container
container := di.NewContainer()
person := NewPerson("james")

// register the concrete type as a Namer interface.
container.RegisterInstance(NamerType, person)

// get the implementation for NamerType, instance is an interface{} so it must be cast
instance, err := container.Resolve(NamerType)
if err != nil{
  log.Fatal(err)
}

// cast
namer, ok := instance.(Namer)
if !ok{
  log.Fatalf("the resolved instance was not a Namer")
}
fmt.Println("The name is %s", namer.Name())
```

## examples

See the [unit tests](container_test.go) for more examples. 
See the [generic unit tests](generic_test.go) for examples that use generics.