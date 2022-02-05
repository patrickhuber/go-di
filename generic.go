//go:build go1.18

package di

import (
	"fmt"
	"reflect"
)

func RegisterInstance[T any](container Container, instance T, options ...RegistrationOption) InstanceRegistration {
	t := reflect.TypeOf((*T)(nil)).Elem()
	return container.RegisterInstance(t, instance, options...)
}

func RegisterDynamic[T any](container Container, delegate func(Resolver) (T, error), options ...RegistrationOption) InstanceRegistration {
	t := reflect.TypeOf((*T)(nil)).Elem()

	return container.RegisterDynamic(t, func(r Resolver) (interface{}, error) {
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
	cast, ok := instance.(T)
	if !ok {
		return zero, fmt.Errorf("Unable to cast instance to %s", t.String())
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
	cast, ok := instance.(T)
	if !ok {
		return zero, fmt.Errorf("Unable to cast instance to %s", t.String())
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
		cast, ok := instance.(T)
		if !ok {
			return nil, fmt.Errorf("Unable to cast instance to %s", t.String())
		}
		casts = append(casts, cast)
	}
	return casts, nil
}
