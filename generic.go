//go:build go1.18

package di

import (
	"fmt"
	"reflect"
)

func RegisterInstance[T any](container Container, instance T, options ...InstanceRegistrationOption) {
	t := reflect.TypeOf((*T)(nil)).Elem()
	container.RegisterInstance(t, instance, options...)
}

func RegisterDynamic[T any](container Container, delegate func(Resolver) (T, error), options ...InstanceRegistrationOption) {
	t := reflect.TypeOf((*T)(nil)).Elem()

	container.RegisterDynamic(t, func(r Resolver) (any, error) {
		return delegate(r)
	}, options...)
}

func ReplaceDynamic[T any](container Container, delegate func(Resolver) (T, error), options ...InstanceRegistrationOption) {
	t := reflect.TypeOf((*T)(nil)).Elem()
	container.ReplaceDynamic(t, func(r Resolver) (any, error) {
		return delegate(r)
	}, options...)
}

// Resolve resolves the given type with the given resolver
func Resolve[T any](resolver Resolver) (T, error) {
	var zero T
	t := reflect.TypeOf((*T)(nil)).Elem()
	instance, err := resolver.Resolve(t)
	if err != nil {
		return zero, err
	}
	cast, err := cast[T](t, instance)
	if err != nil {
		return zero, err
	}
	return cast, nil
}

// ResolveByName resolves the given type with the resolver and name
func ResolveByName[T any](resolver Resolver, name string) (T, error) {
	var zero T
	t := reflect.TypeOf((*T)(nil)).Elem()
	instance, err := resolver.ResolveByName(t, name)
	if err != nil {
		return zero, err
	}
	cast, err := cast[T](t, instance)
	if err != nil {
		return zero, err
	}
	return cast, nil
}

func ResolveAll[T any](resolver Resolver) ([]T, error) {
	t := reflect.TypeOf((*T)(nil)).Elem()
	instances, err := resolver.ResolveAll(t)
	if err != nil {
		return nil, err
	}

	casts := []T{}
	for _, instance := range instances {
		cast, err := cast[T](t, instance)
		if err != nil {
			return nil, err
		}
		casts = append(casts, cast)
	}
	return casts, nil
}

func cast[T any](t reflect.Type, instance any) (T, error) {
	var zero T
	cast, ok := instance.(T)
	if !ok {
		return zero, fmt.Errorf("unable to cast instance to %s", t.String())
	}
	return cast, nil
}
