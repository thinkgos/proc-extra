package excel

import (
	"errors"
	"strings"

	"github.com/spf13/cast"
)

// Cell 解析tag后获取的参数
type Cell struct {
	Column string  // 列名: 如A,B,C,D,E
	Head   string  // 表头名称
	Width  float64 // 单元格列宽度
	Height float64 // 单元格行高(一行只会设置一次)
	Style  int     // 单元格样式, 取值(1: 居中, 2: 边框, 3: 居中+边框, 4: 居中+边框+字体)
}

// parseCellTag 解析单元格tag
// tag格式如下:
//
// 列名 表头 列宽 行高 样式
// A 卡号 40 20 3
// A 卡号 - - 3
// A 卡号
func parseCellTag(value string) (cell Cell, err error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return cell, errors.New("xlsx: need valid cell tag")
	}
	tags := strings.Split(value, " ")
	if len(tags) < 1 {
		return cell, errors.New("xlsx: need valid cell tag")
	}
	cell.Column = tags[0]
	if len(tags) > 1 {
		cell.Head = tags[1]
	}
	if len(tags) > 2 && tags[2] != "-" {
		cell.Width = cast.ToFloat64(tags[2])
	}
	if len(tags) > 3 && tags[3] != "-" {
		cell.Height = cast.ToFloat64(tags[3])
	}
	if len(tags) > 4 && tags[4] != "-" {
		cell.Style = cast.ToInt(tags[4])
	}
	return cell, nil
}
