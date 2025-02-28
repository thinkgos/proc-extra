package excel

import (
	"io"

	"github.com/xuri/excelize/v2"
)

// File 构建excel文档
type File[T any] struct {
	*excelize.File
	transformRowValue func(value T) ([]any, error) // 转换整行值, 用于编码
}

func NewFile[T any]() *File[T] {
	return &File[T]{File: excelize.NewFile()}
}

func OpenFile[T any](filename string, opt ...excelize.Options) (*File[T], error) {
	file, err := excelize.OpenFile(filename, opt...)
	if err != nil {
		return nil, err
	}
	return &File[T]{File: file}, nil
}

func OpenReader[T any](r io.Reader, opt ...excelize.Options) (*File[T], error) {
	file, err := excelize.OpenReader(r, opt...)
	if err != nil {
		return nil, err
	}
	return &File[T]{File: file}, nil
}

// WriteTo implements io.WriterTo to write the file.
func (e *File[T]) WriteTo(w io.Writer) (n int64, err error) {
	return e.File.WriteTo(w)
}

func (e *File[T]) SetTransformRowValue(f func(value T) ([]any, error)) *File[T] {
	e.transformRowValue = f
	return e
}
