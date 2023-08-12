package di

import "reflect"

func Inject(resolver Resolver, instance any) error {
	t := reflect.TypeOf(instance).Elem()
	v := reflect.ValueOf(instance).Elem()

	count := t.NumField()
	for i := 0; i < count; i++ {
		field := t.Field(i)
		_, ok := field.Tag.Lookup("inject")
		if !ok {
			continue
		}
		fieldValue := v.FieldByName(field.Name)
		if !fieldValue.IsValid() || !fieldValue.CanAddr() || !fieldValue.CanSet() {
			continue
		}
		resolved, err := resolver.Resolve(field.Type)
		if err != nil {
			return err
		}
		fieldValue.Set(reflect.ValueOf(resolved))
	}
	return nil
}
