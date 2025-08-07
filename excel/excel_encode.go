package excel

import (
	"errors"
	"reflect"
	"slices"

	"github.com/xuri/excelize/v2"
)

// Encode 反射data数据到工作表
// data 支持两种类型格式
//   - T为结构体切片
//   - T为基础数据类型的切片或数组
//
// T为结构体的字段tag定义:
//
//	`xlsx:"卡号,omitempty"`
//	`xlsx:"卡号"`
func (e *File[T]) Encode(sheet string, data []T, opts ...Option) error {
	c := Config{}
	c.takeOptions(opts...)

	dataElemType := indirectType(reflect.TypeOf(data).Elem())
	if !slices.Contains([]reflect.Kind{reflect.Array, reflect.Slice, reflect.Struct}, dataElemType.Kind()) {
		return errors.New("xlsx: data element must be a struct, slice or array")
	}

	index, err := e.sheetIndex(sheet, c.overrideSheet)
	if err != nil {
		return err
	}
	e.SetActiveSheet(index)
	totalRows, err := e.sheetTotalRows(sheet, c.overrideRow)
	if err != nil {
		return err
	}

	if dataElemType.Kind() == reflect.Struct {
		err = e.encodeSliceStruct(sheet, totalRows, dataElemType, data, &c)
	} else {
		err = e.encodeMatrix(sheet, totalRows, dataElemType, data, &c)
	}
	return err
}

func (e *File[T]) encodeSliceStruct(sheet string, totalRows int, dataElemType reflect.Type, data []T, c *Config) (err error) {
	headers, err := getHeaders(dataElemType)
	if err != nil {
		return err
	}
	//* 设置抬头
	if c.title != nil {
		err = e.setTile(sheet, c.title, c.rowStart, len(headers))
		if err != nil {
			return err
		}
	}
	//* 设置表头
	if c.enableHeader {
		headerRow := c.rowStart
		headerRowValue := headers
		if len(c.customHeaders) > 0 {
			headerRowValue = c.customHeaders
		}
		headerAxis, err := excelize.JoinCellName("A", headerRow)
		if err != nil {
			return err
		}
		err = e.SetSheetRow(sheet, headerAxis, &headerRowValue)
		if err != nil {
			return err
		}
	}

	//* 数据起始行
	dataRowStart := c.rowStart
	if totalRows > 0 { // 有旧数据, 追加
		dataRowStart = totalRows + 1
	} else if c.enableHeader {
		dataRowStart += 1 // 跳过表头
	}
	//* 获取指定行的行高
	dataRowHeight := float64(0)
	if c.dataCellStyleBaseOnRow > 0 {
		dataRowHeight, err = e.GetRowHeight(sheet, c.dataCellStyleBaseOnRow)
		if err != nil {
			return err
		}
	}
	for rowIdx, rowValue := range data {
		//* 当前处理的行
		currentRow := dataRowStart + rowIdx
		//* 设置行高度, 一行只设置一次
		if dataRowHeight > 0 {
			if err = e.SetRowHeight(sheet, currentRow, dataRowHeight); err != nil {
				return err
			}
		}
		if e.transformRowValue != nil {
			rowValues, err := e.transformRowValue(rowValue)
			if err != nil {
				return err
			}
			for colIdx, cellValue := range rowValues {
				currentCol := colIdx + 1
				err = e.writeCell(sheet, currentRow, currentCol, cellValue, c)
				if err != nil {
					return err
				}
			}
		} else {
			vv := indirectValue(reflect.ValueOf(rowValue))
			tt := vv.Type()
			for currentCol, idx := 1, 0; idx < tt.NumField(); idx++ {
				field := tt.Field(idx)
				if !field.IsExported() {
					continue
				}
				tag := field.Tag.Get("xlsx")
				if tag == "-" {
					continue
				}
				_, tagOpts := parseTag(tag)
				if fieldValue := vv.Field(idx); tagOpts.Contains("omitempty") && isEmptyValue(fieldValue) {
					err = e.writeCell(sheet, currentRow, currentCol, "", c)
				} else {
					err = e.writeCell(sheet, currentRow, currentCol, fieldValue.Interface(), c)
				}
				if err != nil {
					return err
				}
				currentCol++
			}
		}
	}
	return nil
}

