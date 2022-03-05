package di_test

import (
	"reflect"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/patrickhuber/go-di"
)

type Injected interface {
}

type injected struct {
}

type Wrapper struct {
	Injected    Injected `inject:""`
	NotInjected *int
}

var InjectedType = reflect.TypeOf((*Injected)(nil)).Elem()

var _ = Describe("Inject", func() {
	It("can inject", func() {
		injected := &injected{}
		container := di.NewContainer()
		container.RegisterInstance(InjectedType, injected)

		instance := &Wrapper{}
		err := di.Inject(container, instance)
		Expect(err).To(BeNil())

		Expect(instance.Injected).ToNot(BeNil())
		Expect(instance.NotInjected).To(BeNil())
	})
})
