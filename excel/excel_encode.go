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
// T为结构体的结构tag定义
// tag格式(除了列名, 其它均可忽略):
//
//	列名 表头 列宽 行高 样式(1：居中, 2：边框, 3：居中+边框, 4: 居中+边框+字体)
//	`xlsx:"A 卡号 40 20 3,omitempty"`
//	`xlsx:"A 卡号 - - 3,omitempty"`
//	`xlsx:"A 卡号"`
func (e *File[T]) Encode(sheet string, data []T, opts ...Option) error {
	c := Config{}
	c.takeOptions(opts...)

	dataElemType := indirectType(reflect.TypeOf(data).Elem())
	if !slices.Contains([]reflect.Kind{reflect.Array, reflect.Slice, reflect.Struct}, dataElemType.Kind()) {
		return errors.New("xlsx: data element must be a struct, slice or array")
	}

	index, err := e.getSheetIndex(sheet, c.overrideSheet)
	if err != nil {
		return err
	}
	e.SetActiveSheet(index)
	totalRows, err := e.getSheetRows(sheet, c.overrideRow)
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
	cellDefine, err := getCellDefine(dataElemType)
	if err != nil {
		return err
	}
	//* 设置抬头
	if totalRows == 0 && c.title != nil {
		err = e.setTile(sheet, c.title, c.rowStart, len(cellDefine.fields))
		if err != nil {
			return err
		}
	}

	// 数据起始行
	rowStart := c.rowStart
	if totalRows > 0 { // 有旧数据, 追加
		rowStart = totalRows + 1
	} else {
		if c.enableHeader {
			rowStart += 1 // skip header
		}
	}

	//* 设置表头
	// 仅工作表无数据时, 才需要设置列宽和表头
	if totalRows == 0 {
		if len(c.headers) > 0 {
			if c.enableHeader {
				axis, err := excelize.JoinCellName("A", rowStart-1)
				if err != nil {
					return err
				}
				err = e.SetSheetRow(sheet, axis, &c.headers)
				if err != nil {
					return err
				}
			}
		} else {
			for colIdx, v := range cellDefine.fields {
				cell := v.cell
				//* 设置列宽
				if cell.Width > 0 {
					err = e.SetColWidth(sheet, cell.Column, cell.Column, cell.Width)
					if err != nil {
						return err
					}
				}
				//* 设置表头
				if c.enableHeader {
					axisTitle, err := excelize.JoinCellName(cell.Column, rowStart-1)
					if err != nil {
						return err
					}
					//* 设置表头名称
					err = e.SetCellValue(sheet, axisTitle, cell.Head)
					if err != nil {
						return err
					}
					//* 设置行高
					// 一行只设置一次, 由第一个元素决定
					if colIdx == 0 {
						err = e.SetRowHeight(sheet, rowStart-1, cell.Height)
						if err != nil {
							return err
						}
					}
					//* 设置表头样式
					if cell.Style > 0 {
						if style := e.cellStyle(cell.Style); style > 0 {
							err = e.SetCellStyle(sheet, axisTitle, axisTitle, style)
							if err != nil {
								return err
							}
						}
					}
				}
			}
		}
	}
	for rowIdx, value := range data {
		if e.transformRowValue != nil {
			rowValue, err := e.transformRowValue(value)
			if err != nil {
				return err
			}
			axis, err := excelize.JoinCellName("A", rowStart+rowIdx)
			if err != nil {
				return err
			}
			err = e.SetSheetRow(sheet, axis, &rowValue)
			if err != nil {
				return err
			}
		} else {
			vv := indirectValue(reflect.ValueOf(value))
			for colIdx, t := 0, vv.Type(); colIdx < t.NumField(); colIdx++ {
				field := t.Field(colIdx)
				if !field.IsExported() {
					continue
				}
				tag := field.Tag.Get("xlsx")
				if tag == "-" {
					continue
				}
				fieldCellDefine := cellDefine.field[field.Name]
				cell := fieldCellDefine.cell
				tagOpts := fieldCellDefine.options

				// 行
				currentRow := rowStart + rowIdx
				// 设置行高度, 一行只设置一次
				if colIdx == 0 && cell.Height > 0 {
					if err = e.SetRowHeight(sheet, currentRow, cell.Height); err != nil {
						return err
					}
				}
				axis, err := excelize.JoinCellName(cell.Column, currentRow)
				if err != nil {
					return err
				}
				// 设置单元格样式
				if cell.Style > 0 {
					if style := e.cellStyle(cell.Style); style > 0 {
						err = e.SetCellStyle(sheet, axis, axis, style)
						if err != nil {
							return err
						}
					}
				}
				fieldValue := vv.Field(colIdx)

				if !tagOpts.Contains("omitempty") || !isEmptyValue(fieldValue) {
					err = e.SetCellValue(sheet, axis, fieldValue.Interface())
					if err != nil {
						return err
					}
				}
			}
		}
	}
	return nil
}

func (e *File[T]) encodeMatrix(sheet string, totalRows int, dataElemType reflect.Type, data []T, c *Config) (err error) {
	_ = dataElemType
	if len(data) == 0 {
		return nil
	}

	//* 设置抬头
	if totalRows == 0 && c.title != nil {
		colNum := reflect.ValueOf(data[0]).Len()
		err = e.setTile(sheet, c.title, c.rowStart, colNum)
		if err != nil {
			return err
		}
	}

	// 数据起始行
	rowStart := c.rowStart
	if totalRows > 0 { // 有旧数据, 追加
		rowStart = totalRows + 1
	} else {
		if c.enableHeader {
			rowStart += 1 // skip header
		}
	}
	//* 设置表头
	// 仅工作表无数据时, 才需要设置列宽和表头
	if totalRows == 0 && c.enableHeader && len(c.headers) > 0 {
		axis, err := excelize.JoinCellName("A", rowStart-1)
		if err != nil {
			return err
		}
		err = e.SetSheetRow(sheet, axis, &c.headers)
		if err != nil {
			return err
		}
	}

	for rowIdx, value := range data {
		axis, err := excelize.JoinCellName("A", rowStart+rowIdx)
		if err != nil {
			return err
		}
		if e.transformRowValue != nil {
			rowValue, err := e.transformRowValue(value)
			if err != nil {
				return err
			}
			err = e.SetSheetRow(sheet, axis, &rowValue)
			if err != nil {
				return err
			}
		} else {
			err = e.SetSheetRow(sheet, axis, &value)
			if err != nil {
				return err
			}
		}
	}
	return nil
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
