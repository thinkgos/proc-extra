package command

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/thinkgos/proc-extra/cmd/enumgen/internal/enumgen"
)

type TsOption struct {
	OutputDir string
	Pattern   []string
	Type      []string
	Tags      []string
	TypeStyle string // 字典类型type风格
	OmitZero  bool   // 忽略零值
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
			fileInfo, err := os.Stat(root.Pattern[0])
			if err != nil {
				return err
			}
			if !fileInfo.IsDir() {
				if len(root.Tags) != 0 {
					slog.Error("--tags option applies only to directories, not when files are specified")
					os.Exit(1)
				}
			}
			g := &enumgen.Gen{
				Pattern:      root.Pattern,
				OutputDir:    root.OutputDir,
				Type:         root.Type,
				Tags:         root.Tags,
				Version:      version,
				Merge:        false,
				Filename:     "",
				OmitZero:     root.OmitZero,
				TypeStyle:    root.TypeStyle,
				SqlDictType:  "",
				SqlDictItem:  "",
				IsOrderValue: false,
			}
			if err = g.Init(); err != nil {
				return err
			}
			return g.GenTs()
		},
		Args: cobra.NoArgs,
	}

	cmd.Flags().StringVarP(&root.OutputDir, "output", "o", ".", "output directory")
	cmd.Flags().StringSliceVarP(&root.Pattern, "pattern", "p", []string{"."}, "the list of files or a directory.")
	cmd.Flags().StringSliceVarP(&root.Type, "type", "t", nil, "the list type of enum names; must be set")
	cmd.Flags().StringSliceVar(&root.Tags, "tags", nil, "comma-separated list of build tags to apply")
	cmd.Flags().StringVar(&root.TypeStyle, "typeStyle", "", "字典类型type风格, 支持snakeCase,pascalCase,smallCamelCase,kebab")
	cmd.Flags().BoolVar(&root.OmitZero, "omitZero", false, "是否忽略零值")

	root.cmd = cmd
	return root
}
