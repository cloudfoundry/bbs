package main

import (
	"fmt"
	"log"
	"slices"
	"strings"

	"google.golang.org/protobuf/compiler/protogen"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/descriptorpb"
)

type bbsGenerateHelperInterface interface {
	genCopysafeStruct(g *protogen.GeneratedFile, msg *protogen.Message)
	genToProtoMethod(g *protogen.GeneratedFile, msg *protogen.Message)
	genProtoMapMethod(g *protogen.GeneratedFile, msg *protogen.Message)
	genFriendlyEnums(g *protogen.GeneratedFile, msg *protogen.Message)
	genAccessors(g *protogen.GeneratedFile, msg *protogen.Message)
}
type bbsGenerateHelper struct{}

var helper bbsGenerateHelperInterface = bbsGenerateHelper{}

func getUnsafeName(g *protogen.GeneratedFile, ident protogen.GoIdent) string {
	return g.QualifiedGoIdent(ident)
}

func getCopysafeName(g *protogen.GeneratedFile, ident protogen.GoIdent) (string, bool) {
	unsafeName := getUnsafeName(g, ident)
	return strings.CutPrefix(unsafeName, *prefix)
}

func (bbsGenerateHelper) genCopysafeStruct(g *protogen.GeneratedFile, msg *protogen.Message) {
	if copysafeName, ok := getCopysafeName(g, msg.GoIdent); ok {
		g.P("// Prevent copylock errors when using ", msg.GoIdent.GoName, " directly")
		g.P("type ", copysafeName, " struct {")
		for _, field := range msg.Fields {
			options := field.Desc.Options().(*descriptorpb.FieldOptions)
			if *debug {
				log.Printf("New Field Detected: %+v\n\n", field)
				log.Printf("Field Options: %+v\n\n", options)

			}
			fieldType := getActualType(g, field)
			g.P(field.GoName, " ", fieldType)
		}
		g.P("}")

		helper.genAccessors(g, msg)
	}
}

func (bbsGenerateHelper) genAccessors(g *protogen.GeneratedFile, msg *protogen.Message) {
	if copysafeName, ok := getCopysafeName(g, msg.GoIdent); ok {
		for _, field := range msg.Fields {
			fieldName := field.GoName
			fieldType := getActualType(g, field)
			options := field.Desc.Options().(*descriptorpb.FieldOptions)

			if *debug {
				log.Printf("Generating accessors for %s...\n", fieldName)
			}

			if options.GetDeprecated() {
				g.P("// DEPRECATED: DO NOT USE")
			}
			defaultValue := getDefaultValueString(field)
			genGetter(g, copysafeName, fieldName, fieldType, defaultValue)
			genSetter(g, copysafeName, fieldName, fieldType)
		}
	}
}

func genGetter(g *protogen.GeneratedFile, copysafeName string, fieldName string, fieldType string, defaultValue string) {
	if *debug {
		log.Print("Getter...")
	}
	g.P("func (m *", copysafeName, ") Get", fieldName, "() ", fieldType, " {")
	g.P("if m != nil {")
	g.P("return m.", fieldName)
	g.P("}")
	g.P("return ", defaultValue)
	g.P("}")
}

func genSetter(g *protogen.GeneratedFile, copysafeName string, fieldName string, fieldType string) {
	if *debug {
		log.Print("Setter...")
	}
	g.P("func (m *", copysafeName, ") Set", fieldName, "(value ", fieldType, ") {")
	g.P("if m != nil {")
	g.P("m.", fieldName, " = value")
	g.P("}")
	g.P("}")
}

func getDefaultValueString(field *protogen.Field) string {
	switch kind := field.Desc.Kind(); kind {
	case protoreflect.BytesKind, protoreflect.GroupKind, protoreflect.MessageKind:
		return "nil"
	case protoreflect.BoolKind:
		return "false"
	case protoreflect.EnumKind:
		return "0"
	case protoreflect.DoubleKind, protoreflect.Fixed32Kind, protoreflect.Fixed64Kind, protoreflect.FloatKind, protoreflect.Int32Kind, protoreflect.Sfixed32Kind, protoreflect.Sfixed64Kind, protoreflect.Sint32Kind, protoreflect.Sint64Kind, protoreflect.Uint32Kind, protoreflect.Uint64Kind, protoreflect.Int64Kind:
		return "0"
	case protoreflect.StringKind:
		if field.Desc.Cardinality() == protoreflect.Repeated {
			return "nil"
		}
		return `""`
	default:
		panic(fmt.Sprintf("Unrecognized type: %s", kind))
	}
}

