//go:build go1.18

package di_test

import (
	"reflect"
	"testing"

	"github.com/patrickhuber/go-di"
	"github.com/stretchr/testify/require"
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

func TestGeneric(t *testing.T) {
	t.Run("can register instance", func(t *testing.T) {
		container := di.NewContainer()
		runner := NewRunner()
		di.RegisterInstance(container, runner)
		instance, err := container.Resolve(RunnerType)
		require.NoError(t, err)
		r, ok := instance.(Runner)
		require.True(t, ok)
		require.NotNil(t, r)
	})
	t.Run("can register dynamic", func(t *testing.T) {
		container := di.NewContainer()
		resolver := func(r di.Resolver) (Runner, error) {
			return &runner{}, nil
		}
		di.RegisterDynamic(container, resolver)
		instance, err := container.Resolve(RunnerType)
		require.NoError(t, err)
		r, ok := instance.(Runner)
		require.True(t, ok)
		require.NotNil(t, r)
	})
	t.Run("can resolve", func(t *testing.T) {
		container := di.NewContainer()
		runner := NewRunner()
		container.RegisterInstance(RunnerType, runner)
		instance, err := di.Resolve[Runner](container)
		require.NoError(t, err)
		require.NotNil(t, instance)
	})
	t.Run("can resolve all", func(t *testing.T) {
		container := di.NewContainer()
		runner1 := NewRunner()
		runner2 := NewRunner()
		container.RegisterInstance(RunnerType, runner1)
		container.RegisterInstance(RunnerType, runner2)
		instances, err := di.ResolveAll[Runner](container)
		require.NoError(t, err)
		require.Equal(t, 2, len(instances))
	})
	t.Run("can resolve by name", func(t *testing.T) {
		container := di.NewContainer()
		runner1 := NewRunner()
		runner2 := NewRunner()

		container.RegisterInstance(RunnerType, runner1, di.WithName("runner1"))
		container.RegisterInstance(RunnerType, runner2, di.WithName("runner2"))
		instance, err := di.ResolveByName[Runner](container, "runner1")

		require.NoError(t, err)
		require.NotNil(t, instance)
	})
	t.Run("can register name", func(t *testing.T) {

		container := di.NewContainer()
		runner := NewRunner()
		di.RegisterInstance(container, runner, di.WithName("runner1"))
		instance, err := di.ResolveByName[Runner](container, "runner1")

		require.NoError(t, err)
		require.NotNil(t, instance)
	})
}
