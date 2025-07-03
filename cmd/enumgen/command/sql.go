package command

import (
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/thinkgos/proc-extra/cmd/enumgen/internal/enumgen"
)

type EnumSqlOption struct {
	Pattern     []string
	Type        []string
	Tags        []string
	SqlDictType string // 字典类型模板
	SqlDictItem string // 字典项模板
}

type SqlCmd struct {
	cmd   *cobra.Command
	level string
	EnumSqlOption
}

func NewSqlCmd() *SqlCmd {
	root := &SqlCmd{}
	cmd := &cobra.Command{
		Use:           "sql",
		Short:         "sql generate enum sql from enum",
		Long:          "sql generate enum sql from enum",
		Version:       BuildVersion(),
		SilenceUsage:  false,
		SilenceErrors: false,
		RunE: func(*cobra.Command, []string) error {
			srcDir := root.Pattern[0]
			fileInfo, err := os.Stat(srcDir)
			if err != nil {
				return err
			}
			if !fileInfo.IsDir() {
				if len(root.Tags) != 0 {
					slog.Error("--tags option applies only to directories, not when files are specified")
					os.Exit(1)
				}
				srcDir = filepath.Dir(srcDir)
			}
			g := &enumgen.Gen{
				Pattern:     root.Pattern,
				OutputDir:   srcDir,
				Type:        root.Type,
				Tags:        root.Tags,
				Version:     version,
				Merge:       false,
				Filename:    "",
				SqlDictType: root.SqlDictType,
				SqlDictItem: root.SqlDictItem,
			}
			if err = g.Init(); err != nil {
				return err
			}
			return g.GenSql()
		},
		Args: cobra.NoArgs,
	}
	cobra.OnInitialize(func() {
		textHandler := slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
			AddSource:   false,
			Level:       level(root.level),
			ReplaceAttr: nil,
		})
		slog.SetDefault(slog.New(textHandler))
	})

	cmd.PersistentFlags().StringVarP(&root.level, "level", "l", "info", "log level(debug,info,warn,error)")
	cmd.Flags().StringSliceVarP(&root.Pattern, "pattern", "p", []string{"."}, "the list of files or a directory.")
	cmd.Flags().StringSliceVarP(&root.Type, "type", "t", nil, "the list type of enum names; must be set")
	cmd.Flags().StringSliceVar(&root.Tags, "tags", nil, "comma-separated list of build tags to apply")
	cmd.Flags().StringVar(&root.SqlDictType, "sqlDictType", enumgen.DefaultDictTypeTpl, "字典类型模板, 按顺序为[类型, 名称, 备注(同名称)]")
	cmd.Flags().StringVar(&root.SqlDictItem, "sqlDictItem", enumgen.DefaultDictItemTpl, "字典项模板, 按顺序为[类型, 标签, 值, 顺序, 备注(同名称)]")

	root.cmd = cmd
	return root
}