func getActualType(g *protogen.GeneratedFile, field *protogen.Field) string {
	var fieldType string
	if field.Desc.Cardinality() == protoreflect.Repeated {
		fieldType = "[]"
	}

	if field.Desc.IsMap() {
		// check for maps first because legacy protobuf would generate "Entry" messages,
		// and for some reason the Message field is still populated
		if *debug {
			log.Printf("Map Field Detected: %+v\n\n", field.Message)
		}
		mapValueType := field.Desc.MapValue().Kind().String()
		if mapValueType == protoreflect.BytesKind.String() {
			mapValueType = "[]byte"
		} else if mapValueType == protoreflect.MessageKind.String() {
			mapValueType = "*" + string(field.Desc.MapValue().Message().Name())
		}

		fieldType = "map[" + field.Desc.MapKey().Kind().String() + "]" + mapValueType
	} else if field.Message != nil {
		if *debug {
			log.Printf("Message Field Detected: %+v\n\n", field.Message)
			log.Printf("Message Description: %+v\n\n", field.Message.Desc)
		}
		messageType, _ := getCopysafeName(g, field.Message.GoIdent)
		fieldType += "*" + messageType
	} else if field.Enum != nil {
		if *debug {
			log.Printf("Enum Field Detected: %+v\n\n", field.Enum)
			log.Printf("Enum Description: %+v\n\n", field.Enum.Desc)
		}
		enumType, _ := getCopysafeName(g, field.Enum.GoIdent)
		fieldType += enumType
	} else {
		fieldType += field.Desc.Kind().String()
	}

	return fieldType
}

func (bbsGenerateHelper) genFriendlyEnums(g *protogen.GeneratedFile, msg *protogen.Message) {
	if len(msg.Enums) > 0 {
		for _, eNuM := range msg.Enums {
			if *debug {
				log.Printf("Nested Enum: %+v\n", eNuM)
			}

			copysafeName, _ := getCopysafeName(g, eNuM.GoIdent)
			log.Printf("%s\n", copysafeName)
			g.P("type ", copysafeName, " int32")
			g.P("const (")
			for _, enumValue := range eNuM.Values {
				enumValueName := getEnumValueName(g, msg, enumValue)
				actualValue := enumValue.Desc.Number()

				g.P(enumValueName, " ", copysafeName, "=", actualValue)
			}
			g.P(")")
			g.P("// Enum value maps for ", copysafeName)
			g.P("var (")
			g.P(copysafeName, "_name = map[int32]string{")
			for _, enumValue := range eNuM.Values {
				enumValueName := enumValue.Desc.Name()
				actualValue := enumValue.Desc.Number()

				g.P(actualValue, `: "`, enumValueName, `",`)
			}
			g.P("}")
			g.P(copysafeName, "_value = map[string]int32{")
			for _, enumValue := range eNuM.Values {
				enumValueName := enumValue.Desc.Name()
				actualValue := enumValue.Desc.Number()

				g.P(`"`, enumValueName, `": `, actualValue, `,`)
			}
			g.P("}")
			g.P(")")
		}
	}
}

func getEnumValueName(g *protogen.GeneratedFile, msg *protogen.Message, enumValue *protogen.EnumValue) string {
	copysafeParentName, _ := getCopysafeName(g, msg.GoIdent)
	copysafeEnumValueName, _ := getCopysafeName(g, enumValue.GoIdent)
	customName := proto.GetExtension(enumValue.Desc.Options().(*descriptorpb.EnumValueOptions), E_BbsEnumvalueCustomname)
	log.Printf("%+v\n", customName)

	result := copysafeEnumValueName
	if len(customName.(string)) > 0 {
		result = copysafeParentName + "_" + customName.(string)
	}
	return result

}

