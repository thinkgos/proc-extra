package excel

// Option config option
type Option func(*Config)

// Config 选项配置
type Config struct {
	// 是否覆盖旧工作表数据, 默认: 追加
	overrideSheet bool
	// 是否覆盖旧行数据, 默认: 追加
	overrideRow bool
	// 使能表头输出, 如果启用则使用tag定义的表头, 可自定义表头, 默认: 无表头
	enableHeader bool
	// 自定义表头, enableHeader = true 有效, 默认: 空
	customHeaders []string
	// 输出起始行(含表头, 如果无表头, 则为数据行)
	// title != nil, 如果 rowStart < 2 则 rowStart = 2, 即至少保留一行给抬头
	// 其它 rowStart = 实际值, 如果 rowStart < 1 则 rowStart = 1, 至少从第一行开始
	// 前面的行数将被合并做为抬头
	rowStart int
	// 抬头配置, 当不为nil时, 启用抬头.
	title *Title
	// 限制读取最大行数, <= 0 不限制
	readMaxLines int
	// 数据单元格样式(行高, 列宽)基于哪一行, 仅对数据行有效
	// 0: 表示不使用样式, 其它: 基于所指定的行
	dataCellStyleBaseRow int
}

func (c *Config) takeOptions(opts ...Option) {
	for _, opt := range opts {
		opt(c)
	}
	if c.title != nil {
		c.rowStart = max(c.rowStart, 2)
	} else {
		c.rowStart = max(c.rowStart, 1)
	}
}

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

// WithCustomHeaders 自定义表头, enableHeader = true 有效, 默认: 空
func WithCustomHeaders(headers []string) Option {
	return func(c *Config) {
		c.customHeaders = headers
	}
}

// WithRowStart 输出的起始行(含表头, 如果无表头, 则为数据行)
// 启用抬头配置(title != nil), 如果 rowStart < 2 则 rowStart = 2, 即保留一行给抬头
// 其它 rowStart = 实际值, 如果 rowStart < 1 则 rowStart = 1, 从第一行开始
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

// WithReadMaxLines 限制读取最大行数, <= 0 不限制
func WithReadMaxLines(limit int) Option {
	return func(c *Config) {
		c.readMaxLines = limit
	}
}

// WithDataCellStyleBaseRow 数据单元格样式基于哪一行, 仅对数据行有效
// 0: 表示不使用样式, 其它: 基于所指定的行
func WithDataCellStyleBaseRow(row int) Option {
	return func(c *Config) {
		c.dataCellStyleBaseRow = row
	}
}