func (e *File[T]) encodeMatrix(sheet string, totalRows int, dataElemType reflect.Type, data []T, c *Config) (err error) {
	_ = dataElemType
	//* 设置抬头
	if c.title != nil {
		colNum := 0
		if len(data) > 0 {
			colNum = reflect.ValueOf(data[0]).Len()
		}
		err = e.setTile(sheet, c.title, c.rowStart, colNum)
		if err != nil {
			return err
		}
	}
	//* 设置表头
	if c.enableHeader && len(c.customHeaders) > 0 {
		// 表头行
		headerRow := c.rowStart
		headerAxis, err := excelize.JoinCellName("A", headerRow)
		if err != nil {
			return err
		}
		err = e.SetSheetRow(sheet, headerAxis, &c.customHeaders)
		if err != nil {
			return err
		}
	}

	if len(data) == 0 {
		return nil
	}

	//* 数据起始行
	dataRowStart := c.rowStart
	if totalRows > 0 { // 有旧数据, 追加
		dataRowStart = totalRows + 1
	} else if c.enableHeader {
		dataRowStart += 1 // 跳过表头
	}

	//* 获取指定行的行高
	dataRowHeight := float64(0)
	if c.dataCellStyleBaseOnRow > 0 {
		dataRowHeight, err = e.GetRowHeight(sheet, c.dataCellStyleBaseOnRow)
		if err != nil {
			return err
		}
	}
	for rowIdx, rowValue := range data {
		//* 当前处理的行
		currentRow := dataRowStart + rowIdx
		//* 设置行高度, 一行只设置一次
		if dataRowHeight > 0 {
			if err = e.SetRowHeight(sheet, currentRow, dataRowHeight); err != nil {
				return err
			}
		}
		if e.transformRowValue != nil {
			rowValues, err := e.transformRowValue(rowValue)
			if err != nil {
				return err
			}
			for colIdx, cellValue := range rowValues {
				currentCol := colIdx + 1
				err = e.writeCell(sheet, currentRow, currentCol, cellValue, c)
				if err != nil {
					return err
				}
			}
		} else {
			vv := indirectValue(reflect.ValueOf(rowValue))
			for colIdx := range vv.Len() {
				currentCol := colIdx + 1
				cellValue := vv.Index(colIdx).Interface()
				err = e.writeCell(sheet, currentRow, currentCol, cellValue, c)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

func (e *File[T]) writeCell(sheet string, row, col int, cellValue any, c *Config) error {
	colName, err := excelize.ColumnNumberToName(col)
	if err != nil {
		return err
	}
	axis, err := excelize.JoinCellName(colName, row)
	if err != nil {
		return err
	}
	if c.dataCellStyleBaseOnRow > 0 && c.dataCellStyleBaseOnRow != row {
		baseAxis, err := excelize.JoinCellName(colName, c.dataCellStyleBaseOnRow)
		if err != nil {
			return err
		}
		//* 获取基于指定行的单元格样式
		styleId, err := e.GetCellStyle(sheet, baseAxis)
		if err != nil {
			return err
		}
		//* 应用到当前单元格
		err = e.SetCellStyle(sheet, axis, axis, styleId)
		if err != nil {
			return err
		}
	}
	return e.SetCellValue(sheet, axis, cellValue)
}

func isEmptyValue(v reflect.Value) bool {
	switch k := v.Kind(); k { // nolint: exhaustive
	case reflect.Bool:
		return !v.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return v.Float() == 0
	case reflect.Array, reflect.Slice, reflect.Map, reflect.String, reflect.Chan:
		return v.Len() == 0
	case reflect.Interface, reflect.Ptr:
		return v.IsNil()
	}
	return false
}

func indirectValue(v reflect.Value) reflect.Value {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
}

func indirectType(v reflect.Type) reflect.Type {
	for v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	return v
}
