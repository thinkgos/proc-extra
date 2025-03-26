package excel

import "github.com/xuri/excelize/v2"

func (e *File[T]) getSheetIndex(sheet string, overrideSheet bool) (int, error) {
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

func (e *File[T]) getSheetRows(sheet string, overrideRow bool) (int, error) {
	if overrideRow {
		return 0, nil
	}
	rows, err := e.Rows(sheet)
	if err != nil {
		return 0, err
	}
	defer rows.Close()// nolint: errcheck
	totalRows := 0
	for rows.Next() {
		totalRows++
	}
	return totalRows, nil
}

func (e *File[T]) setTile(sheet string, tt *Title, rowStart, colNum int) error {
	height := tt.Height()
	title, err := tt.Title()
	if err != nil {
		return err
	}
	err = e.SetRowHeight(sheet, 1, height)
	if err != nil {
		return err
	}
	col, err := excelize.ColumnNumberToName(colNum)
	if err != nil {
		return err
	}
	vcell, err := excelize.JoinCellName(col, rowStart-1)
	if err != nil {
		return err
	}
	// 合并单元格
	err = e.MergeCell(sheet, "A1", vcell)
	if err != nil {
		return err
	}
	style, err := e.NewStyle(&excelize.Style{Alignment: &excelize.Alignment{WrapText: true}})
	if err != nil {
		return err
	}
	err = e.SetCellStyle(sheet, "A1", vcell, style)
	if err != nil {
		return err
	}
	return e.SetCellStr(sheet, "A1", title)
}
