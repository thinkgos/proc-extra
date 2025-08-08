package excel

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"
	"github.com/xuri/excelize/v2"
)

type Line struct {
	CarNo    string `xlsx:"卡号"`
	PrefixNo string `xlsx:"前缀"`
	No       int    `xlsx:"编号,omitempty"`
}

func encodeToNewFile[T any](filename string, data []T, removed bool, opts ...Option) error {
	xls := NewFile[T]()
	err := xls.Encode("Sheet1", data, opts...)
	if err != nil {
		return err
	}
	err = xls.SaveAs(filename)
	if err != nil {
		return err
	}
	if removed {
		os.Remove(filename)
	}
	return nil
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
	return xls.Save()
}

func Test_Encode_SliceStruct(t *testing.T) {
	encodeWithOption := func(opts ...Option) error {
		return encodeToNewFile(
			randExcelFilename(),
			[]*Line{
				{"1111", "1111", 1111},
				{"2222", "2222", 0},
				{"3333", "3333", 3333},
			},
			true,
			opts...,
		)
	}

	t.Run("empty data", func(t *testing.T) {
		err := encodeToNewFile(
			randExcelFilename(),
			[]*Line{},
			true,
			WithEnableHeader(),
		)
		require.NoError(t, err)
	})

	t.Run("empty option", func(t *testing.T) {
		err := encodeWithOption()
		require.NoError(t, err)
	})

	t.Run("title option", func(t *testing.T) {
		err := encodeWithOption(
			NewTitle().
				SetTitle(customTitle()).
				BuildOption(),
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
			NewTitle().
				SetTitle(customTitle()).
				BuildOption(),
		)
		require.NoError(t, err)
	})

	t.Run("title and row start option", func(t *testing.T) {
		err := encodeWithOption(
			WithRowStart(4),
			NewTitle().
				SetTitle(customTitle()).
				BuildOption(),
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
			NewTitle().
				SetTitle(customTitle()).
				BuildOption(),
		)
		require.NoError(t, err)
	})

	t.Run("title, header and row start option(custom header)", func(t *testing.T) {
		err := encodeWithOption(
			WithEnableHeader(),
			WithCustomHeaders([]string{"A1", "A2", "A3"}),
			WithRowStart(4),
			NewTitle().
				SetTitle(customTitle()).
				BuildOption(),
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
		err := encodeToNewFile(filename, want, false, opts...)
		if err != nil {
			return err
		}
		defer os.Remove(filename) // nolint: errcheck
		return encodeToOldFile(filename, appendData, opts...)
	}

	t.Run("empty option", func(t *testing.T) {
		err := encodeWithOption()
		require.NoError(t, err)
	})

	t.Run("title option", func(t *testing.T) {
		err := encodeWithOption(
			NewTitle().
				SetTitle(customTitle()).
				BuildOption(),
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
			NewTitle().
				SetTitle(customTitle()).
				BuildOption(),
		)
		require.NoError(t, err)
	})

	t.Run("title and row start option", func(t *testing.T) {
		err := encodeWithOption(
			WithRowStart(4),
			NewTitle().
				SetTitle(customTitle()).
				BuildOption(),
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

	t.Run("title header and row start option", func(t *testing.T) {
		err := encodeWithOption(
			WithEnableHeader(),
			WithRowStart(4),
			NewTitle().
				SetTitle(customTitle()).
				BuildOption(),
		)
		require.NoError(t, err)
	})
}

func Test_Encode_Matrix(t *testing.T) {
	encodeWithOption := func(opts ...Option) error {
		return encodeToNewFile(
			randExcelFilename(),
			[][]string{
				{"1111", "1111", "1111"},
				{"2222", "2222", "2222"},
				{"3333", "3333", "3333"},
			},
			true,
			opts...,
		)
	}

	t.Run("empty data", func(t *testing.T) {
		filename := randExcelFilename()
		err := encodeToNewFile(filename, [][]string{}, true)
		require.NoError(t, err)
	})

	t.Run("empty option", func(t *testing.T) {
		err := encodeWithOption()
		require.NoError(t, err)
	})

	t.Run("title option", func(t *testing.T) {
		err := encodeWithOption(
			NewTitle().
				SetTitle(customTitle()).
				BuildOption(),
		)
		require.NoError(t, err)
	})

	t.Run("header option", func(t *testing.T) {
		err := encodeWithOption(
			WithEnableHeader(),
			WithCustomHeaders([]string{"卡号", "前缀", "编号"}),
		)
		require.NoError(t, err)
	})

	t.Run("title and header option", func(t *testing.T) {
		err := encodeWithOption(
			WithEnableHeader(),
			WithCustomHeaders([]string{"卡号", "前缀", "编号"}),
			NewTitle().
				SetTitle(customTitle()).
				BuildOption(),
		)
		require.NoError(t, err)
	})

	t.Run("title and row start option", func(t *testing.T) {
		err := encodeWithOption(
			WithRowStart(4),
			NewTitle().
				SetTitle(customTitle()).
				BuildOption(),
		)
		require.NoError(t, err)
	})

	t.Run("header and row start option", func(t *testing.T) {
		err := encodeWithOption(
			WithEnableHeader(),
			WithCustomHeaders([]string{"卡号", "前缀", "编号"}),
			WithRowStart(4),
		)
		require.NoError(t, err)
	})

	t.Run("title header and row start option", func(t *testing.T) {
		err := encodeWithOption(
			WithEnableHeader(),
			WithCustomHeaders([]string{"卡号", "前缀", "编号"}),
			WithRowStart(4),
			NewTitle().
				SetTitle(customTitle()).
				SetStyle(&excelize.Style{
					Border: []excelize.Border{
						{Type: "left", Color: "000000", Style: 1},
						{Type: "right", Color: "000000", Style: 1},
						{Type: "top", Color: "000000", Style: 1},
						{Type: "bottom", Color: "000000", Style: 1},
					},
					Fill: excelize.Fill{Type: "pattern", Pattern: 1, Color: []string{"FFFF00"}, Shading: 0},
					Font: &excelize.Font{
						Bold:         true,
						Italic:       false,
						Underline:    "",
						Family:       "微软雅黑",
						Size:         14,
						Strike:       false,
						Color:        "",
						ColorIndexed: 0,
						ColorTheme:   nil,
						ColorTint:    0,
						VertAlign:    "",
					},
					Alignment: &excelize.Alignment{
						Horizontal:      "center",
						Indent:          0,
						JustifyLastLine: false,
						ReadingOrder:    0,
						RelativeIndent:  0,
						ShrinkToFit:     false,
						TextRotation:    0,
						Vertical:        "center",
						WrapText:        true,
					},
					Protection:    nil,
					NumFmt:        0,
					DecimalPlaces: nil,
					CustomNumFmt:  nil,
					NegRed:        false,
				}).
				BuildOption(),
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
