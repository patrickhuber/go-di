package di_test

import (
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	"github.com/patrickhuber/go-di"
)

var _ = Describe("Invoke", func() {
	It("can invoke function", func() {
		container := di.NewContainer()
		container.RegisterInstance(StringType, "hello")
		container.RegisterInstance(SampleInterfaceType, NewSample("test"))
		myFunction := func(sample SampleInterface, greeting string) string {
			return greeting + " " + sample.Name()
		}
		result, err := di.Invoke(container, myFunction)
		Expect(err).To(BeNil())
		Expect(result).To(Equal("hello test"))
	})
	It("can return struct pointer", func() {
		container := di.NewContainer()
		container.RegisterInstance(SampleInterfaceType, NewSample("test"))
		myFunction := func(sample SampleInterface) *SampleStruct {
			return &SampleStruct{
				name: sample.Name(),
			}
		}
		result, err := di.Invoke(container, myFunction)
		Expect(err).To(BeNil())
		Expect(result).ToNot(BeNil())
		sample, ok := result.(*SampleStruct)
		Expect(ok).To(BeTrue())
		Expect(sample.Name()).To(Equal("test"))
	})
	It("can register struct pointer", func() {
		container := di.NewContainer()
		myFunction := func() *SampleStruct {
			return &SampleStruct{
				name: "test",
			}
		}
		container.RegisterConstructor(myFunction)
		_, err := di.Invoke(container, func(sample *SampleStruct) {
			Expect(sample).ToNot(BeNil())
			Expect(sample.Name()).To(Equal("test"))
		})
		Expect(err).To(BeNil())
	})
	It("can invoke array parameter", func() {})
	It("can invoke variadic parameter", func() {})
	It("can invoke map parameter", func() {})
	It("can invoke with error and value in return", func() {})
	It("can invoke with error in return", func() {})
	It("fails when error is not second return type", func() {})
	It("throws error with no return type", func() {})
})
