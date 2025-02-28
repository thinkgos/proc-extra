package excel

import (
	"reflect"
)

type cellDefine struct {
	cell    Cell
	options tagOptions
}

type fieldValue struct {
	fields []*cellDefine
	field  map[string]*cellDefine
}

var cacheCellDefine = map[reflect.Type]*fieldValue{}

func getCellDefine(t reflect.Type) (*fieldValue, error) {
	v, ok := cacheCellDefine[t]
	if ok {
		return v, nil
	}
	cachedValue := &fieldValue{
		fields: []*cellDefine{},
		field:  map[string]*cellDefine{},
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
		cell, err := parseCellTag(tagName)
		if err != nil {
			return nil, err
		}
		cd := &cellDefine{
			cell:    cell,
			options: tagOpts,
		}
		cachedValue.fields = append(cachedValue.fields, cd)
		cachedValue.field[field.Name] = cd
	}
	return cachedValue, nil
}
