package excel

import (
	"reflect"
	"testing"
)

func Test_ParseCellDef(t *testing.T) {
	tests := []struct {
		name     string
		tagName  string
		wantCell *CellDef
	}{
		{
			"full cell tag",
			"卡号",
			&CellDef{
				"卡号",
				nil,
			},
		},
		{
			"empty cell tag",
			"",
			&CellDef{
				"",
				nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCell := parseCellDef(tt.tagName)
			if !reflect.DeepEqual(gotCell, tt.wantCell) {
				t.Errorf("parseCellTag() gotCell = %v, want %v", gotCell, tt.wantCell)
			}
		})
	}
}
