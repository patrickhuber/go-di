package di

import (
	"reflect"
)

// Resolver resolves an instance from a given type
type Resolver interface {

	// Resolve resolves the instace registered for a given type
	Resolve(t reflect.Type) (any, error)

	// ResolveAll resolves all instances registered for the given type
	ResolveAll(t reflect.Type) ([]any, error)

	// ResolveMap resolves all named instances as a map
	ResolveMap(t reflect.Type) (map[string]any, error)

	// ResolveByName resolves the instance registered for a given type and name
	ResolveByName(t reflect.Type, name string) (any, error)
}
