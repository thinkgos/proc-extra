package command

import (
	"log/slog"
	"os"

	"github.com/spf13/cobra"
	"github.com/thinkgos/proc-extra/cmd/enumgen/internal/astutil"
	"github.com/thinkgos/proc-extra/cmd/enumgen/internal/enumgen"
)

type DocOption struct {
	Pattern  []string
	Type     []string
	Tags     []string
	Filename string // 输出文件名
}

type DocCmd struct {
	cmd *cobra.Command
	DocOption
}

func NewDocCmd() *DocCmd {
	root := &DocCmd{}
	cmd := &cobra.Command{
		Use:           "doc",
		Short:         "doc generate enum doc from enum",
		Long:          "doc generate enum doc from enum",
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
			g := &enumgen.GenDoc{
				AstInspect: astutil.AstInspect{
					Pattern: root.Pattern,
					Type:    root.Type,
					Tags:    root.Tags,
				},
				Version:  version,
				Filename: root.Filename,
			}
			if err = g.Init(); err != nil {
				return err
			}
			return g.Gen()
		},
		Args: cobra.NoArgs,
	}

	cmd.Flags().StringSliceVarP(&root.Pattern, "pattern", "p", []string{"."}, "the list of files or a directory.")
	cmd.Flags().StringSliceVarP(&root.Type, "type", "t", nil, "the list type of enum names; must be set")
	cmd.Flags().StringSliceVar(&root.Tags, "tags", nil, "comma-separated list of build tags to apply")
	cmd.Flags().StringVarP(&root.Filename, "filename", "f", "", "输出文件名")
	cmd.MarkFlagRequired("filename")

	root.cmd = cmd
	return root
}
