package di

import "reflect"

func GetType(instance any) reflect.Type {
	return reflect.TypeOf(instance).Elem()
}
