package di

import "reflect"

// Resolver resolves an instance from a given type
type Resolver interface {

	// Resolve resolves the instace registered for a given type
	Resolve(t reflect.Type) (interface{}, error)

	// ResolveAll resolves all instances registered for the given type
	ResolveAll(t reflect.Type) ([]interface{}, error)
}
