package di

import (
	"fmt"
	"reflect"
)

type Lifetime int

const (
	LifetimeStatic     Lifetime = 0
	LifetimePerRequest Lifetime = 1
)

type Container interface {
	RegisterInstance(t reflect.Type, instance interface{}, options ...RegistrationOption)
	RegisterDynamic(t reflect.Type, delegate FuncResolver, options ...RegistrationOption)
	RegisterConstructor(constructor interface{}, options ...RegistrationOption) error
	Resolver
}

type FuncResolver func(Resolver) (interface{}, error)

type container struct {
	data  map[string][]FuncResolver
	cache map[string][]interface{}
}

type RegistrationOption func(*container, reflect.Type)

func WithLifetime(lifetime Lifetime) RegistrationOption {
	return func(c *container, t reflect.Type) {
		if lifetime == LifetimeStatic {
			c.cache[t.String()] = nil
		}
	}
}

func NewContainer() Container {

	return &container{
		data:  map[string][]FuncResolver{},
		cache: map[string][]interface{}{},
	}
}

func (c *container) RegisterConstructor(constructor interface{}, options ...RegistrationOption) error {
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

func (c *container) RegisterDynamic(t reflect.Type, delegate FuncResolver, options ...RegistrationOption) {
	delegates, ok := c.data[t.String()]
	if !ok {
		delegates = []FuncResolver{}
	}
	delegates = append(delegates, delegate)
	c.data[t.String()] = delegates
	for _, option := range options {
		option(c, t)
	}
}

func (c *container) RegisterInstance(t reflect.Type, instance interface{}, options ...RegistrationOption) {
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
