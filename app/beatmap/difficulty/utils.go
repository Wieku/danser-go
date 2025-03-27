package difficulty

import "reflect"

func rfType[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

func parseConfig[T modSetting[T]](base T, config map[string]any) T {
	rVal := reflect.ValueOf(&base).Elem()
	rType := reflect.TypeOf(base)

	for i := range rType.NumField() {
		sField := rType.Field(i)
		if fTag, ok := sField.Tag.Lookup("json"); ok && fTag != "-" {
			if v, ok2 := config[fTag]; ok2 {
				sVal := reflect.ValueOf(v)

				if rVal.Field(i).CanInt() && !sVal.CanInt() && sVal.CanFloat() {
					rVal.Field(i).SetInt(int64(sVal.Float()))
				} else {
					rVal.Field(i).Set(sVal)
				}
			}
		}
	}

	return base.postLoad()
}

func exportConfig[T any](toExp T) map[string]any {
	expMap := make(map[string]any)

	rVal := reflect.ValueOf(toExp)
	rType := reflect.TypeOf(toExp)

	for i := range rType.NumField() {
		sField := rType.Field(i)
		if fTag, ok := sField.Tag.Lookup("json"); ok && fTag != "-" {
			expMap[fTag] = rVal.Field(i).Interface()
		}
	}

	return expMap
}

func GetModConfig[T any](diff *Difficulty) (T, bool) {
	if s, ok := diff.modSettings[rfType[T]()].(T); ok {
		return s, true
	}

	var ret T
	return ret, false
}

func tryExportConfig[T any](diff *Difficulty) map[string]any {
	if c, ok := GetModConfig[T](diff); ok {
		return exportConfig(c)
	}

	return nil
}

func SetModConfig[T any](diff *Difficulty, config T) {
	diff.modSettings[rfType[T]()] = config
}
