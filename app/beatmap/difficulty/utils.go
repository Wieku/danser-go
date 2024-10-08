package difficulty

import "reflect"

func rfType[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

func parseConfig[T any](base T, config map[string]any) T {
	rVal := reflect.ValueOf(&base).Elem()
	rType := reflect.TypeOf(base)

	for i := range rType.NumField() {
		sField := rType.Field(i)
		if fTag, ok := sField.Tag.Lookup("json"); ok {
			if v, ok2 := config[fTag]; ok2 {
				rVal.Field(i).Set(reflect.ValueOf(v))
			}
		}
	}

	return base
}
