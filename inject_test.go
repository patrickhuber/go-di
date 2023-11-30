package di_test

import (
	"reflect"
	"testing"

	"github.com/patrickhuber/go-di"
	"github.com/stretchr/testify/require"
)

type Injected interface {
}

type injected struct {
}

type Wrapper struct {
	Injected    Injected `inject:""`
	NotInjected *int
}

type Parent struct {
	Child Child `inject:""`
}

type Child struct {
	Something string
}

var InjectedType = reflect.TypeOf((*Injected)(nil)).Elem()
var ChildType = reflect.TypeOf((*Child)(nil)).Elem()
var ParentType = reflect.TypeOf((*Parent)(nil)).Elem()

func TestInject(t *testing.T) {
	t.Run("interface child", func(t *testing.T) {
		injected := &injected{}
		container := di.NewContainer()
		container.RegisterInstance(InjectedType, injected)

		instance := &Wrapper{}
		err := di.Inject(container, instance)
		require.NoError(t, err)
		require.NotNil(t, instance.Injected)
		require.Nil(t, instance.NotInjected)
	})
	t.Run("value type child", func(t *testing.T) {
		child := Child{
			Something: "something",
		}

		container := di.NewContainer()
		container.RegisterInstance(ChildType, child)

		parent := Parent{}
		err := di.Inject(container, &parent)
		require.NoError(t, err)
		require.Equal(t, "something", parent.Child.Something)
	})
}
