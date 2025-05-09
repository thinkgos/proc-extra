package excel

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
)

type Line struct {
	CarNo    string `xlsx:"A 卡号 20 20 3"`
	PrefixNo string `xlsx:"B 前缀 - - 3"`
	No       int    `xlsx:"C 编号 - - 3,omitempty"`
}

func encodeToNewFile[T any](filename string, data []T, opts ...Option) error {
	xls := NewFile[T]()
	err := xls.Encode("Sheet1", data, opts...)
	if err != nil {
		return err
	}
	return xls.SaveAs(filename)
}

func encodeToOldFile[T any](filename string, data []T, opts ...Option) error {
	xls, err := OpenFile[T](filename)
	if err != nil {
		return err
	}
	err = xls.Encode("Sheet1", data, opts...)
	if err != nil {
		return err
	}
	return xls.SaveAs(filename)
}

func Test_Encode_SliceStruct(t *testing.T) {
	encodeWithOption := func(opts ...Option) error {
		want := []*Line{
			{"1111", "1111", 1111},
			{"2222", "2222", 0},
			{"3333", "3333", 3333},
		}

		filename := randExcelFilename()
		err := encodeToNewFile(filename, want, opts...)
		if err != nil {
			return err
		}
		os.Remove(filename) // nolint: errcheck
		return nil
	}
	t.Run("empty data", func(t *testing.T) {
		filename := randExcelFilename()
		err := encodeToNewFile(filename, []*Line{},
			WithEnableHeader(),
		)
		require.NoError(t, err)
		os.Remove(filename) // nolint: errcheck
	})
	t.Run("empty option", func(t *testing.T) {
		err := encodeWithOption()
		require.NoError(t, err)
	})
	t.Run("title option", func(t *testing.T) {
		err := encodeWithOption(
			WithTitle(
				NewTitle().
					SetCustomValueFunc(customTitle),
			),
		)
		require.NoError(t, err)
	})
	t.Run("header option", func(t *testing.T) {
		err := encodeWithOption(
			WithEnableHeader(),
		)
		require.NoError(t, err)
	})
	t.Run("title and header option", func(t *testing.T) {
		err := encodeWithOption(
			WithEnableHeader(),
			WithTitle(
				NewTitle().
					SetCustomValueFunc(customTitle),
			),
		)
		require.NoError(t, err)
	})
	t.Run("title and row start option", func(t *testing.T) {
		err := encodeWithOption(
			WithRowStart(4),
			WithTitle(
				NewTitle().
					SetCustomValueFunc(customTitle),
			),
		)
		require.NoError(t, err)
	})

	t.Run("header and row start option", func(t *testing.T) {
		err := encodeWithOption(
			WithEnableHeader(),
			WithRowStart(4),
		)
		require.NoError(t, err)
	})
	t.Run("title,header and row start option", func(t *testing.T) {
		err := encodeWithOption(
			WithEnableHeader(),
			WithRowStart(4),
			WithTitle(
				NewTitle().
					SetCustomValueFunc(customTitle),
			),
		)
		require.NoError(t, err)
	})
	t.Run("title, header and row start option(custom header)", func(t *testing.T) {
		err := encodeWithOption(
			WithEnableHeader(),
			WithHeaders([]string{"A1", "A2", "A3"}),
			WithRowStart(4),
			WithTitle(
				NewTitle().
					SetCustomValueFunc(customTitle),
			),
		)
		require.NoError(t, err)
	})
}

func Test_Encode_SliceStruct_Append(t *testing.T) {
	encodeWithOption := func(opts ...Option) error {
		want := []*Line{
			{"1111", "1111", 1111},
			{"2222", "2222", 2222},
			{"3333", "3333", 3333},
		}
		appendData := []*Line{
			{"4444", "4444", 4444},
			{"5555", "5555", 5555},
		}

		filename := randExcelFilename()
		err := encodeToNewFile(filename, want, opts...)
		if err != nil {
			return err
		}
		defer os.Remove(filename) // nolint: errcheck

		err = encodeToOldFile(filename, appendData, opts...)
		if err != nil {
			return err
		}
		return nil
	}
	t.Run("empty option", func(t *testing.T) {
		err := encodeWithOption()
		require.NoError(t, err)
	})
	t.Run("title option", func(t *testing.T) {
		err := encodeWithOption(
			WithTitle(
				NewTitle().
					SetCustomValueFunc(customTitle),
			),
		)
		require.NoError(t, err)
	})
	t.Run("header option", func(t *testing.T) {
		err := encodeWithOption(
			WithEnableHeader(),
		)
		require.NoError(t, err)
	})
	t.Run("title and header option", func(t *testing.T) {
		err := encodeWithOption(
			WithEnableHeader(),
			WithTitle(
				NewTitle().
					SetCustomValueFunc(customTitle),
			),
		)
		require.NoError(t, err)
	})
	t.Run("title and row start option", func(t *testing.T) {
		err := encodeWithOption(
			WithRowStart(4),
			WithTitle(
				NewTitle().
					SetCustomValueFunc(customTitle),
			),
		)
		require.NoError(t, err)
	})

	t.Run("header and row start option", func(t *testing.T) {
		err := encodeWithOption(
			WithEnableHeader(),
			WithRowStart(4),
		)
		require.NoError(t, err)
	})
	t.Run("title,header and row start option", func(t *testing.T) {
		err := encodeWithOption(
			WithEnableHeader(),
			WithRowStart(4),
			WithTitle(
				NewTitle().
					SetCustomValueFunc(customTitle),
			),
		)
		require.NoError(t, err)
	})
}

