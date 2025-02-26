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
	genFromProtoMethod(g *protogen.GeneratedFile, msg *protogen.Message)
	genToProtoSliceMethod(g *protogen.GeneratedFile, msg *protogen.Message)
	genFromProtoSliceMethod(g *protogen.GeneratedFile, msg *protogen.Message)
	genMessageEnums(g *protogen.GeneratedFile, msg *protogen.Message)
	genAccessors(g *protogen.GeneratedFile, msg *protogen.Message)
	genEqual(g *protogen.GeneratedFile, msg *protogen.Message)

	genGlobalEnum(g *protogen.GeneratedFile, eNuM *protogen.Enum)
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
	return result
}

func getJsonTag(field *protogen.Field) string {
	jsonName := field.Desc.JSONName()
	jsonEmit := ",omitempty"
	if isAlwaysEmit(field) {
		jsonEmit = ""
	}
	tag := fmt.Sprintf("`json:\"%s%s\"`", jsonName, jsonEmit)
	return tag
}

func isAlwaysEmit(field *protogen.Field) bool {
	isAlwaysEmit := proto.GetExtension(field.Desc.Options().(*descriptorpb.FieldOptions), E_BbsJsonAlwaysEmit)
	return isAlwaysEmit.(bool)
}

func isMessageDeprecated(msg *protogen.Message) bool {
	options := msg.Desc.Options().(*descriptorpb.MessageOptions)
	return options.GetDeprecated()
}

func isFieldDeprecated(field *protogen.Field) bool {
	options := field.Desc.Options().(*descriptorpb.FieldOptions)
	return options.GetDeprecated()
}

func isEnumValueDeprecated(enumValue *protogen.EnumValue) bool {
	options := enumValue.Desc.Options().(*descriptorpb.EnumValueOptions)
	return options.GetDeprecated()
}

