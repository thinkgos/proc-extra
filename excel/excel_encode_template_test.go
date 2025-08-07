package excel

import (
	"testing"

	"github.com/stretchr/testify/require"
)

const testTemplateFile = "./template.xlsx"

type testTemplateTitle struct {
	Name  string
	Area  string
	Total int
}
type testTemplateData struct {
	Idx    int
	Name   string
	Status string
	Model  string
	Type   string
	Ip     string
	Remark string
}

func Test_Encode_Template(t *testing.T) {
	xls, err := OpenFile[testTemplateData](testTemplateFile)
	require.NoError(t, err)
	err = xls.Encode("关联摄像机",
		[]testTemplateData{
			{
				Idx:    1,
				Name:   "测试1",
				Status: "正常",
				Model:  "测试模型1",
				Type:   "测试类型1",
				Ip:     "192.168.1.1",
				Remark: "测试备注1",
			},
			{
				Idx:    2,
				Name:   "测试2",
				Status: "正常",
				Model:  "测试模型2",
				Type:   "测试类型2",
				Ip:     "192.168.1.2",
				Remark: "测试备注2",
			},
		},

		NewTitle().SetUseTemplate(&testTemplateTitle{
			Name:  "测一下",
			Area:  "测试区域",
			Total: 2,
		}).
			BuildOption(),
		WithOverrideRow(),
		WithRowStart(3),
		WithDataCellStyleBaseOnRow(3),
	)
	require.NoError(t, err)
	filename := randExcelFilename()
	err = xls.SaveAs(filename)
	require.NoError(t, err)
}
