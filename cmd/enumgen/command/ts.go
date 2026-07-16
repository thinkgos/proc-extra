package command

import (
	"github.com/spf13/cobra"
	"github.com/thinkgos/proc-extra/cmd/enumgen/internal/enumgen"
)

type TsOption struct {
	Uri       string // uri: file://enum.json 或 http://xxx.com/a.json
	TypeStyle string // 字典类型type风格
	Filename  string // 输出文件名
}

type TsCmd struct {
	cmd *cobra.Command
	TsOption
}

func NewTsCmd() *TsCmd {
	root := &TsCmd{}
	cmd := &cobra.Command{
		Use:           "ts",
		Short:         "ts generate enum ts from enum",
		Long:          "ts generate enum ts from enum",
		Version:       BuildVersion(),
		SilenceUsage:  false,
		SilenceErrors: false,
		RunE: func(*cobra.Command, []string) error {
			g := &enumgen.GenTs{
				Uri:       root.Uri,
				Version:   version,
				Filename:  root.Filename,
				TypeStyle: root.TypeStyle,
			}
			return g.Gen()
		},
		Args: cobra.NoArgs,
	}
	cmd.Flags().StringVar(&root.Uri, "uri", "", "uri: file://enum.json 或 http://xxx.com/a.json")
	cmd.Flags().StringVar(&root.TypeStyle, "typeStyle", "", "字典类型type风格, 支持snakecase,pascalcase,camelcase,kebabcase")
	cmd.Flags().StringVarP(&root.Filename, "filename", "f", "", "输出文件名")
	cmd.MarkFlagRequired("filename")

	root.cmd = cmd
	return root
}