func (bbsGenerateHelper) genToProtoMethod(g *protogen.GeneratedFile, msg *protogen.Message) {
	unsafeName := getUnsafeName(g, msg.GoIdent)
	if copysafeName, ok := getCopysafeName(g, msg.GoIdent); ok {
		g.P("func(x *", copysafeName, ") ToProto() *", unsafeName, " {")
		g.P("proto := &", unsafeName, "{")
		for _, field := range msg.Fields {
			if field.Message != nil {
				if field.Desc.Cardinality() == protoreflect.Repeated {
					fieldCopysafeName, _ := getCopysafeName(g, field.Message.GoIdent)
					if field.Desc.IsList() {
						g.P(field.GoName, ": ", fieldCopysafeName, "ProtoMap(x.", field.GoName, "),")
					} else if field.Desc.IsMap() {
						g.P(field.GoName, ": ", "x.", field.GoName, ",")
					} else {
						panic("Unrecognized Repeated field found")
					}
				} else {
					g.P(field.GoName, ": x.", field.GoName, ".ToProto(),")
				}
			} else if field.Enum != nil {
				g.P(field.GoName, ": ", field.GoIdent, "(x.", field.GoName, "),")
			} else {
				if field.Oneof != nil {
					g.P(field.GoName, ": &x.", field.GoName, ",")
				} else {
					g.P(field.GoName, ": x.", field.GoName, ",")
				}
			}
		}
		g.P("}")
		g.P("return proto")
		g.P("}")
		g.P()
	}
}

func (bbsGenerateHelper) genProtoMapMethod(g *protogen.GeneratedFile, msg *protogen.Message) {
	unsafeName := getUnsafeName(g, msg.GoIdent)
	if copysafeName, ok := getCopysafeName(g, msg.GoIdent); ok {
		g.P("func ", copysafeName, "ProtoMap(values []*", copysafeName, ") []*", unsafeName, " {")
		g.P("result := make([]*", unsafeName, ", len(values))")
		g.P("for i, val := range values {")
		g.P("result[i] = val.ToProto()")
		g.P("}")
		g.P("return result")
		g.P("}")
		g.P()
	}
}

func generateFile(plugin *protogen.Plugin, file *protogen.File) *protogen.GeneratedFile {
	filename := file.GeneratedFilenamePrefix + "_bbs.pb.go"
	g := plugin.NewGeneratedFile(filename, file.GoImportPath)
	g.P("// Code generated by protoc-gen-go-bbs. DO NOT EDIT.")
	g.P("// versions:")
	g.P("// - protoc-gen-go-bbs v", version) // version from main.go
	g.P("// - protoc            ", protocVersion(plugin))

	if file.Proto.GetOptions().GetDeprecated() {
		g.P("// ", file.Desc.Path(), " is a deprecated file.")
	} else {
		g.P("// source: ", file.Desc.Path())
	}
	g.P()
	g.P("package ", file.GoPackageName)
	g.P()
	generateFileContent(file, g)
	return g
}

func protocVersion(plugin *protogen.Plugin) string {
	v := plugin.Request.GetCompilerVersion()
	if v == nil {
		return "(unknown)"
	}
	var suffix string
	if s := v.GetSuffix(); s != "" {
		suffix = "-" + s
	}
	return fmt.Sprintf("v%d.%d.%d%s", v.GetMajor(), v.GetMinor(), v.GetPatch(), suffix)
}

var ignoredMessages []string = []string{"ProtoRoutes"}

func generateFileContent(file *protogen.File, g *protogen.GeneratedFile) {
	for _, msg := range file.Messages {
		if *debug {
			log.Printf("New Message Detected: %+v\n\n", msg)
		}

		if slices.Contains(ignoredMessages, getUnsafeName(g, msg.GoIdent)) {
			log.Printf("Ignoring message %s", msg.GoIdent)
			continue
		}
		helper.genFriendlyEnums(g, msg)
		helper.genCopysafeStruct(g, msg)
		helper.genToProtoMethod(g, msg)
		helper.genProtoMapMethod(g, msg)
	}
}
