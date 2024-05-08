package main

import (
	"fmt"
	"log"
	"slices"
	"strings"
	"text/template"

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
	genEqual(g *protogen.GeneratedFile, msg *protogen.Message)
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

func getFieldName(goName string) string {
	result := goName
	// if strings.Contains(strings.ToLower(result), "lrp") {
	// 	result = strings.Replace(goName, "Lrp", "LRP", -1)
	// }
	return result
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
			fieldName := getFieldName(field.GoName)
			fieldType := getActualType(g, field)
			g.P(fieldName, " ", fieldType)
		}
		g.P("}")

		helper.genEqual(g, msg)
		helper.genAccessors(g, msg)
	}
}

type EqualFunc struct {
	CopysafeName string
}

type EqualField struct {
	FieldName string
}

func (bbsGenerateHelper) genEqual(g *protogen.GeneratedFile, msg *protogen.Message) {
	if copysafeName, ok := getCopysafeName(g, msg.GoIdent); ok {
		if copysafeName == "Routes" {
			log.Printf("FOUNDROUTES: %+v\n", msg)
		}

		equalBuilder := new(strings.Builder)
		equal, err := template.New("equal").Parse(
			`
	if that == nil {
		return this == nil
	}

	that1, ok := that.(*{{.CopysafeName}})
	if !ok {
		that2, ok := that.({{.CopysafeName}})
		if ok {
			that1 = &that2
		} else {
			return false
		}
	}

	if that1 == nil {
		return this == nil
	} else if this == nil {
		return false
	}
	`)
		if err != nil {
			panic(err)
		}
		equal.Execute(equalBuilder, EqualFunc{CopysafeName: copysafeName})

		g.P("func (this *", copysafeName, ") Equal(that interface{}) bool {")
		g.P(equalBuilder.String())
		g.P()
		for _, field := range msg.Fields {
			fieldName := getFieldName(field.GoName)
			if field.Desc.Cardinality() == protoreflect.Repeated {
				g.P("if len(this.", fieldName, ") != len(that1.", fieldName, ") {")
				g.P("return false")
				g.P("}")
				g.P("for i := range this.", fieldName, " {")
				if field.Message != nil && !field.Desc.IsMap() {
					g.P("if !this.", fieldName, "[i].Equal(that1.", fieldName, "[i]) {")
				} else if field.Desc.IsMap() && field.Desc.MapValue().Kind() == protoreflect.BytesKind {
					bytesEqual := protogen.GoIdent{GoName: "Equal", GoImportPath: "bytes"}
					g.P("if !", g.QualifiedGoIdent(bytesEqual), "(this.", fieldName, "[i], that1.", fieldName, "[i]) {")
				} else {
					g.P("if this.", fieldName, "[i] != that1.", fieldName, "[i] {")
				}
				g.P("return false")
				g.P("}")
			} else if field.Message != nil {
				pointer := ""
				if fieldName == "Routes" {
					pointer = "*"
					log.Printf("Adding dereference because of Routes")
				}
				g.P("if !this.", fieldName, ".Equal(", pointer, "that1.", fieldName, ") {")
				g.P("return false")
			} else {
				g.P("if this.", fieldName, " != that1.", fieldName, " {")
				g.P("return false")
			}
			g.P("}")
		}
		g.P("return true")
		g.P("}")
	}
}

func (bbsGenerateHelper) genAccessors(g *protogen.GeneratedFile, msg *protogen.Message) {
	if copysafeName, ok := getCopysafeName(g, msg.GoIdent); ok {
		for _, field := range msg.Fields {
			fieldName := getFieldName(field.GoName)
			fieldType := getActualType(g, field)
			options := field.Desc.Options().(*descriptorpb.FieldOptions)

			if *debug {
				log.Printf("Generating accessors for %s...\n", fieldName)
			}

			if options.GetDeprecated() {
				g.P("// DEPRECATED: DO NOT USE")
			}
			genGetter(g, copysafeName, field) //fieldName, fieldType, defaultValue)
			genSetter(g, copysafeName, fieldName, fieldType)
		}
	}
}

func genExists(g *protogen.GeneratedFile, copysafeName string, field *protogen.Field) {
	if *debug {
		log.Print("Exists...")
	}
	fieldName := getFieldName(field.GoName)
	g.P("func (m *", copysafeName, ") ", fieldName, "Exists() bool {")
	g.P("return m != nil && m.", fieldName, " != nil")
	g.P("}")
}

