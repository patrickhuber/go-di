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
	ErrNotExist     = errors.New("item does not exist in the container")
	ErrNameNotExist = errors.New("item with the given name does not exist in the container")
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

type registrationOption struct {
	name     string
	key      string
	resolver FuncResolver
	lifetime Lifetime
}

type containerItem struct {
	data   interface{}
	err    error
	option *registrationOption
}

func (i *containerItem) resolve(r Resolver) (interface{}, error) {

	// was the error cached?
	if i.err != nil {
		return nil, i.err
	}

	// was the data cached?
	if i.data != nil {
		return i.data, nil
	}

	// execute the resolver
	data, err := i.option.resolver(r)

	// if static lifetime, cache the results
	if i.option.lifetime == LifetimeStatic {
		i.data = data
		i.err = err
	}

	return data, err
}

// containerItemGroup holds a group of container items
type containerItemGroup struct {
	items      []*containerItem
	namedItems map[string]*containerItem
}

type container struct {
	groups         map[string]*containerItemGroup
	defaultOptions []DefaultRegistrationOption
}

type InstanceRegistrationOption func(*registrationOption)
type DefaultRegistrationOption func(*registrationOption)

// WithLifetime sets the lifetime of the registration
func WithLifetime(lifetime Lifetime) InstanceRegistrationOption {
	return withLifetime(lifetime)
}

// WithDefaultLifetime sets the lifetime of the registration
func WithDefaultLifetime(lifetime Lifetime) DefaultRegistrationOption {
	return withLifetime(lifetime)
}

func withLifetime(lifetime Lifetime) func(i *registrationOption) {
	return func(i *registrationOption) {
		i.lifetime = lifetime
	}
}

func WithName(name string) InstanceRegistrationOption {
	return func(i *registrationOption) {
		i.name = name
	}
}

// NewContainer returns a new container with the specified default options applied to all objects registered in the container
func NewContainer(options ...DefaultRegistrationOption) Container {

	return &container{
		groups:         map[string]*containerItemGroup{},
		defaultOptions: options,
	}
}

func (c *container) RegisterConstructor(constructor interface{}, options ...InstanceRegistrationOption) error {
	t := reflect.TypeOf(constructor)
	err := validateDelegateTypeIsConstructor(c, t)
	if err != nil {
		return err
	}

	delegate := func(r Resolver) (interface{}, error) {
		return Invoke(r, constructor)
	}

	returnType := t.Out(0)
	c.RegisterDynamic(returnType, delegate, options...)
	return nil
}

func validateDelegateTypeIsConstructor(r Resolver, t reflect.Type) error {
	err := validateDelegateType(r, t)
	if err != nil {
		return err
	}
	outCount := t.NumOut()
	if outCount == 0 {
		return fmt.Errorf("function must have a return value and optional error")
	} else if outCount == 2 {
		errorType := t.Out(1)
		if !errorType.Implements(reflect.TypeOf((*error)(nil)).Elem()) {
			return fmt.Errorf("if a function has two return parameters, the second must implement error")
		}
	} else if outCount != 1 {
		return fmt.Errorf("function must have a return value and optional error")
	}
	return nil
}

func (c *container) RegisterDynamic(t reflect.Type, delegate FuncResolver, options ...InstanceRegistrationOption) {
	// try to find the existing container item group
	key := t.String()

	o := &registrationOption{
		key:      key,
		resolver: delegate,
	}

	// apply the default options
	for _, option := range c.defaultOptions {
		option(o)
	}

	// apply the override options
	for _, option := range options {
		option(o)
	}

	group, ok := c.groups[key]
	if !ok {
		group = &containerItemGroup{
			items:      []*containerItem{},
			namedItems: map[string]*containerItem{},
		}
		c.groups[key] = group
	}

	item := &containerItem{
		option: o,
	}

	// if the name is empty, append to the list of unnamed items
	if o.name == "" {
		group.items = append(group.items, item)
	} else {
		group.namedItems[o.name] = item
	}
}

func (c *container) RegisterInstance(t reflect.Type, instance interface{}, options ...InstanceRegistrationOption) {
	c.RegisterDynamic(t, func(r Resolver) (interface{}, error) {
		return instance, nil
	}, options...)
}

func (c *container) group(t reflect.Type) (*containerItemGroup, error) {
	key := t.String()
	group, ok := c.groups[key]
	if !ok {
		return nil, fmt.Errorf("%w: '%s'", ErrNotExist, key)
	}
	return group, nil
}

func (c *container) Resolve(t reflect.Type) (interface{}, error) {
	results, err := c.ResolveAll(t)
	if err != nil {
		return nil, err
	}
	return results[0], nil
}

func (c *container) ResolveByName(t reflect.Type, name string) (interface{}, error) {
	group, err := c.group(t)
	if err != nil {
		return nil, err
	}
	item, ok := group.namedItems[name]
	if !ok {
		return nil, fmt.Errorf("%w: '%s'", ErrNameNotExist, name)
	}
	return item.resolve(c)
}

func (c *container) ResolveAll(t reflect.Type) ([]interface{}, error) {
	group, err := c.group(t)
	if err != nil {
		return nil, err
	}

	// loop over the group named instances and collect
	var all []interface{}
	for _, v := range group.namedItems {
		data, err := v.resolve(c)
		if err != nil {
			return nil, err
		}
		all = append(all, data)
	}
	// loop over regular instances and collect
	for _, v := range group.items {
		data, err := v.resolve(c)
		if err != nil {
			return nil, err
		}
		all = append(all, data)
	}
	return all, nil
}

func (c *container) ResolveMap(t reflect.Type) (map[string]interface{}, error) {
	group, err := c.group(t)
	if err != nil {
		return nil, err
	}

	result := map[string]interface{}{}
	for k, v := range group.namedItems {
		data, err := v.resolve(c)
		if err != nil {
			return nil, err
		}
		result[k] = data
	}
	return result, nil
}
