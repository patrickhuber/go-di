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

var InjectedType = reflect.TypeOf((*Injected)(nil)).Elem()

func TestInject(t *testing.T) {
	injected := &injected{}
	container := di.NewContainer()
	container.RegisterInstance(InjectedType, injected)

	instance := &Wrapper{}
	err := di.Inject(container, instance)
	require.NoError(t, err)
	require.NotNil(t, instance.Injected)
	require.Nil(t, instance.NotInjected)
}
