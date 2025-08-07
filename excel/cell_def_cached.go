package excel

import "reflect"

// TODO: 优化缓存, 目前没有考虑并发安全问题
var cacheCellDefine = map[reflect.Type]*fieldValue{}

type fieldValue struct {
	fields []*CellDef
	field  map[string]*CellDef
}

func getCellDefine(t reflect.Type) (*fieldValue, error) {
	v, ok := cacheCellDefine[t]
	if ok {
		return v, nil
	}
	cachedValue := &fieldValue{
		fields: []*CellDef{},
		field:  map[string]*CellDef{},
	}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		tag := field.Tag.Get("xlsx")
		if tag == "-" {
			continue
		}
		tagName, tagOpts := parseTag(tag)
		cell := &CellDef{
			Head:       tagName,
			tagOptions: tagOpts,
		}
		cachedValue.fields = append(cachedValue.fields, cell)
		cachedValue.field[field.Name] = cell
	}
	return cachedValue, nil
}
