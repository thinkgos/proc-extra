package excel

import (
	"reflect"
	"testing"
)

func Test_ParseCellTag(t *testing.T) {
	tests := []struct {
		name     string
		tagName  string
		wantCell Cell
		wantErr  bool
	}{
		{
			"invalid cell tag",
			"",
			Cell{},
			true,
		},
		{
			"full cell tag",
			"A 卡号 40 20 3",
			Cell{
				"A",
				"卡号",
				40,
				20,
				3,
			},
			false,
		},
		{
			"omit if not interest",
			"A 卡号 - - 3",
			Cell{
				"A",
				"卡号",
				0,
				0,
				3,
			},
			false,
		},
		{
			"ignore with not set",
			"A 卡号",
			Cell{
				"A",
				"卡号",
				0,
				0,
				0,
			},
			false,
		},
		{
			"only column",
			"A",
			Cell{
				"A",
				"",
				0,
				0,
				0,
			},
			false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCell, err := parseCellTag(tt.tagName)
			if (err != nil) != tt.wantErr {
				t.Errorf("parseCellTag() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotCell, tt.wantCell) {
				t.Errorf("parseCellTag() gotCell = %v, want %v", gotCell, tt.wantCell)
			}
		})
	}
}
