package toon

import (
	"bytes"
	"fmt"
	"reflect"
	"strings"
)

type encoder struct {
	buf  bytes.Buffer
	opts MarshalOptions
}

func newEncoder(opts MarshalOptions) *encoder {
	return &encoder{
		opts: opts,
	}
}

func (e *encoder) encode(v any) ([]byte, error) {
	rv := reflect.ValueOf(v)
	if err := e.encodeValue(rv, 0, ""); err != nil {
		return nil, err
	}
	return e.buf.Bytes(), nil
}

func (e *encoder) encodeValue(v reflect.Value, depth int, key string) error {
	if !v.IsValid() {
		return nil
	}

	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			if key != "" {
				e.writeIndent(depth)
				e.buf.WriteString(key)
				e.buf.WriteString(": null\n")
			}
			return nil
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.Struct:
		return e.encodeStruct(v, depth, key)
	case reflect.Map:
		return e.encodeMap(v, depth, key)
	case reflect.Slice, reflect.Array:
		return e.encodeSlice(v, depth, key)
	default:
		return e.encodePrimitive(v, depth, key)
	}
}

func (e *encoder) encodeStruct(v reflect.Value, depth int, key string) error {
	if key != "" {
		e.writeIndent(depth)
		e.buf.WriteString(key)
		e.buf.WriteString(":\n")
		depth++
	}

	t := v.Type()
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		fieldValue := v.Field(i)

		if !field.IsExported() {
			continue
		}

		name := e.getFieldName(field)
		if name == "-" {
			continue
		}

		if err := e.encodeValue(fieldValue, depth, name); err != nil {
			return err
		}
	}
	return nil
}

func (e *encoder) encodeMap(v reflect.Value, depth int, key string) error {
	if key != "" {
		e.writeIndent(depth)
		e.buf.WriteString(key)
		e.buf.WriteString(":\n")
		depth++
	}

	keys := v.MapKeys()
	for _, k := range keys {
		keyStr := fmt.Sprintf("%v", k.Interface())
		if err := e.encodeValue(v.MapIndex(k), depth, keyStr); err != nil {
			return err
		}
	}
	return nil
}

func (e *encoder) encodeSlice(v reflect.Value, depth int, key string) error {
	length := v.Len()

	if length == 0 {
		if key != "" {
			e.writeIndent(depth)
			e.buf.WriteString(key)
			e.buf.WriteString("[0]:\n")
		}
		return nil
	}

	elemType := v.Type().Elem()
	for elemType.Kind() == reflect.Ptr {
		elemType = elemType.Elem()
	}

	switch elemType.Kind() {
	case reflect.Struct:
		if e.opts.UseTabular && e.isUniformStructSlice(v) {
			return e.encodeTabularSlice(v, depth, key)
		}
		return e.encodeListSlice(v, depth, key)
	case reflect.Map:
		return e.encodeListSlice(v, depth, key)
	default:
		return e.encodePrimitiveSlice(v, depth, key)
	}
}

func (e *encoder) encodePrimitiveSlice(v reflect.Value, depth int, key string) error {
	length := v.Len()

	e.writeIndent(depth)
	if key != "" {
		e.buf.WriteString(key)
	}
	e.buf.WriteString(fmt.Sprintf("[%d]: ", length))

	for i := 0; i < length; i++ {
		if i > 0 {
			e.buf.WriteString(string(e.opts.Delimiter))
		}
		e.writePrimitiveValue(v.Index(i))
	}
	e.buf.WriteString("\n")
	return nil
}

func (e *encoder) encodeTabularSlice(v reflect.Value, depth int, key string) error {
	length := v.Len()
	if length == 0 {
		return nil
	}

	// Get first element to determine fields
	firstElem := v.Index(0)
	for firstElem.Kind() == reflect.Ptr || firstElem.Kind() == reflect.Interface {
		if firstElem.IsNil() {
			return nil
		}
		firstElem = firstElem.Elem()
	}

	fields := e.getStructFieldNames(firstElem)

	e.writeIndent(depth)
	if key != "" {
		e.buf.WriteString(key)
	}
	e.buf.WriteString(fmt.Sprintf("[%d]{%s}:\n", length, strings.Join(fields, ",")))

	for i := 0; i < length; i++ {
		elem := v.Index(i)
		for elem.Kind() == reflect.Ptr || elem.Kind() == reflect.Interface {
			elem = elem.Elem()
		}

		e.writeIndent(depth + 1)
		e.writeStructAsRow(elem)
		e.buf.WriteString("\n")
	}
	return nil
}

func (e *encoder) encodeListSlice(v reflect.Value, depth int, key string) error {
	length := v.Len()

	e.writeIndent(depth)
	if key != "" {
		e.buf.WriteString(key)
	}
	e.buf.WriteString(fmt.Sprintf("[%d]:\n", length))

	for i := 0; i < length; i++ {
		elem := v.Index(i)

		e.writeIndent(depth + 1)
		e.buf.WriteString("- ")

		// Handle the element inline or as nested
		for elem.Kind() == reflect.Ptr || elem.Kind() == reflect.Interface {
			if elem.IsNil() {
				e.buf.WriteString("null\n")
				continue
			}
			elem = elem.Elem()
		}

		switch elem.Kind() {
		case reflect.Struct:
			e.encodeListItem(elem, depth+2)
		case reflect.Map:
			e.encodeListItemMap(elem, depth+2)
		default:
			e.writePrimitiveValue(elem)
			e.buf.WriteString("\n")
		}
	}
	return nil
}

