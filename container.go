package di

import (
	"errors"
	"fmt"
	"reflect"
)

type Lifetime int

const (
	LifetimeStatic     Lifetime = 0
	LifetimePerRequest Lifetime = 1
)

var (
	ErrNotExist = errors.New("item does not exist")
)

// Container represents a dependency injection container
type Container interface {
	// RegisterInstance registers a type with a single instace with the given registration options
	RegisterInstance(t reflect.Type, instance interface{}, options ...InstanceRegistrationOption)

	// RegisterDynamic registers a type with a dynamic resolver
	RegisterDynamic(t reflect.Type, delegate FuncResolver, options ...InstanceRegistrationOption)

	// RegisterConstructor registers a type dynamically by instpecting the constructor signature
	RegisterConstructor(constructor interface{}, options ...InstanceRegistrationOption) error

	// Resolver is required as a Container must allow resolution
	Resolver
}

type FuncResolver func(Resolver) (interface{}, error)

type container struct {
	data  map[string][]FuncResolver
	cache map[string][]interface{}
	// nameLookup looks up by [type][name][index] where index is the position in the data[type][] array
	nameLookup     map[string]map[string]int
	defaultOptions []DefaultRegistrationOption
}

type InstanceRegistrationOption func(*container, reflect.Type)
type DefaultRegistrationOption func(*container, reflect.Type)

// WithLifetime sets the lifetime of the registration
func WithLifetime(lifetime Lifetime) InstanceRegistrationOption {
	return withLifetime(lifetime)
}

// WithDefaultLifetime sets the lifetime of the registration
func WithDefaultLifetime(lifetime Lifetime) DefaultRegistrationOption {
	return withLifetime(lifetime)
}

func withLifetime(lifetime Lifetime) func(c *container, t reflect.Type) {
	return func(c *container, t reflect.Type) {
		// if the lifetime is per request, make sure to clear any static lifetimes that were set
		if lifetime == LifetimePerRequest {
			delete(c.cache, t.String())
		}
		// if the lifetime is static, set the cache key to cache the first invocation
		if lifetime == LifetimeStatic {
			c.cache[t.String()] = nil
		}
	}
}

func WithKey(key string) InstanceRegistrationOption {
	return func(c *container, t reflect.Type) {
		nameToIndex, ok := c.nameLookup[t.String()]
		if !ok {
			nameToIndex = map[string]int{}
			c.nameLookup[t.String()] = nameToIndex
		}

		index := len(c.data[t.String()]) - 1
		nameToIndex[key] = index
	}
}

// NewContainer returns a new container with the specified default options applied to all objects registered in the container
func NewContainer(options ...DefaultRegistrationOption) Container {

	return &container{
		data:           map[string][]FuncResolver{},
		cache:          map[string][]interface{}{},
		nameLookup:     map[string]map[string]int{},
		defaultOptions: options,
	}
}

func (c *container) RegisterConstructor(constructor interface{}, options ...InstanceRegistrationOption) error {
	t := reflect.TypeOf(constructor)
	if t.Kind() != reflect.Func {
		return fmt.Errorf("constructor '%s' must be a method", t.Elem())
	}

	outCount := t.NumOut()
	if outCount == 0 {
		return fmt.Errorf("constructor must have a return value and optional error")
	}
	returnType := t.Out(0)
	if outCount == 2 {
		errorType := t.Out(1)
		if !errorType.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			return fmt.Errorf("if a constructor has two parameters, the second must implement error")
		}
	} else if outCount != 1 {
		return fmt.Errorf("constructor must have a return value and optional error")
	}

	delegate := func(r Resolver) (interface{}, error) {
		inCount := t.NumIn()
		values := []reflect.Value{}
		for i := 0; i < inCount; i++ {
			parameterType := t.In(i)
			if parameterType.Kind() == reflect.Array || parameterType.Kind() == reflect.Slice {
				valueArray, err := r.ResolveAll(parameterType.Elem())
				if err != nil {
					return nil, err
				}
				// is the function variadic and is this the last parameter?
				if t.IsVariadic() && i == inCount-1 {
					for _, v := range valueArray {
						values = append(values, reflect.ValueOf(v))
					}
				} else {
					slice := reflect.MakeSlice(parameterType, 0, 0)
					for i := 0; i < len(valueArray); i++ {
						slice = reflect.Append(slice, reflect.ValueOf(valueArray[i]))
					}
					values = append(values, slice)
				}
			} else {
				value, err := r.Resolve(parameterType)
				if err != nil {
					return nil, err
				}
				values = append(values, reflect.ValueOf(value))
			}
		}
		constructorValue := reflect.ValueOf(constructor)
		results := constructorValue.Call(values)
		if len(results) == 0 {
			return nil, fmt.Errorf("no result while executing constructor '%s'", t.String())
		}
		var instance interface{}
		if !results[0].IsNil() {
			instance = results[0].Interface()
		}
		var err error = nil
		if len(results) == 2 {
			if !results[1].IsNil() {
				err = results[1].Interface().(error)
			}
		}
		return instance, err
	}
	c.RegisterDynamic(returnType, delegate, options...)
	return nil
}

func (c *container) RegisterDynamic(t reflect.Type, delegate FuncResolver, options ...InstanceRegistrationOption) {
	delegates, ok := c.data[t.String()]
	if !ok {
		delegates = []FuncResolver{}
	}
	delegates = append(delegates, delegate)
	c.data[t.String()] = delegates

	// apply the default options
	for _, option := range c.defaultOptions {
		option(c, t)
	}
	// apply the override options
	for _, option := range options {
		option(c, t)
	}
}

func (c *container) RegisterInstance(t reflect.Type, instance interface{}, options ...InstanceRegistrationOption) {
	c.RegisterDynamic(t, func(r Resolver) (interface{}, error) {
		return instance, nil
	}, options...)
}

func (c *container) Resolve(t reflect.Type) (interface{}, error) {
	results, err := c.ResolveAll(t)
	if err != nil {
		return nil, err
	}
	return results[0], nil
}

func (c *container) ResolveByName(t reflect.Type, name string) (interface{}, error) {
	results, err := c.ResolveAll(t)
	if err != nil {
		return nil, err
	}
	indexMap, typeExists := c.nameLookup[t.String()]
	if !typeExists {
		return nil, ErrNotExist
	}
	index, nameExists := indexMap[name]
	if !nameExists {
		return nil, ErrNotExist
	}
	if index >= len(results) {
		return nil, ErrNotExist
	}
	return results[index], nil
}

func (c *container) ResolveAll(t reflect.Type) ([]interface{}, error) {
	cached, shouldCache := c.cache[t.String()]
	isCached := cached != nil
	if shouldCache && isCached {
		return cached, nil
	}

	delegates, ok := c.data[t.String()]
	if !ok {
		return nil, fmt.Errorf("type %s not found", t.String())
	}
	if len(delegates) == 0 {
		return nil, fmt.Errorf("type %s not found", t.String())
	}
	results := []interface{}{}
	for _, d := range delegates {
		result, err := d(c)
		if err != nil {
			return nil, err
		}
		results = append(results, result)
	}
	if shouldCache && !isCached {
		c.cache[t.String()] = results
	}
	return results, nil
}
