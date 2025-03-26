package excel

import (
	"errors"
	"reflect"

	"github.com/spf13/cast"
	"github.com/xuri/excelize/v2"
)

// Decode 反射出工作表到数据
// T 支持两种类型格式
//   - T为结构体切片
//   - T为[]string
func (e *File[T]) Decode(sheet string, opts ...Option) ([]T, error) {
	c := Config{}
	c.takeOptions(opts...)

	reflectValue := reflect.ValueOf([]T{})
	dataElemType := reflectValue.Type().Elem()
	isPtr := dataElemType.Kind() == reflect.Ptr
	if isPtr {
		dataElemType = dataElemType.Elem()
	}
	if dataElemType.Kind() != reflect.Struct &&
		(dataElemType.Kind() != reflect.Slice || dataElemType.Elem().Kind() != reflect.String) {
		return nil, errors.New("xlsx: slice element not a struct or []string")
	}

	// 获取工作表
	index, err := e.GetSheetIndex(sheet)
	if err != nil {
		return nil, err
	}
	if index < 0 {
		return nil, errors.New("xlsx: not found active sheet")
	}
	e.SetActiveSheet(index)

	// 数据起始行
	rowStart := c.rowStart
	if c.enableHeader {
		rowStart += 1 // skip header
	}

	rows, err := e.Rows(sheet)
	if err != nil {
		return nil, err
	}
	defer rows.Close() // nolint: errcheck
	for totalRows := 1; rows.Next(); totalRows++ {
		if totalRows < rowStart {
			continue
		}
		// 如果有限制, 则超出行数则跳过
		if c.limitReadMaxLine > 0 && totalRows >= (c.limitReadMaxLine+rowStart) {
			break
		}
		line, err := rows.Columns()
		if err != nil {
			return nil, err
		}

		if dataElemType.Kind() == reflect.Struct {
			elem := reflect.New(dataElemType)
			// scan to struct
			elem, err = scanIntoStruct(elem, line)
			if err != nil {
				return nil, err
			}
			if isPtr {
				elem = elem.Addr()
			}
			reflectValue = reflect.Append(reflectValue, elem)
		} else {
			reflectValue = reflect.Append(reflectValue, reflect.ValueOf(line))
		}
	}
	return reflectValue.Interface().([]T), nil
}

func scanIntoStruct(values reflect.Value, line []string) (reflect.Value, error) {
	vv := values
	for vv.Kind() == reflect.Ptr {
		vv = vv.Elem()
	}
	for colIdx, t := 0, vv.Type(); colIdx < t.NumField(); colIdx++ {
		field := t.Field(colIdx)
		if !field.IsExported() {
			continue
		}
		tag := field.Tag.Get("xlsx")
		if tag == "-" {
			continue
		}
		cellDefine, err := getCellDefine(t)
		if err != nil {
			return vv, err
		}
		fieldCellDefine := cellDefine.field[field.Name]
		cell := fieldCellDefine.cell
		colIndex, err := excelize.ColumnNameToNumber(cell.Column)
		if err != nil {
			return vv, err
		}
		index := colIndex - 1
		if index >= len(line) {
			continue
		}
		fieldValue := vv.Field(colIdx)
		val := line[index]
		switch field.Type.Kind() { // nolint: exhaustive
		case reflect.String:
			fieldValue.SetString(val)
		case reflect.Bool:
			fieldValue.SetBool(cast.ToBool(val))
		case reflect.Int,
			reflect.Int8,
			reflect.Int16,
			reflect.Int32,
			reflect.Int64:
			fieldValue.SetInt(cast.ToInt64(val))
		case reflect.Uint,
			reflect.Uint8,
			reflect.Uint16,
			reflect.Uint32,
			reflect.Uint64:
			fieldValue.SetUint(cast.ToUint64(val))
		case reflect.Float32,
			reflect.Float64:
			fieldValue.SetFloat(cast.ToFloat64(val))
		default:
			continue
		}
	}
	return vv, nil
}
