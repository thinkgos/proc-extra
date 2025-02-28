package excel

import (
	"github.com/xuri/excelize/v2"
)

var font = excelize.Font{
	Bold:      false,
	Italic:    false,
	Underline: "",
	Family:    "",
	Size:      20,
	Strike:    false,
	Color:     "#000000",
}

var alignment = excelize.Alignment{
	Horizontal:      "center",
	Indent:          1,
	JustifyLastLine: true,
	ReadingOrder:    0,
	RelativeIndent:  1,
	ShrinkToFit:     false,
	TextRotation:    0,
	Vertical:        "center",
	WrapText:        false,
}

var border = []excelize.Border{
	{Type: "left", Color: "000000", Style: 1},
	{Type: "top", Color: "000000", Style: 1},
	{Type: "bottom", Color: "000000", Style: 1},
	{Type: "right", Color: "000000", Style: 1},
}

func (e *File[T]) cellStyle(t int) int {
	var style excelize.Style
	switch t {
	case 1: // 居中
		style.Alignment = &alignment
	case 2: // 边框
		style.Border = border
	case 3: // 居中 + 边框
		style.Alignment = &alignment
		style.Border = border
	case 4: // 居中 + 边框 + 字体
		style.Alignment = &alignment
		style.Border = border
		style.Font = &font
	default: // 默认: 居中 + 无边框
		style.Alignment = &alignment
	}
	styleId, err := e.NewStyle(&style)
	if err != nil {
		return -1
	}
	return styleId
}
