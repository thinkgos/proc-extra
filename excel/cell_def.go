package excel

import (
	"strings"
)

// CellDef 单元格定义, 解析tag后获取的
type CellDef struct {
	Head       string // 表头名称
	tagOptions tagOptions
}

// parseCellDef 解析单元格定义
// tag格式如下:
//
// 表头 列宽 行高 样式
// 卡号 40 20 3
// 卡号 - - 3
// 卡号
func parseCellDef(tagName string) *CellDef {
	cell := &CellDef{}
	tagName = strings.TrimSpace(tagName)
	if tagName == "" {
		return cell
	}
	tags := strings.Split(tagName, " ")
	if len(tags) > 0 {
		cell.Head = tags[0]
	}
	return cell
}
