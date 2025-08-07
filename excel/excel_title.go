package excel

import "github.com/xuri/excelize/v2"

// Title 抬头配置
type Title struct {
	rowHeight   float64 // 行高度, 默认: 20
	colNum      int     // 合并列数, 默认: 0, 不合并
	style       *excelize.Style
	value       string // 抬头数据: 默认: 无
	useTemplate bool   // 是否使用模板, 默认: false
	data        any    // 模板数据
}

func NewTitle() *Title {
	return &Title{rowHeight: 20}
}

func (t *Title) SetRowHeight(v float64) *Title {
	if v > 0 {
		t.rowHeight = v
	}
	return t
}

func (t *Title) SetTitle(v string) *Title {
	t.value = v
	return t
}

// SetUseTemplate 使用模板方式设置抬头, data为模板数据
func (t *Title) SetUseTemplate(data any) *Title {
	t.useTemplate = true
	t.data = data
	return t
}

// SetColNum 设置合并列数
func (t *Title) SetColNum(v int) *Title {
	t.colNum = v
	return t
}

func (t *Title) SetStyle(v *excelize.Style) *Title {
	t.style = v
	return t
}

func (t *Title) BuildOption() Option {
	return func(c *Config) {
		c.title = t
	}
}
