package excel

import (
	"fmt"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"
)

func decodeFromFile[T any](filename string, opts ...Option) ([]T, error) {
	xls, err := OpenFile[T](filename)
	if err != nil {
		return nil, err
	}
	return xls.Decode("Sheet1", opts...)
}

func Test_Encode_Decode(t *testing.T) {
	encodeDecodeWithOption := func(opts ...Option) error {
		want := []Line{
			{"1111", "1111", 1111},
			{"2222", "2222", 2222},
			{"3333", "3333", 3333},
		}

		filename := randExcelFilename()
		err := encodeToNewFile(filename, want, opts...)
		if err != nil {
			return err
		}
		defer os.Remove(filename) // nolint: errcheck

		got, err := decodeFromFile[Line](filename, opts...)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(got, want) {
			return fmt.Errorf("got = %v, want %v", got, want)
		}
		return nil
	}

	tests := []struct {
		name    string
		options []Option
	}{
		{
			"empty",
			[]Option{},
		},
		{
			"only title",
			[]Option{
				WithTitle(
					NewTitle().
						SetCustomValueFunc(customTitle),
				),
			},
		},
		{
			"only header",
			[]Option{
				WithEnableHeader(),
			},
		},
		{
			"only row start",
			[]Option{
				WithRowStart(4),
			},
		},
		{
			"title and header",
			[]Option{
				WithEnableHeader(),
				WithTitle(
					NewTitle().
						SetCustomValueFunc(customTitle),
				),
			},
		},
		{
			"title and row start",
			[]Option{
				WithRowStart(4),
				WithTitle(
					NewTitle().
						SetCustomValueFunc(customTitle),
				),
			},
		},
		{
			"header and row start",
			[]Option{
				WithEnableHeader(),
				WithRowStart(4),
			},
		},
		{
			"title, header and row start",
			[]Option{
				WithEnableHeader(),
				WithRowStart(4),
				WithTitle(
					NewTitle().
						SetCustomValueFunc(customTitle),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, encodeDecodeWithOption(tt.options...))
		})
	}
}

func Test_Encode_Decode_With_Append(t *testing.T) {
	encodeDecodeWithOption := func(opts ...Option) error {
		wantData1 := []*Line{
			{"1111", "1111", 1111},
			{"2222", "2222", 2222},
			{"3333", "3333", 3333},
		}
		appendData := []*Line{
			{"4444", "4444", 4444},
			{"5555", "5555", 5555},
		}
		wantData2 := append(wantData1, appendData...)

		filename := randExcelFilename()
		err := encodeToNewFile(filename, wantData1, opts...)
		if err != nil {
			return err
		}
		defer os.Remove(filename) // nolint: errcheck

		gotData1, err := decodeFromFile[*Line](filename, opts...)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(gotData1, wantData1) {
			return fmt.Errorf("data1: got = %v, want %v", gotData1, wantData1)
		}

		err = encodeToOldFile(filename, appendData, opts...)
		if err != nil {
			return err
		}

		gotData2, err := decodeFromFile[*Line](filename, opts...)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(gotData2, wantData2) {
			return fmt.Errorf("data2: got = %v, want %v", gotData2, wantData2)
		}
		return nil
	}

	tests := []struct {
		name    string
		options []Option
	}{
		{
			"empty",
			[]Option{},
		},
		{
			"only title",
			[]Option{
				WithTitle(
					NewTitle().
						SetCustomValueFunc(customTitle),
				),
			},
		},
		{
			"only header",
			[]Option{
				WithEnableHeader(),
			},
		},
		{
			"only row start",
			[]Option{
				WithRowStart(4),
			},
		},
		{
			"title and header",
			[]Option{
				WithEnableHeader(),
				WithTitle(
					NewTitle().
						SetCustomValueFunc(customTitle),
				),
			},
		},
		{
			"title and row start",
			[]Option{
				WithRowStart(4),
				WithTitle(
					NewTitle().
						SetCustomValueFunc(customTitle),
				),
			},
		},
		{
			"header and row start",
			[]Option{
				WithEnableHeader(),
				WithRowStart(4),
			},
		},
		{
			"title, header and row start",
			[]Option{
				WithEnableHeader(),
				WithRowStart(4),
				WithTitle(
					NewTitle().
						SetCustomValueFunc(customTitle),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, encodeDecodeWithOption(tt.options...))
		})
	}
}

func Test_Encode_Empty_Matrix(t *testing.T) {
	encodeDecodeWithOption := func(opts ...Option) error {
		want := [][]string{}

		filename := randAlphabet(10) + ".xlsx"
		err := encodeToNewFile(filename, want, opts...)
		if err != nil {
			return err
		}
		defer os.Remove(filename) // nolint: errcheck

		got, err := decodeFromFile[[]string](filename, opts...)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(got, want) {
			return fmt.Errorf("got = %v, want %v", got, want)
		}
		return nil
	}

	err := encodeDecodeWithOption()
	require.NoError(t, err)
}

