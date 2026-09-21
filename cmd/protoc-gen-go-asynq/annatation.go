package main

import (
	"slices"
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
	Pattern string
}

func IsEnableDeriveTask(s protogen.Comments) bool {
	derives, _ := proc.NewCommentLines(string(s)).FindDerives(Identity)
	return slices.ContainsFunc(derives, func(p *proc.Derive) bool {
		return p.Identity == Identity && slices.ContainsFunc(p.Attrs, func(attr *proc.NameValue) bool {
			if attr.Name != Attribute_Name_Pattern {
				return false
			}
			vv, ok := attr.Value.(proc.String)
			return ok && strings.TrimSpace(vv.Value) != ""
		})
	})
}

func ParseDeriveTask(s protogen.Comments) *Task {
	ret := &Task{}
	derives, _ := proc.NewCommentLines(string(s)).FindDerives(Identity)
	for _, annotate := range derives {
		if annotate.Identity == Identity {
			for _, attr := range annotate.Attrs {
				switch attr.Name {
				case Attribute_Name_Pattern:
					if vv, ok := attr.Value.(proc.String); ok {
						pattern := strings.TrimSpace(vv.Value)
						if pattern != "" {
							ret.Pattern = pattern
							return ret
						}
					}
				}
			}
		}
	}
	return nil
}