func Test_Encode_Matrix(t *testing.T) {
	encodeWithOption := func(opts ...Option) error {
		want := [][]string{
			{"1111", "1111", "1111"},
			{"2222", "2222", "2222"},
			{"3333", "3333", "3333"},
		}

		filename := randExcelFilename()
		err := encodeToNewFile(filename, want, opts...)
		if err != nil {
			return err
		}
		defer os.Remove(filename) // nolint: errcheck
		return nil
	}

	t.Run("empty data", func(t *testing.T) {
		filename := randExcelFilename()
		err := encodeToNewFile(filename, [][]string{})
		require.NoError(t, err)
		defer os.Remove(filename) // nolint: errcheck
	})
	t.Run("empty option", func(t *testing.T) {
		err := encodeWithOption()
		require.NoError(t, err)
	})
	t.Run("title option", func(t *testing.T) {
		err := encodeWithOption(
			WithTitle(
				NewTitle().
					SetCustomValueFunc(customTitle),
			),
		)
		require.NoError(t, err)
	})
	t.Run("header option", func(t *testing.T) {
		err := encodeWithOption(
			WithEnableHeader(),
			WithHeaders([]string{"卡号", "前缀", "编号"}),
		)
		require.NoError(t, err)
	})
	t.Run("title and header option", func(t *testing.T) {
		err := encodeWithOption(
			WithEnableHeader(),
			WithHeaders([]string{"卡号", "前缀", "编号"}),
			WithTitle(
				NewTitle().
					SetCustomValueFunc(customTitle),
			),
		)
		require.NoError(t, err)
	})
	t.Run("title and row start option", func(t *testing.T) {
		err := encodeWithOption(
			WithRowStart(4),
			WithTitle(
				NewTitle().
					SetCustomValueFunc(customTitle),
			),
		)
		require.NoError(t, err)
	})

	t.Run("header and row start option", func(t *testing.T) {
		err := encodeWithOption(
			WithEnableHeader(),
			WithHeaders([]string{"卡号", "前缀", "编号"}),
			WithRowStart(4),
		)
		require.NoError(t, err)
	})
	t.Run("title,header and row start option", func(t *testing.T) {
		err := encodeWithOption(
			WithEnableHeader(),
			WithHeaders([]string{"卡号", "前缀", "编号"}),
			WithRowStart(4),
			WithTitle(
				NewTitle().
					SetCustomValueFunc(customTitle),
			),
		)
		require.NoError(t, err)
	})
}

func encodeToNewFileWithTransformRow[T any](filename string, data []T, f func(value T) ([]any, error), opts ...Option) error {
	xls := NewFile[T]().SetTransformRowValue(f)
	err := xls.Encode("Sheet1", data, opts...)
	if err != nil {
		return err
	}
	return xls.SaveAs(filename)
}

func Test_Encode_SliceStruct_TransformRowValue(t *testing.T) {
	t.Run(" data", func(t *testing.T) {
		want := []*Line{
			{"1111", "1111", 1111},
			{"2222", "2222", 0},
			{"3333", "3333", 3333},
		}
		filename := randExcelFilename()
		err := encodeToNewFileWithTransformRow(filename, want,
			func(v *Line) ([]any, error) {
				return []any{"a" + v.CarNo, "a" + v.PrefixNo, "a" + v.CarNo}, nil
			},
		)
		require.NoError(t, err)
		os.Remove(filename) // nolint: errcheck
	})
}

func Test_Encode_Matrix_TransformRowValue(t *testing.T) {
	t.Run(" data", func(t *testing.T) {
		want := [][]string{
			{"1111", "1111", "1111"},
			{"2222", "2222", ""},
			{"3333", "3333", "3333"},
		}
		filename := randExcelFilename()
		err := encodeToNewFileWithTransformRow(filename, want,
			func(v []string) ([]any, error) {
				return []any{"a" + v[0], "a" + v[1], "a" + v[2]}, nil
			},
		)
		require.NoError(t, err)
		os.Remove(filename) // nolint: errcheck
	})
}