func genGetter(g *protogen.GeneratedFile, copysafeName string, field *protogen.Field) { //fieldName string, fieldType string, defaultValue string) {
	if *debug {
		log.Print("Getter...")
	}
	fieldName := getFieldName(field.GoName)
	fieldType := getActualType(g, field)
	defaultValue := getDefaultValueString(field)
	isOptional := field.Desc.HasOptionalKeyword()
	optionalCheck := ""
	if isOptional {
		defaultValue = "nil"
		optionalCheck = fmt.Sprintf("&& m.%s != nil ", fieldName) //extra space intentional
		genExists(g, copysafeName, field)
	}
	g.P("func (m *", copysafeName, ") Get", fieldName, "() ", fieldType, " {")
	g.P("if m != nil ", optionalCheck, "{")
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
	if field.Desc.Cardinality() == protoreflect.Repeated {
		return "nil"
	}

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
		mapValueKind := field.Desc.MapValue().Kind()
		mapValueType := mapValueKind.String()
		if mapValueKind == protoreflect.BytesKind {
			mapValueType = "[]byte"
		} else if mapValueKind == protoreflect.MessageKind {
			valueField := field.Desc.MapValue().Message().FullName()
			rawGoIdent := strings.Split(string(valueField), ".")
			valueFieldType, _ := strings.CutPrefix(rawGoIdent[1], *prefix)
			mapValueType = "*" + valueFieldType
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
		optional := ""
		if field.Desc.HasOptionalKeyword() {
			optional = "*"
		}
		fieldType += optional + field.Desc.Kind().String()
	}

	return fieldType
}

func (bbsGenerateHelper) genFriendlyEnums(g *protogen.GeneratedFile, msg *protogen.Message) {
	for _, eNuM := range msg.Enums {
		if *debug {
			log.Printf("Nested Enum: %+v\n", eNuM)
		}

		genEnumTypeWithValues(g, msg, eNuM)
		genEnumValueMaps(g, eNuM)
		genEnumStringFunc(g, eNuM)
	}
}

func genEnumTypeWithValues(g *protogen.GeneratedFile, msg *protogen.Message, eNuM *protogen.Enum) {
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
}

func genEnumValueMaps(g *protogen.GeneratedFile, eNuM *protogen.Enum) {
	copysafeName, _ := getCopysafeName(g, eNuM.GoIdent)
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

func genEnumStringFunc(g *protogen.GeneratedFile, eNuM *protogen.Enum) {
	copysafeName, _ := getCopysafeName(g, eNuM.GoIdent)
	strconvItoa := protogen.GoIdent{GoName: "Itoa", GoImportPath: "strconv"}
	g.P("func (m ", copysafeName, ") String() string {")
	g.P("s, ok :=", copysafeName, "_name[int32(m)]")
	g.P("if ok {")
	g.P("return s")
	g.P("}")
	g.P("return ", g.QualifiedGoIdent(strconvItoa), "(int(m))")
	g.P("}")
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
			protoFieldName := field.GoName
			if field.Message != nil {
				fieldCopysafeName, _ := getCopysafeName(g, field.Message.GoIdent)
				fieldCopysafeName = getFieldName(fieldCopysafeName)
				if field.Desc.Cardinality() == protoreflect.Repeated {
					if field.Desc.IsList() {
						g.P(protoFieldName, ": ", fieldCopysafeName, "ProtoMap(x.", getFieldName(protoFieldName), "),")
					} else if field.Desc.IsMap() {
						mapValueKind := field.Desc.MapValue().Kind()
						if mapValueKind == protoreflect.MessageKind {
							g.P(protoFieldName, ": ", copysafeName, getFieldName(protoFieldName), "ProtoMap(x.", protoFieldName, "),")
						} else {
							g.P(protoFieldName, ": ", "x.", protoFieldName, ",")
						}
					} else {
						panic("Unrecognized Repeated field found")
					}
				} else {
					g.P(protoFieldName, ": x.", getFieldName(protoFieldName), ".ToProto(),")
				}
			} else if field.Enum != nil {
				g.P(protoFieldName, ": ", *prefix, getActualType(g, field), "(x.", protoFieldName, "),")
			} else {
				// we weren't using oneof correctly, so we're not going to support it
				// if field.Oneof != nil {
				// 	g.P(protoFieldName, ": &x.", protoFieldName, ",")
				// } else {
				g.P(protoFieldName, ": x.", protoFieldName, ",")
				// }
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

		for _, field := range msg.Fields {
			if field.Desc.IsMap() {
				mapValueKind := field.Desc.MapValue().Kind()
				if mapValueKind == protoreflect.MessageKind {
					valueField := field.Desc.MapValue().Message().FullName()
					rawGoIdent := strings.Split(string(valueField), ".")
					protoValueFieldType := rawGoIdent[1]
					valueFieldType, _ := strings.CutPrefix(protoValueFieldType, *prefix)
					mapValueType := "*" + valueFieldType
					mapKeyType := field.Desc.MapKey().Kind().String()
					fieldType := "map[" + mapKeyType + "]" + mapValueType
					protoMapValueType := "*" + protoValueFieldType
					protoFieldType := "map[" + mapKeyType + "]" + protoMapValueType

					g.P("func ", copysafeName, getFieldName(field.GoName), "ProtoMap(values ", fieldType, ") ", protoFieldType, " {")
					g.P("result := make(map[", mapKeyType, "]*", protoValueFieldType, ", len(values))")
					g.P("for i, val := range values {")
					g.P("result[i] = val.ToProto()")
					g.P("}")
					g.P("return result")
					g.P("}")
					g.P()
				}
			}
		}
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
