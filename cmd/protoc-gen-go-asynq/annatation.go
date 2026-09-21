package main

import (
	"strings"

	"github.com/thinkgos/proc-extra/proc"
	"google.golang.org/protobuf/compiler/protogen"
)

// annotation const value
const (
	Identity               = "asynq"
	Attribute_Name_Pattern = "pattern"
)

type Task struct {
	Enabled bool
	Pattern string
}

func IsDeriveTaskEnabled(s protogen.Comments) bool {
	derives, _ := proc.NewCommentLines(string(s)).FindDerives(Identity)
	return proc.Derives(derives).ContainHeadless(Identity)
}

func ParserDeriveTask(s protogen.Comments) *Task {
	ret := &Task{}
	derives, _ := proc.NewCommentLines(string(s)).FindDerives(Identity)
	for _, annotate := range derives {
		if annotate.Headless() {
			ret.Enabled = true
			continue
		}
		for _, attr := range annotate.Attrs {
			switch attr.Name {
			case Attribute_Name_Pattern:
				if vv, ok := attr.Value.(proc.String); ok {
					ret.Pattern = strings.TrimSpace(vv.Value)
				}
			}
		}
	}
	if ret.Pattern != "" {
		return ret
	}
	return nil
}