func (e *encoder) encodeListItem(v reflect.Value, depth int) error {
	t := v.Type()
	first := true

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		name := e.getFieldName(field)
		if name == "-" {
			continue
		}

		fieldValue := v.Field(i)

		if first {
			// First field on same line as -
			e.buf.WriteString(name)
			e.buf.WriteString(": ")
			e.writePrimitiveValue(fieldValue)
			e.buf.WriteString("\n")
			first = false
		} else {
			// Subsequent fields on new lines
			e.writeIndent(depth)
			e.buf.WriteString(name)
			e.buf.WriteString(": ")
			e.writePrimitiveValue(fieldValue)
			e.buf.WriteString("\n")
		}
	}
	return nil
}

func (e *encoder) encodeListItemMap(v reflect.Value, depth int) error {
	keys := v.MapKeys()
	first := true

	for _, k := range keys {
		keyStr := fmt.Sprintf("%v", k.Interface())
		val := v.MapIndex(k)

		if first {
			e.buf.WriteString(keyStr)
			e.buf.WriteString(": ")
			e.writePrimitiveValue(val)
			e.buf.WriteString("\n")
			first = false
		} else {
			e.writeIndent(depth)
			e.buf.WriteString(keyStr)
			e.buf.WriteString(": ")
			e.writePrimitiveValue(val)
			e.buf.WriteString("\n")
		}
	}
	return nil
}

func (e *encoder) encodePrimitive(v reflect.Value, depth int, key string) error {
	e.writeIndent(depth)
	if key != "" {
		e.buf.WriteString(key)
		e.buf.WriteString(": ")
	}
	e.writePrimitiveValue(v)
	e.buf.WriteString("\n")
	return nil
}

func (e *encoder) writePrimitiveValue(v reflect.Value) {
	for v.Kind() == reflect.Ptr || v.Kind() == reflect.Interface {
		if v.IsNil() {
			e.buf.WriteString("null")
			return
		}
		v = v.Elem()
	}

	switch v.Kind() {
	case reflect.String:
		s := v.String()
		if strings.ContainsAny(s, ",|\t\n") {
			e.buf.WriteString("\"")
			e.buf.WriteString(strings.ReplaceAll(s, "\"", "\\\""))
			e.buf.WriteString("\"")
		} else {
			e.buf.WriteString(s)
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		e.buf.WriteString(fmt.Sprintf("%d", v.Int()))
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		e.buf.WriteString(fmt.Sprintf("%d", v.Uint()))
	case reflect.Float32:
		e.buf.WriteString(fmt.Sprintf("%g", v.Float()))
	case reflect.Float64:
		e.buf.WriteString(fmt.Sprintf("%g", v.Float()))
	case reflect.Bool:
		e.buf.WriteString(fmt.Sprintf("%t", v.Bool()))
	default:
		e.buf.WriteString(fmt.Sprintf("%v", v.Interface()))
	}
}

func (e *encoder) writeStructAsRow(v reflect.Value) {
	t := v.Type()
	first := true

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		name := e.getFieldName(field)
		if name == "-" {
			continue
		}

		if !first {
			e.buf.WriteString(string(e.opts.Delimiter))
		}
		first = false

		e.writePrimitiveValue(v.Field(i))
	}
}

func (e *encoder) getStructFieldNames(v reflect.Value) []string {
	t := v.Type()
	var fields []string

	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		name := e.getFieldName(field)
		if name == "-" {
			continue
		}

		fields = append(fields, name)
	}
	return fields
}

func (e *encoder) getFieldName(field reflect.StructField) string {
	if tag := field.Tag.Get("toon"); tag != "" {
		parts := strings.Split(tag, ",")
		return parts[0]
	}

	if tag := field.Tag.Get("json"); tag != "" {
		parts := strings.Split(tag, ",")
		return parts[0]
	}

	name := field.Name
	if len(name) > 0 {
		return strings.ToLower(name[:1]) + name[1:]
	}
	return name
}

func (e *encoder) writeIndent(depth int) {
	for i := 0; i < depth*e.opts.Indent; i++ {
		e.buf.WriteByte(' ')
	}
}

func (e *encoder) isUniformStructSlice(v reflect.Value) bool {
	if v.Len() == 0 {
		return false
	}

	firstElem := v.Index(0)
	for firstElem.Kind() == reflect.Ptr || firstElem.Kind() == reflect.Interface {
		if firstElem.IsNil() {
			return false
		}
		firstElem = firstElem.Elem()
	}

	if firstElem.Kind() != reflect.Struct {
		return false
	}

	t := firstElem.Type()
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		kind := field.Type.Kind()
		if kind == reflect.Struct || kind == reflect.Slice || kind == reflect.Array || kind == reflect.Map {
			return false
		}
	}

	return true
}
