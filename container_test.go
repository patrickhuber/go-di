package di_test

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/patrickhuber/go-di"
	"github.com/stretchr/testify/require"
)

type SampleStruct struct {
	name string
}

func (s *SampleStruct) Name() string {
	return s.name
}

type AggregateStruct struct {
	names []string
}

func (a *AggregateStruct) Names() []string {
	return a.names
}

type SampleInterface interface {
	Name() string
}

func NewSample(name string) SampleInterface {
	return &SampleStruct{
		name: name,
	}
}

type DependencyInterface interface {
	Name() string
}

type AggregateInterface interface {
	Names() []string
}

func NewVariadic(dependencies ...DependencyInterface) AggregateInterface {
	names := []string{}
	for _, d := range dependencies {
		names = append(names, d.Name())
	}
	return &AggregateStruct{
		names: names,
	}
}

func NewAggregate(dependencies []DependencyInterface) AggregateInterface {
	names := []string{}
	for _, d := range dependencies {
		names = append(names, d.Name())
	}
	return &AggregateStruct{
		names: names,
	}
}

type MapInterface interface {
	Keys() []string
	Lookup(name string) (DependencyInterface, bool)
}

type MapStruct struct {
	items map[string]DependencyInterface
}

func (m *MapStruct) Keys() []string {
	var keys []string
	for k := range m.items {
		keys = append(keys, k)
	}
	return keys
}

func (m *MapStruct) Lookup(key string) (DependencyInterface, bool) {
	value, ok := m.items[key]
	return value, ok
}

func NewMap(dependencies map[string]DependencyInterface) MapInterface {
	items := map[string]DependencyInterface{}
	for k, v := range dependencies {
		items[k] = v
	}
	return &MapStruct{
		items: items,
	}
}

func NewWithNilError() (SampleInterface, error) {
	return &SampleStruct{
		name: "test",
	}, nil
}

func NewWithError() (SampleInterface, error) {
	return nil, fmt.Errorf("this is an error")
}

func TwoReturnTypes() (SampleInterface, AggregateInterface) {
	return nil, nil
}

type Storage interface {
	Get(id int) string
	Set(id int, value string)
}

type storage struct {
	data map[int]string
}

func NewStorage() Storage {
	return &storage{
		data: map[int]string{},
	}
}

func (s *storage) Get(id int) string {
	value, ok := s.data[id]
	if !ok {
		return ""
	}
	return value
}

func (s *storage) Set(id int, value string) {
	s.data[id] = value
}

var StringType = reflect.TypeOf((*string)(nil)).Elem()
var SampleInterfaceType = reflect.TypeOf((*SampleInterface)(nil)).Elem()
var DependencyInterfaceType = reflect.TypeOf((*DependencyInterface)(nil)).Elem()
var AggregateInterfaceType = reflect.TypeOf((*AggregateInterface)(nil)).Elem()
var MapInterfaceType = reflect.TypeOf((*MapInterface)(nil)).Elem()
var StorageType = reflect.TypeOf((*Storage)(nil)).Elem()

func TestContainer(t *testing.T) {
	t.Run("resolve type", func(t *testing.T) {
		container := di.NewContainer()
		sample := NewSample("test")
		container.RegisterInstance(SampleInterfaceType, sample)
		instance, err := container.Resolve(SampleInterfaceType)
		require.NoError(t, err)
		require.NotNil(t, instance)
		_, ok := instance.(SampleInterface)
		require.True(t, ok)
	})
	t.Run("dynamic", func(t *testing.T) {
		container := di.NewContainer()
		name := "myname"
		container.RegisterDynamic(StringType, func(r di.Resolver) (any, error) {
			return name, nil
		})

		instance, err := container.Resolve(StringType)
		require.NoError(t, err)
		require.NotNil(t, instance)

		value, ok := instance.(string)
		require.True(t, ok)
		require.Equal(t, name, value)
	})
}

