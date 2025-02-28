package excel

// Config 选项配置
type Config struct {
	// 是否覆盖旧工作表数据, 默认: 追加
	overrideSheet bool
	// 是否覆盖旧行数据, 默认: 追加
	overrideRow bool
	// 使能表头输出, 默认: 无表头
	enableHeader bool
	// 自定义表头, enableHeader = true 有效, 默认: 空
	headers []string
	// 数据起始行(含表头, 如果无表头, 则为数据行)
	// title != nil, 如果 rowStart < 2 则 rowStart = 2, 即至少保留一行给抬头
	// 其它 rowStart = 实际值, 如果 rowStart < 1 则 rowStart = 1, 至少从第一行开始
	// 前面的行数将被合并做为抬头
	rowStart int
	// 抬头配置, 当不为nil时, 启用抬头.
	title *Title
	// 限制读取最大行数,
	// <= 0 不限制
	limitReadMaxLine int
}

func (c *Config) takeOptions(opts ...Option) {
	for _, opt := range opts {
		opt(c)
	}
	if c.title != nil {
		if c.rowStart < 2 {
			c.rowStart = 2
		}
	} else {
		if c.rowStart < 1 {
			c.rowStart = 1
		}
	}
}