func (bbsGenerateHelper) genCopysafeStruct(g *protogen.GeneratedFile, msg *protogen.Message) {
	if copysafeName, ok := getCopysafeName(g, msg.GoIdent); ok {
		if isMessageDeprecated(msg) {
			g.P("// Deprecated: marked deprecated in ", msg.Location.SourceFile)
		}
		g.P("// Prevent copylock errors when using ", msg.GoIdent.GoName, " directly")
		g.P("type ", copysafeName, " struct {")
		for _, field := range msg.Fields {
			if isFieldDeprecated(field) {
				g.P("// Deprecated: marked deprecated in ", msg.Location.SourceFile)
			}
			if *debug {
				options := field.Desc.Options().(*descriptorpb.FieldOptions)
				log.Printf("New Field Detected: %+v\n\n", field)
				log.Printf("Field Options: %+v\n\n", options)
			}
			fieldName := getFieldName(field.GoName)
			fieldType := getActualType(g, field)
			jsonTag := getJsonTag(field)
			g.P(fieldName, " ", fieldType, " ", jsonTag)
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
			if isExcludedFromEqual(field) {
				continue
			}
			fieldName := getFieldName(field.GoName)
			if field.Desc.Cardinality() == protoreflect.Repeated {
				g.P("if this.", fieldName, " == nil {")
				g.P("if that1.", fieldName, " != nil {")
				g.P("return false")
				g.P("}")
				g.P("} else if len(this.", fieldName, ") != len(that1.", fieldName, ") {")
				g.P("return false")
				g.P("}")
				g.P("for i := range this.", fieldName, " {")
				if field.Message != nil && !field.Desc.IsMap() {
					g.P("if !this.", fieldName, "[i].Equal(that1.", fieldName, "[i]) {")
				} else if field.Desc.IsMap() && field.Desc.MapValue().Kind() == protoreflect.BytesKind {
					bytesEqual := protogen.GoIdent{GoName: "Equal", GoImportPath: "bytes"}
					g.P("if !", g.QualifiedGoIdent(bytesEqual), "(this.", fieldName, "[i], that1.", fieldName, "[i]) {")
				} else if field.Desc.IsMap() && field.Desc.MapValue().Kind() == protoreflect.MessageKind {
					g.P("if !this.", fieldName, "[i].Equal(that1.", fieldName, "[i]) {")
				} else {
					g.P("if this.", fieldName, "[i] != that1.", fieldName, "[i] {")
				}
				g.P("return false")
				g.P("}")
			} else if field.Message != nil {
				pointer := "*"
				if isByValueType(field) {
					pointer = ""
				} else {
					g.P("if this.", fieldName, " == nil {")
					g.P("if that1.", fieldName, " != nil {")
					g.P("return false")
					g.P("}")
					g.P("} else ")
				}
				g.P("if !this.", fieldName, ".Equal(", pointer, "that1.", fieldName, ") {")
				g.P("return false")
			} else {
				pointer := ""
				if field.Desc.HasOptionalKeyword() {
					pointer = "*"
					g.P("if this.", fieldName, " == nil {")
					g.P("if that1.", fieldName, " != nil {")
					g.P("return false")
					g.P("}")
					g.P("} else ")
				}
				g.P("if ", pointer, "this.", fieldName, " != ", pointer, "that1.", fieldName, " {")
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
			deprecated := isFieldDeprecated(field)

			if *debug {
				log.Printf("Generating accessors for %s...\n", fieldName)
			}

			if !isByValueType(field) {
				if deprecated {
					g.P("// Deprecated: marked deprecated in ", msg.Location.SourceFile)
				}
				genGetter(g, copysafeName, field)
			}

			if deprecated {
				g.P("// Deprecated: marked deprecated in ", msg.Location.SourceFile)
			}
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

func genGetter(g *protogen.GeneratedFile, copysafeName string, field *protogen.Field) {
	if *debug {
		log.Print("Getter...")
	}
	fieldName := getFieldName(field.GoName)
	fieldType := getActualType(g, field)
	defaultValue := getDefaultValueString(field)
	defaultReturn := "defaultValue"
	isOptional := field.Desc.HasOptionalKeyword()
	optionalCheck := ""
	if isOptional {
		defaultReturn = "&" + defaultReturn
		optionalCheck = fmt.Sprintf("&& m.%s != nil ", fieldName) //extra space intentional
		genExists(g, copysafeName, field)
	}
	g.P("func (m *", copysafeName, ") Get", fieldName, "() ", fieldType, " {")
	g.P("if m != nil ", optionalCheck, "{")
	g.P("return m.", fieldName)
	g.P("}")
	if defaultValue == "nil" {
		g.P("return nil")
	} else {
		valueType, _ := strings.CutPrefix(fieldType, "*")
		g.P("var defaultValue ", valueType)
		g.P("defaultValue = ", defaultValue)
		g.P("return ", defaultReturn)
	}
	g.P("}")
}

func genSetter(g *protogen.GeneratedFile, copysafeName string, fieldName string, fieldType string) {
	if *debug {
		log.Print("Setter...")
	}
	setValue := " = value"
	g.P("func (m *", copysafeName, ") Set", fieldName, "(value ", fieldType, ") {")
	g.P("if m != nil {")
	g.P("m.", fieldName, setValue)
	g.P("}")
	g.P("}")
}

func hasDefaultValue(field *protogen.Field) bool {
	defaultValue := proto.GetExtension(field.Desc.Options().(*descriptorpb.FieldOptions), E_BbsDefaultValue)
	return len(defaultValue.(string)) > 0
}

func getDefaultValue(field *protogen.Field) string {
	defaultValue := proto.GetExtension(field.Desc.Options().(*descriptorpb.FieldOptions), E_BbsDefaultValue)
	return defaultValue.(string)
}

func getDefaultValueString(field *protogen.Field) string {
	if hasDefaultValue(field) {
		return getDefaultValue(field)
	}

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

	if isCustomType(field) {
		customType := getCustomType(field)
		pointer := "*"
		if isByValueType(field) {
			pointer = ""
		}
		fieldType += pointer + customType
	} else if field.Desc.IsMap() {
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
		pointer := "*"
		if isByValueType(field) {
			pointer = ""
		}
		fieldType += pointer + messageType
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

func isByValueType(field *protogen.Field) bool {
	isByValueType := proto.GetExtension(field.Desc.Options().(*descriptorpb.FieldOptions), E_BbsByValue)
	return isByValueType.(bool)
}

func isExcludedFromEqual(field *protogen.Field) bool {
	isExcludedFromEqual := proto.GetExtension(field.Desc.Options().(*descriptorpb.FieldOptions), E_BbsExcludeFromEqual)
	return isExcludedFromEqual.(bool)
}

func isCustomType(field *protogen.Field) bool {
	customType := proto.GetExtension(field.Desc.Options().(*descriptorpb.FieldOptions), E_BbsCustomType)
	return len(customType.(string)) > 0
}

func getCustomType(field *protogen.Field) string {
	customType := proto.GetExtension(field.Desc.Options().(*descriptorpb.FieldOptions), E_BbsCustomType)
	return customType.(string)
}

func (bbsGenerateHelper) genGlobalEnum(g *protogen.GeneratedFile, eNuM *protogen.Enum) {
	genEnumTypeWithValues(g, eNuM, nil)
	genEnumValueMaps(g, eNuM)
	genEnumStringFunc(g, eNuM)
}

func (bbsGenerateHelper) genMessageEnums(g *protogen.GeneratedFile, msg *protogen.Message) {
	for _, eNuM := range msg.Enums {
		if *debug {
			log.Printf("Nested Enum: %+v\n", eNuM)
		}

		genEnumTypeWithValues(g, eNuM, msg)
		genEnumValueMaps(g, eNuM)
		genEnumStringFunc(g, eNuM)
	}
}

func genEnumTypeWithValues(g *protogen.GeneratedFile, eNuM *protogen.Enum, msg *protogen.Message) {
	copysafeName, _ := getCopysafeName(g, eNuM.GoIdent)
	g.P("type ", copysafeName, " int32")
	g.P("const (")
	for _, enumValue := range eNuM.Values {
		if isEnumValueDeprecated(enumValue) {
			g.P("// Deprecated: marked deprecated in proto file")
		}
		enumValueName := getEnumValueName(g, enumValue, msg)
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

func getEnumValueName(g *protogen.GeneratedFile, enumValue *protogen.EnumValue, msg *protogen.Message) string {
	copysafeEnumValueName, _ := getCopysafeName(g, enumValue.GoIdent)
	customName := proto.GetExtension(enumValue.Desc.Options().(*descriptorpb.EnumValueOptions), E_BbsEnumvalueCustomname)

	result := copysafeEnumValueName
	if len(customName.(string)) > 0 {
		if msg == nil {
			result = customName.(string)
		} else {
			copysafeParentName, _ := getCopysafeName(g, msg.GoIdent)
			result = copysafeParentName + "_" + customName.(string)
		}
	}
	return result

}

func (bbsGenerateHelper) genToProtoMethod(g *protogen.GeneratedFile, msg *protogen.Message) {
	unsafeName := getUnsafeName(g, msg.GoIdent)
	if copysafeName, ok := getCopysafeName(g, msg.GoIdent); ok {
		g.P("func(x *", copysafeName, ") ToProto() *", unsafeName, " {")
		g.P("if x == nil {")
		g.P("return nil")
		g.P("}")
		g.P()
		g.P("proto := &", unsafeName, "{")
		for _, field := range msg.Fields {
			protoFieldName := field.GoName
			if field.Message != nil {
				fieldCopysafeName, _ := getCopysafeName(g, field.Message.GoIdent)
				fieldCopysafeName = getFieldName(fieldCopysafeName)
				if field.Desc.Cardinality() == protoreflect.Repeated {
					if field.Desc.IsList() {
						g.P(protoFieldName, ": ", fieldCopysafeName, "ToProtoSlice(x.", getFieldName(protoFieldName), "),")
					} else if field.Desc.IsMap() {
						mapValueKind := field.Desc.MapValue().Kind()
						if mapValueKind == protoreflect.MessageKind {
							g.P(protoFieldName, ": ", copysafeName, getFieldName(protoFieldName), "ToProtoMap(x.", protoFieldName, "),")
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

func (bbsGenerateHelper) genFromProtoMethod(g *protogen.GeneratedFile, msg *protogen.Message) {
	unsafeName := getUnsafeName(g, msg.GoIdent)
	if copysafeName, ok := getCopysafeName(g, msg.GoIdent); ok {
		g.P("func(x *", unsafeName, ") FromProto() *", copysafeName, " {")
		g.P("if x == nil {")
		g.P("return nil")
		g.P("}")
		g.P()
		g.P("copysafe := &", copysafeName, "{")
		for _, field := range msg.Fields {
			protoFieldName := field.GoName
			if field.Message != nil {
				fieldCopysafeName, _ := getCopysafeName(g, field.Message.GoIdent)
				fieldCopysafeName = getFieldName(fieldCopysafeName)
				if field.Desc.Cardinality() == protoreflect.Repeated {
					if field.Desc.IsList() {
						g.P(protoFieldName, ": ", fieldCopysafeName, "FromProtoSlice(x.", getFieldName(protoFieldName), "),")
					} else if field.Desc.IsMap() {
						mapValueKind := field.Desc.MapValue().Kind()
						if mapValueKind == protoreflect.MessageKind {
							g.P(protoFieldName, ": ", copysafeName, getFieldName(protoFieldName), "FromProtoMap(x.", protoFieldName, "),")
						} else {
							g.P(protoFieldName, ": ", "x.", protoFieldName, ",")
						}
					} else {
						panic("Unrecognized Repeated field found")
					}
				} else {
					// note the reversal of pointer logic here compared to other isByValueType checks
					// FromProto() returns a pointer so we need to dereference that before calling FromProto() on a value type
					pointer := ""
					if isByValueType(field) {
						pointer = "*"
					}
					g.P(protoFieldName, ": ", pointer, "x.", getFieldName(protoFieldName), ".FromProto(),")
				}
			} else if field.Enum != nil {
				g.P(protoFieldName, ": ", getActualType(g, field), "(x.", protoFieldName, "),")
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
		g.P("return copysafe")
		g.P("}")
		g.P()
	}
}

func (bbsGenerateHelper) genToProtoSliceMethod(g *protogen.GeneratedFile, msg *protogen.Message) {
	unsafeName := getUnsafeName(g, msg.GoIdent)
	if copysafeName, ok := getCopysafeName(g, msg.GoIdent); ok {
		g.P("func ", copysafeName, "ToProtoSlice(values []*", copysafeName, ") []*", unsafeName, " {")
		g.P("if values == nil {")
		g.P("return nil")
		g.P("}")
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

					g.P("func ", copysafeName, getFieldName(field.GoName), "ToProtoMap(values ", fieldType, ") ", protoFieldType, " {")
					g.P("if values == nil {")
					g.P("return nil")
					g.P("}")
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

func (bbsGenerateHelper) genFromProtoSliceMethod(g *protogen.GeneratedFile, msg *protogen.Message) {
	unsafeName := getUnsafeName(g, msg.GoIdent)
	if copysafeName, ok := getCopysafeName(g, msg.GoIdent); ok {
		g.P("func ", copysafeName, "FromProtoSlice(values []*", unsafeName, ") []*", copysafeName, " {")
		g.P("if values == nil {")
		g.P("return nil")
		g.P("}")
		g.P("result := make([]*", copysafeName, ", len(values))")
		g.P("for i, val := range values {")
		g.P("result[i] = val.FromProto()")
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

					g.P("func ", copysafeName, getFieldName(field.GoName), "FromProtoMap(values ", protoFieldType, ") ", fieldType, " {")
					g.P("if values == nil {")
					g.P("return nil")
					g.P("}")
					g.P("result := make(map[", mapKeyType, "]*", valueFieldType, ", len(values))")
					g.P("for i, val := range values {")
					g.P("result[i] = val.FromProto()")
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
var ignoredEnums []string = []string{}

func generateFileContent(file *protogen.File, g *protogen.GeneratedFile) {
	for _, eNuM := range file.Enums {
		if *debug {
			log.Printf("New Enum Detected: %+v\n\n", eNuM)
		}

		if slices.Contains(ignoredMessages, getUnsafeName(g, eNuM.GoIdent)) {
			log.Printf("\tIgnoring enum %s", eNuM.Desc.Name())
			continue
		}

		helper.genGlobalEnum(g, eNuM)
	}

	for _, msg := range file.Messages {
		if *debug {
			log.Printf("New Message Detected: %+v\n\n", msg)
		}

		if slices.Contains(ignoredMessages, getUnsafeName(g, msg.GoIdent)) {
			log.Printf("\tIgnoring message %s", msg.Desc.Name())
			continue
		}
		helper.genMessageEnums(g, msg)
		helper.genCopysafeStruct(g, msg)
		helper.genToProtoMethod(g, msg)
		helper.genFromProtoMethod(g, msg)
		helper.genToProtoSliceMethod(g, msg)
		helper.genFromProtoSliceMethod(g, msg)
	}
}