func Test_Encode_Decode_Matrix(t *testing.T) {
	encodeDecodeWithOption := func(opts ...Option) error {
		want := [][]string{
			{"1111", "1111", "1111"},
			{"2222", "2222", "2222"},
			{"3333", "3333", "3333"},
		}

		filename := randAlphabet(10) + ".xlsx"
		err := encodeToNewFile(filename, want, opts...)
		if err != nil {
			return err
		}
		defer os.Remove(filename) // nolint: errcheck

		got, err := decodeFromFile[[]string](filename, opts...)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(got, want) {
			return fmt.Errorf("got = %v, want %v", got, want)
		}
		return nil
	}

	tests := []struct {
		name    string
		options []Option
	}{
		{
			"empty",
			[]Option{},
		},
		{
			"only title",
			[]Option{
				WithTitle(
					NewTitle().
						SetCustomValueFunc(customTitle),
				),
			},
		},
		{
			"only header",
			[]Option{
				WithEnableHeader(),
			},
		},
		{
			"only row start",
			[]Option{
				WithRowStart(4),
			},
		},
		{
			"title and header",
			[]Option{
				WithEnableHeader(),
				WithTitle(
					NewTitle().
						SetCustomValueFunc(customTitle),
				),
			},
		},
		{
			"title and row start",
			[]Option{
				WithRowStart(4),
				WithTitle(
					NewTitle().
						SetCustomValueFunc(customTitle),
				),
			},
		},
		{
			"header and row start",
			[]Option{
				WithEnableHeader(),
				WithRowStart(4),
			},
		},
		{
			"title, header and row start",
			[]Option{
				WithEnableHeader(),
				WithRowStart(4),
				WithTitle(
					NewTitle().
						SetCustomValueFunc(customTitle),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, encodeDecodeWithOption(tt.options...))
		})
	}
}

func Test_Matrix_Encode_Decode_With_APPEND(t *testing.T) {
	encodeDecodeWithOption := func(opts ...Option) error {
		wantData1 := [][]string{
			{"1111", "1111", "1111"},
			{"2222", "2222", "2222"},
			{"3333", "3333", "3333"},
		}
		appendData := [][]string{
			{"4444", "4444", "4444"},
			{"5555", "5555", "5555"},
		}
		wantData2 := append(wantData1, appendData...)

		filename := randAlphabet(10) + ".xlsx"
		err := encodeToNewFile(filename, wantData1, opts...)
		if err != nil {
			return err
		}
		defer os.Remove(filename) // nolint: errcheck

		gotData1, err := decodeFromFile[[]string](filename, opts...)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(gotData1, wantData1) {
			return fmt.Errorf("data1: got = %v, want %v", gotData1, wantData1)
		}

		err = encodeToOldFile(filename, appendData, opts...)
		if err != nil {
			return err
		}

		gotData2, err := decodeFromFile[[]string](filename, opts...)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(gotData2, wantData2) {
			return fmt.Errorf("data2: got = %v, want %v", gotData2, wantData2)
		}
		return nil
	}

	tests := []struct {
		name    string
		options []Option
	}{
		{
			"empty",
			[]Option{},
		},
		{
			"only title",
			[]Option{
				WithTitle(
					NewTitle().
						SetCustomValueFunc(customTitle),
				),
			},
		},
		{
			"only header",
			[]Option{
				WithEnableHeader(),
			},
		},
		{
			"only row start",
			[]Option{
				WithRowStart(4),
			},
		},
		{
			"title and header",
			[]Option{
				WithEnableHeader(),
				WithTitle(
					NewTitle().
						SetCustomValueFunc(customTitle),
				),
			},
		},
		{
			"title and row start",
			[]Option{
				WithRowStart(4),
				WithTitle(
					NewTitle().
						SetCustomValueFunc(customTitle),
				),
			},
		},
		{
			"header and row start",
			[]Option{
				WithEnableHeader(),
				WithRowStart(4),
			},
		},
		{
			"title, header and row start",
			[]Option{
				WithEnableHeader(),
				WithRowStart(4),
				WithTitle(
					NewTitle().
						SetCustomValueFunc(customTitle),
				),
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NoError(t, encodeDecodeWithOption(tt.options...))
		})
	}
}

func decodeFromFile2[T any](filename string, opts ...Option) ([]T, error) {
	xls, err := OpenFile[T](filename)
	if err != nil {
		return nil, err
	}
	return xls.Decode("Sheet1", opts...)
}
func Test_Encode_Decode2(t *testing.T) {
	encodeDecodeWithOption := func(opts ...Option) error {
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

		got, err := decodeFromFile2[[]string](filename, opts...)
		if err != nil {
			return err
		}
		if !reflect.DeepEqual(got, want) {
			return fmt.Errorf("got = %v, want %v", got, want)
		}
		return nil
	}
	err := encodeDecodeWithOption()
	require.NoError(t, err)
}
