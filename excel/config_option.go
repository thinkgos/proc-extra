package excel

// Option config option
type Option func(*Config)

// WithOverrideSheet 覆盖旧工作表数据, 默认: 追加
func WithOverrideSheet() Option {
	return func(c *Config) {
		c.overrideSheet = true
	}
}

// WithOverrideRow 覆盖旧行数据, 默认: 追加
func WithOverrideRow() Option {
	return func(c *Config) {
		c.overrideRow = true
	}
}

// WithEnableHeader 使能表头输出
func WithEnableHeader() Option {
	return func(c *Config) {
		c.enableHeader = true
	}
}

// WithEnableHeader 使能表头输出
func WithHeaders(headers []string) Option {
	return func(c *Config) {
		c.headers = headers
	}
}

// WithRowStart 输出的起始行(可含表头, 如果无表头, 则为数据行)
// title != nil, 如果 rowStart < 2 则 rowStart = 2, 即至少保留一行给抬头
// 其它 rowStart = 实际值, 如果 rowStart < 1 则 rowStart = 1, 至少从第一行开始
func WithRowStart(rowStart int) Option {
	return func(c *Config) {
		c.rowStart = rowStart
	}
}

// WithTitle 抬头配置, 当不为nil时, 启用抬头.
func WithTitle(title *Title) Option {
	return func(c *Config) {
		c.title = title
	}
}

// 限制读取最大行数
// <= 0 不限制
func WithLimitReadMaxLine(limit int) Option {
	return func(c *Config) {
		c.limitReadMaxLine = limit
	}
}
