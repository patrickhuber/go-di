package di

import "reflect"

type Resolver interface {
	Resolve(t reflect.Type) (interface{}, error)

	ResolveAll(t reflect.Type) ([]interface{}, error)
}
