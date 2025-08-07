package excel

// Title 抬头配置
type Title struct {
	height      float64 // 行高度, 默认: 20
	value       string  // 抬头数据: 默认: 无
	useTemplate bool    // 是否使用模板方式设置抬头
	data        any     // 模板数据
}

func NewTitle() *Title {
	return &Title{height: 20}
}

func (t *Title) Height() float64 {
	if t.height > 0 {
		return t.height
	}
	return 20
}

func (t *Title) SetHeight(height float64) *Title {
	t.height = height
	return t
}

func (t *Title) SetTitle(v string) *Title {
	t.value = v
	return t
}

// WithUseTemplate 使用模板方式设置抬头, data为模板数据
func (t *Title) WithUseTemplate(data any) *Title {
	t.useTemplate = true
	t.data = data
	return t
}
