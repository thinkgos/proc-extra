package excel

import (
	"strings"
)

// Title 抬头配置
type Title struct {
	height          float64                // 行高度, 默认: 20
	customValueFunc func() (string, error) // 抬头数据: 默认: 无
	builder         strings.Builder        // 动态增加
}

func NewTitle() *Title {
	return &Title{height: 20, builder: strings.Builder{}}
}

func (t *Title) Height() float64 {
	if t.height > 0 {
		return t.height
	}
	return 20
}

func (t *Title) Title() (string, error) {
	if t.customValueFunc != nil {
		return t.customValueFunc()
	} else {
		return t.builder.String(), nil
	}
}

func (t *Title) SetHeight(height float64) *Title {
	t.height = height
	return t
}

// AddKvLine 以#开头, 格式: # {{key}}: {{value}}
func (t *Title) AddKvLine(k, v string) *Title {
	if t.builder.Len() > 0 {
		t.builder.WriteString("\n")
	}
	t.builder.WriteString("# ")
	t.builder.WriteString(k)
	t.builder.WriteString(": ")
	t.builder.WriteString(v)
	return t
}

// AddLine 以#开头, 格式: # {{s}}
func (t *Title) AddLine(s string) *Title {
	if t.builder.Len() > 0 {
		t.builder.WriteString("\n")
	}
	t.builder.WriteString("# ")
	t.builder.WriteString(s)
	return t
}

// AddRawLine 格式: {{s}}
func (t *Title) AddRawLine(s string) *Title {
	if t.builder.Len() > 0 {
		t.builder.WriteString("\n")
	}
	t.builder.WriteString(s)
	return t
}

func (t *Title) SetCustomValueFunc(f func() (string, error)) *Title {
	t.customValueFunc = f
	return t
}
