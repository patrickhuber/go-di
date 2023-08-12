package di_test

import (
	"testing"

	"github.com/patrickhuber/go-di"
	"github.com/stretchr/testify/require"
)

func TestInvoke(t *testing.T) {
	t.Run("function", func(t *testing.T) {
		container := di.NewContainer()
		container.RegisterInstance(StringType, "hello")
		container.RegisterInstance(SampleInterfaceType, NewSample("test"))
		myFunction := func(sample SampleInterface, greeting string) string {
			return greeting + " " + sample.Name()
		}
		result, err := di.Invoke(container, myFunction)
		require.NoError(t, err)
		require.Equal(t, "hello test", result)
	})
	t.Run("struct pointer", func(t *testing.T) {
		container := di.NewContainer()
		container.RegisterInstance(SampleInterfaceType, NewSample("test"))
		myFunction := func(sample SampleInterface) *SampleStruct {
			return &SampleStruct{
				name: sample.Name(),
			}
		}
		result, err := di.Invoke(container, myFunction)
		require.NoError(t, err)
		require.NotNil(t, result)
		sample, ok := result.(*SampleStruct)
		require.True(t, ok)
		require.Equal(t, "test", sample.name)
	})
	t.Run("register struct pointer", func(t *testing.T) {
		container := di.NewContainer()
		myFunction := func() *SampleStruct {
			return &SampleStruct{
				name: "test",
			}
		}
		container.RegisterConstructor(myFunction)
		_, err := di.Invoke(container, func(sample *SampleStruct) {
			require.NotNil(t, sample)
			require.Equal(t, "test", sample.Name())
		})
		require.NoError(t, err)
	})
	t.Run("can invoke array parameter", func(t *testing.T) {})
	t.Run("can invoke variadic parameter", func(t *testing.T) {})
	t.Run("can invoke map parameter", func(t *testing.T) {})
	t.Run("can invoke with error and value in return", func(t *testing.T) {})
	t.Run("can invoke with error in return", func(t *testing.T) {})
	t.Run("fails when error is not second return type", func(t *testing.T) {})
	t.Run("throws error with no return type", func(t *testing.T) {})
}
