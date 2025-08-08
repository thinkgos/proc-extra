package excel

import (
	"strings"
	"text/template"

	"github.com/xuri/excelize/v2"
)

func (e *File[T]) sheetIndex(sheet string, overrideSheet bool) (int, error) {
	if !overrideSheet { // 覆盖工作表, 直接新建
		index, err := e.GetSheetIndex(sheet)
		if err != nil {
			return 0, err
		}
		if index >= 0 {
			return index, nil
		}
		// 未找到表, 直接新建
	}
	return e.NewSheet(sheet)
}

func (e *File[T]) sheetTotalRows(sheet string, overrideRow bool) (int, error) {
	if overrideRow {
		return 0, nil
	}
	rows, err := e.Rows(sheet)
	if err != nil {
		return 0, err
	}
	defer rows.Close() // nolint: errcheck
	totalRows := 0
	for rows.Next() {
		totalRows++
	}
	return totalRows, nil
}

func (e *File[T]) GetCellDirectlyStyle(sheet, cell string) (*excelize.Style, error) {
	styleId, err := e.GetCellStyle(sheet, cell)
	if err != nil {
		return nil, err
	}
	return e.GetStyle(styleId)
}

func (e *File[T]) setTile(sheet string, tt *Title, rowStart, colNum int) error {
	if tt.useTemplate {
		value, err := e.GetCellValue(sheet, "A1")
		if err != nil {
			return err
		}
		tpl, err := template.New("title").Parse(value)
		if err != nil {
			return err
		}
		v := &strings.Builder{}
		err = tpl.Execute(v, tt.data)
		if err != nil {
			return err
		}
		return e.SetCellStr(sheet, "A1", v.String())
	} else {
		err := e.SetRowHeight(sheet, 1, tt.rowHeight)
		if err != nil {
			return err
		}
		if tt.colNum > 0 {
			colNum = tt.colNum
		}
		vcell := "A1"
		if colNum > 0 {
			col, err := excelize.ColumnNumberToName(colNum)
			if err != nil {
				return err
			}
			vcell, err = excelize.JoinCellName(col, rowStart-1)
			if err != nil {
				return err
			}
			// 合并单元格
			err = e.MergeCell(sheet, "A1", vcell)
			if err != nil {
				return err
			}
		}
		if tt.style != nil {
			styleId, err := e.NewStyle(tt.style)
			if err != nil {
				return err
			}
			err = e.SetCellStyle(sheet, "A1", vcell, styleId)
			if err != nil {
				return err
			}
		}
		return e.SetCellStr(sheet, "A1", tt.value)
	}
}
