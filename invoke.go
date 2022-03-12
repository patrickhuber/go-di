package di

type Invoker interface {
	// Invoke the given function using the container to supply parameters
	Invoke(delegate interface{}) (interface{}, error)
}
