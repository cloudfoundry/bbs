package models

import (
	"flag"
	"fmt"
	"log"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/types/pluginpb"
)

const version = "0.0.1"

var prefix *string
var debug *bool

func main() {
	showVersion := flag.Bool("version", false, "print the version and exit")
	flag.Parse()
	if *showVersion {
		fmt.Printf("protoc-gen-go-bbs %v\n", version)
		return
	}

	var flags flag.FlagSet
	prefix = flags.String("prefix", "Proto", "Prefix to strip from protobuf-generated structs")
	debug = flags.Bool("debug", false, "Set to true to output codegen debugging lines")

	protogen.Options{
		ParamFunc: flags.Set,
	}.Run(func(plugin *protogen.Plugin) error {
		plugin.SupportedFeatures = uint64(pluginpb.CodeGeneratorResponse_FEATURE_PROTO3_OPTIONAL)

		for _, file := range plugin.Files {
			if file.GeneratedFilenamePrefix == "bbs" {
				// ignore the bbs.proto file for our plugin always
				file.Generate = false
			}
			if !file.Generate {
				continue
			}
			log.Printf("Generating %s\n", file.Desc.Path())

			generateFile(plugin, file)
		}
		return nil
	})
}
