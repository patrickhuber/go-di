package di_test

import (
	"fmt"
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/patrickhuber/go-di"
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

var _ = Describe("Container", func() {
	It("can resolve type", func() {
		container := di.NewContainer()
		sample := NewSample("test")
		container.RegisterInstance(SampleInterfaceType, sample)
		instance, err := container.Resolve(SampleInterfaceType)
		Expect(err).To(BeNil())
		Expect(instance).ToNot(BeNil())
		_, ok := instance.(SampleInterface)
		Expect(ok).To(BeTrue())
	})
	It("can register dynamic", func() {
		container := di.NewContainer()
		name := "myname"
		container.RegisterDynamic(StringType, func(r di.Resolver) (interface{}, error) {
			return name, nil
		})

		instance, err := container.Resolve(StringType)
		Expect(err).To(BeNil())
		Expect(instance).ToNot(BeNil())

		value, ok := instance.(string)
		Expect(ok).To(BeTrue())
		Expect(value).To(Equal(name))

	})
	Context("Constructor", func() {
		It("can register constructor", func() {
			container := di.NewContainer()
			name := "myname"

			container.RegisterInstance(StringType, name)
			err := container.RegisterConstructor(NewSample)
			Expect(err).To(BeNil())

			instance, err := container.Resolve(SampleInterfaceType)
			Expect(err).To(BeNil())
			Expect(instance).ToNot(BeNil())

			sample, ok := instance.(SampleInterface)
			Expect(ok).To(BeTrue())
			Expect(sample.Name()).To(Equal(name))
		})
		It("can register array parameter", func() {
			container := di.NewContainer()
			dependencies := []*SampleStruct{
				{name: "sample 1"},
				{name: "sample 2"},
			}
			container.RegisterInstance(DependencyInterfaceType, dependencies[0])
			container.RegisterInstance(DependencyInterfaceType, dependencies[1])
			err := container.RegisterConstructor(NewAggregate)
			Expect(err).To(BeNil())

			instance, err := container.Resolve(AggregateInterfaceType)
			Expect(err).To(BeNil())
			Expect(instance).ToNot(BeNil())
			a, ok := instance.(AggregateInterface)
			Expect(ok).To(BeTrue())
			Expect(a).ToNot(BeNil())
			Expect(len(a.Names())).To(Equal(2))
		})
		It("can register variadic parameter", func() {
			container := di.NewContainer()
			dependencies := []*SampleStruct{
				{name: "sample 1"},
				{name: "sample 2"},
			}
			container.RegisterInstance(DependencyInterfaceType, dependencies[0])
			container.RegisterInstance(DependencyInterfaceType, dependencies[1])
			err := container.RegisterConstructor(NewVariadic)
			Expect(err).To(BeNil())

			instance, err := container.Resolve(AggregateInterfaceType)
			Expect(err).To(BeNil())
			Expect(instance).ToNot(BeNil())
			_, ok := instance.(AggregateInterface)
			Expect(ok).To(BeTrue())
		})
		It("can register map parameter", func() {
			container := di.NewContainer()
			dependencies := []*SampleStruct{
				{name: "sample 1"},
				{name: "sample 2"},
			}
			for _, d := range dependencies {
				container.RegisterInstance(DependencyInterfaceType, d, di.WithName(d.Name()))
			}
			err := container.RegisterConstructor(NewMap)
			Expect(err).To(BeNil())

			instance, err := container.Resolve(MapInterfaceType)
			Expect(err).To(BeNil())
			Expect(instance).ToNot(BeNil())

			mapInstance, ok := instance.(MapInterface)
			Expect(ok).To(BeTrue())
			Expect(len(mapInstance.Keys())).To(Equal(2))
		})
		It("returns empty map when no keys specified", func() {
			container := di.NewContainer()
			dependencies := []*SampleStruct{
				{name: "sample 1"},
				{name: "sample 2"},
			}
			for _, d := range dependencies {
				container.RegisterInstance(DependencyInterfaceType, d)
			}
			err := container.RegisterConstructor(NewMap)
			Expect(err).To(BeNil())

			instance, err := container.Resolve(MapInterfaceType)
			Expect(err).To(BeNil())
			Expect(instance).ToNot(BeNil())

			mapInstance, ok := instance.(MapInterface)
			Expect(ok).To(BeTrue())
			Expect(len(mapInstance.Keys())).To(Equal(0))
		})
		It("can invoke constructor that returns error", func() {
			container := di.NewContainer()

			err := container.RegisterConstructor(NewWithError)
			Expect(err).To(BeNil())

			i, err := container.Resolve(SampleInterfaceType)
			Expect(err).ToNot(BeNil())
			Expect(i).To(BeNil())
		})
		It("can invoke constructor that returns value and nil error", func() {
			container := di.NewContainer()

			err := container.RegisterConstructor(NewWithNilError)
			Expect(err).To(BeNil())

			i, err := container.Resolve(SampleInterfaceType)
			Expect(err).To(BeNil())
			Expect(i).ToNot(BeNil())

			_, ok := i.(SampleInterface)
			Expect(ok).To(BeTrue())
		})
		It("throws error when second return type is not error", func() {
			container := di.NewContainer()
			err := container.RegisterConstructor(TwoReturnTypes)
			Expect(err).ToNot(BeNil())
		})
		It("throws error when no return type", func() {
			container := di.NewContainer()
			err := container.RegisterConstructor(func() {})
			Expect(err).ToNot(BeNil())
		})
	})
	It("can resolve all", func() {
		container := di.NewContainer()
		container.RegisterInstance(SampleInterfaceType, NewSample("one"))
		container.RegisterInstance(SampleInterfaceType, NewSample("two"))
		all, err := container.ResolveAll(SampleInterfaceType)
		Expect(err).To(BeNil())
		Expect(len(all)).To(Equal(2))
	})
	It("can resolve map", func() {
		container := di.NewContainer()
		keys := []string{"one", "two"}
		for _, key := range keys {
			container.RegisterInstance(SampleInterfaceType, NewSample(key), di.WithName(key))
		}
		m, err := container.ResolveMap(SampleInterfaceType)
		Expect(err).To(BeNil())
		Expect(len(m)).To(Equal(len(keys)))
	})
	Context("key", func() {
		It("can resolve by key", func() {
			container := di.NewContainer()
			container.RegisterInstance(SampleInterfaceType, NewSample("one"), di.WithName("one"))
			container.RegisterInstance(SampleInterfaceType, NewSample("two"), di.WithName("two"))
			instance, err := container.ResolveByName(SampleInterfaceType, "two")
			Expect(err).To(BeNil())
			Expect(instance).ToNot(BeNil())
		})
	})
	Context("lifetime", func() {
		var (
			container di.Container
		)
		It("can register default lifetime", func() {
			container = di.NewContainer(di.WithDefaultLifetime(di.LifetimeStatic))
			err := container.RegisterConstructor(NewStorage)
			Expect(err).To(BeNil())
		})
		It("can register lifetime", func() {
			container = di.NewContainer()
			err := container.RegisterConstructor(NewStorage, di.WithLifetime(di.LifetimeStatic))
			Expect(err).To(BeNil())
		})
		It("can override default lifetime", func() {
			container = di.NewContainer(di.WithDefaultLifetime(di.LifetimePerRequest))
			err := container.RegisterConstructor(NewStorage, di.WithLifetime(di.LifetimeStatic))
			Expect(err).To(BeNil())
		})
		AfterEach(func() {
			obj, err := container.Resolve(StorageType)
			Expect(err).To(BeNil())

			storage, ok := obj.(Storage)
			Expect(ok).To(BeTrue())
			storage.Set(1, "test")

			obj, err = container.Resolve(StorageType)
			Expect(err).To(BeNil())

			storage, ok = obj.(Storage)
			Expect(ok).To(BeTrue())

			value := storage.Get(1)
			Expect(value).To(Equal("test"))
		})
	})
})