func TestConstructor(t *testing.T) {
	t.Run("register", func(t *testing.T) {
		container := di.NewContainer()
		name := "myname"

		container.RegisterInstance(StringType, name)
		err := container.RegisterConstructor(NewSample)
		require.NoError(t, err)

		instance, err := container.Resolve(SampleInterfaceType)
		require.NoError(t, err)
		require.NotNil(t, instance)

		sample, ok := instance.(SampleInterface)
		require.True(t, ok)
		require.Equal(t, name, sample.Name())
	})
	t.Run("array parameter", func(t *testing.T) {
		container := di.NewContainer()
		dependencies := []*SampleStruct{
			{name: "sample 1"},
			{name: "sample 2"},
		}
		container.RegisterInstance(DependencyInterfaceType, dependencies[0])
		container.RegisterInstance(DependencyInterfaceType, dependencies[1])
		err := container.RegisterConstructor(NewAggregate)
		require.NoError(t, err)

		instance, err := container.Resolve(AggregateInterfaceType)
		require.NoError(t, err)
		require.NotNil(t, instance)
		a, ok := instance.(AggregateInterface)
		require.True(t, ok)
		require.NotNil(t, a)
		require.Equal(t, 2, len(a.Names()))
	})
	t.Run("variadic", func(t *testing.T) {
		container := di.NewContainer()
		dependencies := []*SampleStruct{
			{name: "sample 1"},
			{name: "sample 2"},
		}
		container.RegisterInstance(DependencyInterfaceType, dependencies[0])
		container.RegisterInstance(DependencyInterfaceType, dependencies[1])
		err := container.RegisterConstructor(NewVariadic)
		require.NoError(t, err)

		instance, err := container.Resolve(AggregateInterfaceType)
		require.NoError(t, err)
		require.NotNil(t, instance)
		_, ok := instance.(AggregateInterface)
		require.True(t, ok)
	})
	t.Run("func", func(t *testing.T) {
		container := di.NewContainer()
		dependencies := []*SampleStruct{
			{name: "sample 1"},
			{name: "sample 2"},
		}
		for _, d := range dependencies {
			container.RegisterInstance(DependencyInterfaceType, d, di.WithName(d.Name()))
		}
		err := container.RegisterConstructor(NewMap)
		require.NoError(t, err)

		instance, err := container.Resolve(MapInterfaceType)
		require.NoError(t, err)
		require.NotNil(t, instance)

		mapInstance, ok := instance.(MapInterface)
		require.True(t, ok)
		require.Equal(t, 2, len(mapInstance.Keys()))
	})
	t.Run("no keys", func(t *testing.T) {
		container := di.NewContainer()
		dependencies := []*SampleStruct{
			{name: "sample 1"},
			{name: "sample 2"},
		}
		for _, d := range dependencies {
			container.RegisterInstance(DependencyInterfaceType, d)
		}
		err := container.RegisterConstructor(NewMap)
		require.NoError(t, err)

		instance, err := container.Resolve(MapInterfaceType)
		require.NoError(t, err)
		require.NotNil(t, instance)

		mapInstance, ok := instance.(MapInterface)
		require.True(t, ok)
		require.Equal(t, 0, len(mapInstance.Keys()))
	})
	t.Run("mix keyed and unkeyed", func(t *testing.T) {
		container := di.NewContainer()
		dependencies := []*SampleStruct{
			{name: "sample 1"},
			{name: "sample 2"},
		}
		for _, d := range dependencies {
			container.RegisterInstance(DependencyInterfaceType, d, di.WithName(d.Name()))
		}
		const sample3Name = "sample 3"
		container.RegisterInstance(DependencyInterfaceType, &SampleStruct{name: sample3Name})
		err := container.RegisterConstructor(NewMap)
		require.NoError(t, err)

		instance, err := container.Resolve(MapInterfaceType)
		require.NoError(t, err)
		require.NotNil(t, instance)

		mapInstance, ok := instance.(MapInterface)
		require.True(t, ok)
		require.Equal(t, 2, len(mapInstance.Keys()))

		singleInstance, err := container.Resolve(DependencyInterfaceType)
		require.NoError(t, err)
		require.NotNil(t, singleInstance)

		sample3, ok := singleInstance.(DependencyInterface)
		require.True(t, ok)
		require.Equal(t, sample3Name, sample3.Name())
	})
	t.Run("err ret", func(t *testing.T) {
		container := di.NewContainer()

		err := container.RegisterConstructor(NewWithError)
		require.NoError(t, err)

		i, err := container.Resolve(SampleInterfaceType)
		require.NotNil(t, err)
		require.Nil(t, i)
	})
	t.Run("value err ret", func(t *testing.T) {

		container := di.NewContainer()

		err := container.RegisterConstructor(NewWithNilError)
		require.NoError(t, err)

		i, err := container.Resolve(SampleInterfaceType)
		require.NoError(t, err)
		require.NotNil(t, i)

		_, ok := i.(SampleInterface)
		require.True(t, ok)
	})
	t.Run("error must be last", func(t *testing.T) {
		container := di.NewContainer()
		err := container.RegisterConstructor(TwoReturnTypes)
		require.NotNil(t, err)
	})
	t.Run("must have return type", func(t *testing.T) {
		container := di.NewContainer()
		err := container.RegisterConstructor(func() {})
		require.NotNil(t, err)
	})
	t.Run("resolve all", func(t *testing.T) {
		container := di.NewContainer()
		container.RegisterInstance(SampleInterfaceType, NewSample("one"))
		container.RegisterInstance(SampleInterfaceType, NewSample("two"))
		all, err := container.ResolveAll(SampleInterfaceType)
		require.NoError(t, err)
		require.Equal(t, 2, len(all))
	})
	t.Run("resolve map", func(t *testing.T) {
		container := di.NewContainer()
		keys := []string{"one", "two"}
		for _, key := range keys {
			container.RegisterInstance(SampleInterfaceType, NewSample(key), di.WithName(key))
		}
		m, err := container.ResolveMap(SampleInterfaceType)
		require.NoError(t, err)
		require.Equal(t, len(keys), len(m))
	})
	t.Run("resolve by key", func(t *testing.T) {
		container := di.NewContainer()
		container.RegisterInstance(SampleInterfaceType, NewSample("one"), di.WithName("one"))
		container.RegisterInstance(SampleInterfaceType, NewSample("two"), di.WithName("two"))
		instance, err := container.ResolveByName(SampleInterfaceType, "two")
		require.NoError(t, err)
		require.NotNil(t, instance)
	})
	t.Run("lifetime", func(t *testing.T) {
		type test struct {
			name      string
			container di.Container
		}
		tests := []test{
			{"can register default lifetime",
				func() di.Container {
					container := di.NewContainer(di.WithDefaultLifetime(di.LifetimeStatic))
					err := container.RegisterConstructor(NewStorage)
					require.NoError(t, err)
					return container
				}()},
			{"can register lifetime",
				func() di.Container {
					container := di.NewContainer()
					err := container.RegisterConstructor(NewStorage, di.WithLifetime(di.LifetimeStatic))
					require.NoError(t, err)
					return container
				}()},
			{"can override default lifetime", func() di.Container {
				container := di.NewContainer(di.WithDefaultLifetime(di.LifetimePerRequest))
				err := container.RegisterConstructor(NewStorage, di.WithLifetime(di.LifetimeStatic))
				require.NoError(t, err)
				return container
			}()},
		}
		for _, test := range tests {
			obj, err := test.container.Resolve(StorageType)
			require.NoError(t, err)

			storage, ok := obj.(Storage)
			require.True(t, ok)
			storage.Set(1, "test")

			obj, err = test.container.Resolve(StorageType)
			require.NoError(t, err)

			storage, ok = obj.(Storage)
			require.True(t, ok)

			value := storage.Get(1)
			require.Equal(t, "test", value)
		}
	})
	t.Run("remove all", func(t *testing.T) {
		names := []string{"one", "two", "three"}
		container := di.NewContainer()
		for _, name := range names {
			name := name
			container.RegisterInstance(
				SampleInterfaceType,
				NewSample(name),
				di.WithName(name))
		}

		all, err := container.ResolveAll(SampleInterfaceType)
		require.NoError(t, err)
		require.Equal(t, 3, len(all))

		container.RemoveAll(SampleInterfaceType)

		_, err = container.ResolveAll(SampleInterfaceType)
		require.Error(t, err)
	})
	t.Run("replace", func(t *testing.T) {
		names := []string{"one", "two", "three"}
		container := di.NewContainer()
		for _, name := range names {
			name := name
			container.RegisterInstance(
				SampleInterfaceType,
				NewSample(name),
				di.WithName(name))
		}

		all, err := container.ResolveAll(SampleInterfaceType)
		require.NoError(t, err)
		require.Equal(t, 3, len(all))
		container.ReplaceInstance(SampleInterfaceType, NewSample("four"))

		all, err = container.ResolveAll(SampleInterfaceType)
		require.NoError(t, err)
		require.Equal(t, 1, len(all))
	})
}
