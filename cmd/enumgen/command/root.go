package command

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/thinkgos/proc-extra/cmd/enumgen/internal/enumgen"
)

type EnumOption struct {
	Pattern  []string
	Type     []string
	Tags     []string
	Merge    bool
	Filename string
}

type RootCmd struct {
	cmd   *cobra.Command
	level string
	EnumOption
}

func NewRootCmd() *RootCmd {
	root := &RootCmd{}
	cmd := &cobra.Command{
		Use:           "enumgen",
		Short:         "enumgen generate enum code from enum",
		Long:          "enumgen generate enum code from enum",
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
				Pattern:   root.Pattern,
				OutputDir: srcDir,
				Type:      root.Type,
				Tags:      root.Tags,
				Version:   version,
				Merge:     root.Merge,
				Filename:  root.Filename,
			}
			err = g.GenEnum()
			if err != nil {
				slog.Error("生成失败", slog.Any("err", err))
			}
			return nil
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
	cmd.Flags().BoolVar(&root.Merge, "merge", false, "merge in a file")
	cmd.Flags().StringVar(&root.Filename, "filename", "", "filename when merge enabled, default: enums")

	root.cmd = cmd
	return root
}

// Execute adds all child commands to the root command and sets flags appropriately.
func (r *RootCmd) Execute() error {
	return r.cmd.Execute()
}
func level(s string) slog.Level {
	switch strings.ToUpper(s) {
	case "DEBUG":
		return slog.LevelDebug
	case "INFO":
		return slog.LevelInfo
	case "WARN":
		return slog.LevelWarn
	case "ERROR":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
