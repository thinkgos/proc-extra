package main

import (
	"flag"
	"fmt"

	"google.golang.org/protobuf/compiler/protogen"
)

var args = &struct {
	ShowVersion     bool // 显示版本
	DisableValidate bool // 禁用客户端校验
}{
	ShowVersion:     false,
	DisableValidate: false,
}

func init() {
	flag.BoolVar(&args.ShowVersion, "version", false, "print the version and exit")
	flag.BoolVar(&args.DisableValidate, "disable_validate", false, "disable client validation")
}

func main() {
	flag.Parse()
	if args.ShowVersion {
		fmt.Printf("protoc-gen-go-asynq %v\n", version)
		return
	}
	protogen.Options{ParamFunc: flag.CommandLine.Set}.Run(runProtoGen)
}
