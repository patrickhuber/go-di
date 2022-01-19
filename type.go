package di

import "reflect"

func GetType(instance interface{}) reflect.Type {
	return reflect.TypeOf(instance).Elem()
}
