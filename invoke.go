package di

import (
	"fmt"
	"reflect"
)

func Invoke(resolver Resolver, delegate any) (any, error) {
	t := reflect.TypeOf(delegate)
	err := validateDelegateType(resolver, t)
	if err != nil {
		return nil, err
	}
	parameters, err := resolveParameters(resolver, t)
	if err != nil {
		return nil, err
	}
	constructorValue := reflect.ValueOf(delegate)
	results := constructorValue.Call(parameters)
	if len(results) == 0 {
		return nil, nil
	}
	var instance any
	if !results[0].IsZero() {
		instance = results[0].Interface()
	} else {
		instance = reflect.Zero(results[0].Type()).Interface()
	}

	if len(results) == 2 {
		if !results[1].IsZero() {
			err = results[1].Interface().(error)
		}
	}
	return instance, err
}

func validateDelegateType(r Resolver, t reflect.Type) error {
	if t.Kind() != reflect.Func {
		return fmt.Errorf("function '%s' must be a method", t.Elem())
	}
	return nil
}

func resolveParameters(resolver Resolver, t reflect.Type) ([]reflect.Value, error) {
	// build up the parameter list
	inCount := t.NumIn()
	values := []reflect.Value{}
	for i := 0; i < inCount; i++ {
		parameterType := t.In(i)
		if parameterType.Kind() == reflect.Array || parameterType.Kind() == reflect.Slice {

			// is the function variadic and is this the last parameter?
			if t.IsVariadic() && i == inCount-1 {
				valueArray, err := resolver.ResolveAll(parameterType.Elem())
				if err != nil {
					return nil, err
				}
				for _, v := range valueArray {
					values = append(values, reflect.ValueOf(v))
				}
			} else {
				slice, err := resolveSlice(resolver, parameterType)
				if err != nil {
					return nil, err
				}
				values = append(values, slice)
			}
		} else if parameterType.Kind() == reflect.Map && parameterType.Key().Kind() == reflect.String {
			mapValue, err := resolveMap(resolver, parameterType.Elem())
			if err != nil {
				return nil, err
			}
			values = append(values, mapValue)
		} else {
			value, err := resolver.Resolve(parameterType)
			if err != nil {
				return nil, err
			}
			values = append(values, reflect.ValueOf(value))
		}
	}
	return values, nil
}

func resolveSlice(resolver Resolver, t reflect.Type) (reflect.Value, error) {
	var zero reflect.Value
	valueArray, err := resolver.ResolveAll(t.Elem())
	if err != nil {
		return zero, err
	}

	// make the slice the right size
	slice := reflect.MakeSlice(t, len(valueArray), len(valueArray))

	// set indexes of the slice
	for i, value := range valueArray {
		ptr := slice.Index(i)
		ptr.Set(reflect.ValueOf(value))
	}

	return slice, nil
}

func resolveMap(resolver Resolver, t reflect.Type) (reflect.Value, error) {
	var zero reflect.Value
	m, err := resolver.ResolveMap(t)
	if err != nil {
		return zero, err
	}

	keyType := reflect.TypeOf((*string)(nil)).Elem()
	valueType := t
	mapType := reflect.MapOf(keyType, valueType)
	mapValue := reflect.MakeMap(mapType)

	for k, v := range m {
		mapValue.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v))
	}
	return mapValue, nil
}
