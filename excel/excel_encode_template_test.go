package excel

import (
	"os"
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
			{
				Idx:    3,
				Name:   "测试3",
				Status: "正常",
				Model:  "测试模型3",
				Type:   "测试类型3",
				Ip:     "192.168.1.3",
				Remark: "测试备注3",
			},
		},

		WithOverrideRow(),
		WithRowStart(3),
		WithDataCellStyleBaseOnRow(3),
		NewTitle().
			SetUseTemplate(&testTemplateTitle{
				Name:  "测一下",
				Area:  "测试区域",
				Total: 2,
			}).
			BuildOption(),
	)
	style, err := xls.GetCellDirectlyStyle("关联摄像机", "A2")
	require.NoError(t, err)
	t.Log(style)
	t.Log(style.Font)
	t.Log(style.Alignment)

	require.NoError(t, err)
	filename := randExcelFilename()
	err = xls.SaveAs(filename)
	require.NoError(t, err)
	os.Remove(filename)
}
