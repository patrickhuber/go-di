//go:build go1.18

package di_test

import (
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/patrickhuber/go-di"
)

type Runner interface {
	Run()
}

func NewRunner() Runner {
	return &runner{}
}

type runner struct{}

func (r *runner) Run() {
}

var RunnerType = reflect.TypeOf((*Runner)(nil)).Elem()

var _ = Describe("Generic", func() {
	It("can register instance", func() {
		container := di.NewContainer()
		runner := NewRunner()
		di.RegisterInstance(container, runner)
		instance, err := container.Resolve(RunnerType)
		Expect(err).To(BeNil())
		r, ok := instance.(Runner)
		Expect(ok).To(BeTrue())
		Expect(r).ToNot(BeNil())
	})
	It("can register dynamic", func() {
		container := di.NewContainer()
		resolver := func(r di.Resolver) (Runner, error) {
			return &runner{}, nil
		}
		di.RegisterDynamic(container, resolver)
		instance, err := container.Resolve(RunnerType)
		Expect(err).To(BeNil())
		r, ok := instance.(Runner)
		Expect(ok).To(BeTrue())
		Expect(r).ToNot(BeNil())
	})
	It("can resolve", func() {
		container := di.NewContainer()
		runner := NewRunner()
		container.RegisterInstance(RunnerType, runner)
		instance, err := di.Resolve[Runner](container)
		Expect(err).To(BeNil())
		Expect(instance).ToNot(BeNil())
	})
	It("can resolve all", func() {
		container := di.NewContainer()
		runner1 := NewRunner()
		runner2 := NewRunner()
		container.RegisterInstance(RunnerType, runner1)
		container.RegisterInstance(RunnerType, runner2)
		instances, err := di.ResolveAll[Runner](container)
		Expect(err).To(BeNil())
		Expect(len(instances)).To(Equal(2))
	})
	It("can resolve by name", func() {
		container := di.NewContainer()
		runner1 := NewRunner()
		runner2 := NewRunner()

		container.RegisterInstance(RunnerType, runner1).WithKey("runner1")
		container.RegisterInstance(RunnerType, runner2).WithKey("runner2")
		instance, err := di.ResolveByName[Runner](container, "runner1")

		Expect(err).To(BeNil())
		Expect(instance).ToNot(BeNil())
	})
	It("can register name", func() {

		container := di.NewContainer()
		runner := NewRunner()
		di.RegisterInstance(container, runner).WithKey("runner1")
		instance, err := di.ResolveByName[Runner](container, "runner1")

		Expect(err).To(BeNil())
		Expect(instance).ToNot(BeNil())
	})
})
