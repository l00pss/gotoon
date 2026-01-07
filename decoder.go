package toon

import (
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

type decoder struct {
	data  []byte
	lines []string
	pos   int
}

func newDecoder(data []byte) *decoder {
	input := string(data)
	lines := strings.Split(input, "\n")
	return &decoder{
		data:  data,
		lines: lines,
		pos:   0,
	}
}

func (d *decoder) decode(v any) error {
	rv := reflect.ValueOf(v)
	if rv.Kind() != reflect.Ptr {
		return ErrUnmarshalType
	}
	if rv.IsNil() {
		return ErrNilPointer
	}

	return d.decodeValue(rv.Elem(), 0)
}

func (d *decoder) hasMore() bool {
	for i := d.pos; i < len(d.lines); i++ {
		if strings.TrimSpace(d.lines[i]) != "" && !strings.HasPrefix(strings.TrimSpace(d.lines[i]), "#") {
			return true
		}
	}
	return false
}

func (d *decoder) currentLine() string {
	if d.pos >= len(d.lines) {
		return ""
	}
	return d.lines[d.pos]
}

func (d *decoder) advance() {
	d.pos++
}

func (d *decoder) skipEmptyLines() {
	for d.pos < len(d.lines) {
		line := strings.TrimSpace(d.lines[d.pos])
		if line != "" && !strings.HasPrefix(line, "#") {
			break
		}
		d.pos++
	}
}

func (d *decoder) getIndent(line string) int {
	count := 0
	for _, ch := range line {
		if ch == ' ' {
			count++
		} else {
			break
		}
	}
	return count
}

func (d *decoder) decodeValue(v reflect.Value, expectedIndent int) error {
	d.skipEmptyLines()
	if !d.hasMore() {
		return nil
	}

	switch v.Kind() {
	case reflect.Struct:
		return d.decodeStruct(v, expectedIndent)
	case reflect.Map:
		return d.decodeMap(v, expectedIndent)
	case reflect.Slice:
		return d.decodeSlice(v, expectedIndent)
	case reflect.Ptr:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return d.decodeValue(v.Elem(), expectedIndent)
	case reflect.Interface:
		m := make(map[string]any)
		mv := reflect.ValueOf(&m).Elem()
		if err := d.decodeMap(mv, expectedIndent); err != nil {
			return err
		}
		v.Set(mv)
		return nil
	default:
		d.skipEmptyLines()
		if !d.hasMore() {
			return nil
		}
		line := d.currentLine()
		trimmed := strings.TrimSpace(line)
		d.advance()
		return d.setPrimitiveValue(v, trimmed)
	}
}

func (d *decoder) decodeStruct(v reflect.Value, expectedIndent int) error {
	t := v.Type()
	fieldMap := make(map[string]int)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}

		name := getFieldName(field)
		if name != "-" {
			fieldMap[name] = i
		}
	}

	for d.hasMore() {
		d.skipEmptyLines()
		if !d.hasMore() {
			break
		}

		line := d.currentLine()
		indent := d.getIndent(line)

		if expectedIndent > 0 && indent < expectedIndent {
			break
		}

		trimmed := strings.TrimSpace(line)
		if !strings.Contains(trimmed, ":") {
			d.advance()
			continue
		}

		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			d.advance()
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		arrayLen, fieldNames := d.parseArrayDeclaration(key)
		if arrayLen >= 0 {
			key = d.extractKeyFromArray(key)
		}

		fieldIdx, ok := fieldMap[key]
		if !ok {
			d.advance()
			continue
		}

		fieldValue := v.Field(fieldIdx)
		d.advance()

		if arrayLen >= 0 {
			if err := d.decodeArrayField(fieldValue, arrayLen, fieldNames, value, indent); err != nil {
				return err
			}
		} else if value == "" {
			if err := d.decodeValue(fieldValue, indent+2); err != nil {
				return err
			}
		} else {
			if err := d.setPrimitiveValue(fieldValue, value); err != nil {
				return err
			}
		}
	}

	return nil
}

func (d *decoder) decodeMap(v reflect.Value, expectedIndent int) error {
	if v.IsNil() {
		v.Set(reflect.MakeMap(v.Type()))
	}

	keyType := v.Type().Key()
	elemType := v.Type().Elem()

	for d.hasMore() {
		d.skipEmptyLines()
		if !d.hasMore() {
			break
		}

		line := d.currentLine()
		indent := d.getIndent(line)

		if expectedIndent > 0 && indent < expectedIndent {
			break
		}

		trimmed := strings.TrimSpace(line)
		if !strings.Contains(trimmed, ":") {
			d.advance()
			continue
		}

		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			d.advance()
			continue
		}

		keyStr := strings.TrimSpace(parts[0])
		valueStr := strings.TrimSpace(parts[1])

		key := reflect.New(keyType).Elem()
		if err := d.setPrimitiveValue(key, keyStr); err != nil {
			return err
		}

		elem := reflect.New(elemType).Elem()
		d.advance()

		if valueStr == "" {
			if err := d.decodeValue(elem, indent+2); err != nil {
				return err
			}
		} else {
			if err := d.setPrimitiveValue(elem, valueStr); err != nil {
				return err
			}
		}

		v.SetMapIndex(key, elem)
	}

	return nil
}

func (d *decoder) decodeSlice(v reflect.Value, expectedIndent int) error {
	elemType := v.Type().Elem()
	slice := reflect.MakeSlice(v.Type(), 0, 0)

	for d.hasMore() {
		d.skipEmptyLines()
		if !d.hasMore() {
			break
		}

		line := d.currentLine()
		indent := d.getIndent(line)

		if expectedIndent > 0 && indent < expectedIndent {
			break
		}

		trimmed := strings.TrimSpace(line)
		if !strings.HasPrefix(trimmed, "- ") {
			break
		}

		// Remove "- " prefix
		itemContent := strings.TrimSpace(trimmed[2:])
		d.advance()

		elem := reflect.New(elemType).Elem()

		if elemType.Kind() == reflect.Struct {
			// For struct, parse the first field inline, then continue with nested fields
			if strings.Contains(itemContent, ":") {
				// Decode as struct with first field inline
				if err := d.decodeStructFromListItem(elem, itemContent, indent+2); err != nil {
					return err
				}
			}
		} else {
			// For primitive, set value directly
			if err := d.setPrimitiveValue(elem, itemContent); err != nil {
				return err
			}
		}

		slice = reflect.Append(slice, elem)
	}

	v.Set(slice)
	return nil
}

func (d *decoder) decodeArrayField(v reflect.Value, length int, fieldNames []string, value string, indent int) error {
	if len(fieldNames) > 0 {
		// Tabular format
		return d.decodeTabularArray(v, length, fieldNames, indent)
	} else if value != "" {
		// Inline format
		return d.decodeInlineArray(v, value)
	} else {
		// List format
		return d.decodeValue(v, indent+2)
	}
}

func (d *decoder) decodeInlineArray(v reflect.Value, value string) error {
	// Split by delimiter (comma, tab, or pipe)
	var parts []string
	if strings.Contains(value, "\t") {
		parts = strings.Split(value, "\t")
	} else if strings.Contains(value, "|") {
		parts = strings.Split(value, "|")
	} else {
		parts = strings.Split(value, ",")
	}

	elemType := v.Type().Elem()
	slice := reflect.MakeSlice(v.Type(), 0, len(parts))

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		elem := reflect.New(elemType).Elem()
		if err := d.setPrimitiveValue(elem, part); err != nil {
			return err
		}
		slice = reflect.Append(slice, elem)
	}

	v.Set(slice)
	return nil
}

func (d *decoder) decodeTabularArray(v reflect.Value, length int, fieldNames []string, indent int) error {
	elemType := v.Type().Elem()
	if elemType.Kind() != reflect.Struct {
		return fmt.Errorf("tabular arrays require struct elements")
	}

	// Build field mapping
	fieldMap := make(map[string]int)
	t := elemType
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		name := getFieldName(field)
		fieldMap[name] = i
	}

	slice := reflect.MakeSlice(v.Type(), 0, length)

	// Read tabular data
	for i := 0; i < length && d.hasMore(); i++ {
		d.skipEmptyLines()
		if !d.hasMore() {
			break
		}

		line := d.currentLine()
		if d.getIndent(line) <= indent {
			if strings.TrimSpace(line) == "" {
				d.advance()
				continue
			}
		}

		rowData := strings.TrimSpace(line)
		d.advance()

		// Split by delimiter
		var values []string
		if strings.Contains(rowData, "\t") {
			values = strings.Split(rowData, "\t")
		} else if strings.Contains(rowData, "|") {
			values = strings.Split(rowData, "|")
		} else {
			values = strings.Split(rowData, ",")
		}

		elem := reflect.New(elemType).Elem()

		// Map values to fields
		for j, fieldName := range fieldNames {
			if j < len(values) {
				if fieldIdx, ok := fieldMap[fieldName]; ok {
					fieldValue := elem.Field(fieldIdx)
					value := strings.TrimSpace(values[j])
					if err := d.setPrimitiveValue(fieldValue, value); err != nil {
						return err
					}
				}
			}
		}

		slice = reflect.Append(slice, elem)
	}

	v.Set(slice)
	return nil
}

func (d *decoder) decodeStructFromListItem(v reflect.Value, firstLine string, expectedIndent int) error {
	t := v.Type()
	fieldMap := make(map[string]int)

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if !field.IsExported() {
			continue
		}
		name := getFieldName(field)
		if name != "-" {
			fieldMap[name] = i
		}
	}

	// Parse first line
	if strings.Contains(firstLine, ":") {
		parts := strings.SplitN(firstLine, ":", 2)
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if fieldIdx, ok := fieldMap[key]; ok {
			if err := d.setPrimitiveValue(v.Field(fieldIdx), value); err != nil {
				return err
			}
		}
	}

	// Parse remaining fields from subsequent lines
	for d.hasMore() {
		d.skipEmptyLines()
		if !d.hasMore() {
			break
		}

		line := d.currentLine()
		indent := d.getIndent(line)

		if indent < expectedIndent {
			break
		}

		trimmed := strings.TrimSpace(line)
		if !strings.Contains(trimmed, ":") {
			break
		}

		parts := strings.SplitN(trimmed, ":", 2)
		if len(parts) != 2 {
			d.advance()
			continue
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])

		if fieldIdx, ok := fieldMap[key]; ok {
			if err := d.setPrimitiveValue(v.Field(fieldIdx), value); err != nil {
				return err
			}
		}

		d.advance()
	}

	return nil
}

func (d *decoder) parseArrayDeclaration(key string) (int, []string) {
	// Match patterns like: key[3], key[3,], key[3|], key[3]{field1,field2}
	re := regexp.MustCompile(`^(.+?)\[(\d+)(?:[,\t|])?\](?:\{([^}]+)\})?`)
	matches := re.FindStringSubmatch(key)
	if len(matches) == 0 {
		return -1, nil
	}

	length, _ := strconv.Atoi(matches[2])

	var fieldNames []string
	if len(matches) > 3 && matches[3] != "" {
		fields := strings.Split(matches[3], ",")
		for _, field := range fields {
			fieldNames = append(fieldNames, strings.TrimSpace(field))
		}
	}

	return length, fieldNames
}

func (d *decoder) extractKeyFromArray(key string) string {
	re := regexp.MustCompile(`^(.+?)\[`)
	matches := re.FindStringSubmatch(key)
	if len(matches) > 1 {
		return matches[1]
	}
	return key
}

func (d *decoder) setPrimitiveValue(v reflect.Value, s string) error {
	s = strings.TrimSpace(s)

	// Handle quoted strings
	if len(s) >= 2 && s[0] == '"' && s[len(s)-1] == '"' {
		s = s[1 : len(s)-1]
		s = strings.ReplaceAll(s, "\\\"", "\"")
	}

	switch v.Kind() {
	case reflect.String:
		v.SetString(s)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		i, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			return err
		}
		v.SetInt(i)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		u, err := strconv.ParseUint(s, 10, 64)
		if err != nil {
			return err
		}
		v.SetUint(u)
	case reflect.Float32, reflect.Float64:
		f, err := strconv.ParseFloat(s, 64)
		if err != nil {
			return err
		}
		v.SetFloat(f)
	case reflect.Bool:
		b, err := strconv.ParseBool(s)
		if err != nil {
			return err
		}
		v.SetBool(b)
	case reflect.Interface:
		// Try to determine type
		if s == "null" {
			v.Set(reflect.Zero(v.Type()))
		} else if i, err := strconv.ParseInt(s, 10, 64); err == nil {
			v.Set(reflect.ValueOf(i))
		} else if f, err := strconv.ParseFloat(s, 64); err == nil {
			v.Set(reflect.ValueOf(f))
		} else if b, err := strconv.ParseBool(s); err == nil {
			v.Set(reflect.ValueOf(b))
		} else {
			v.Set(reflect.ValueOf(s))
		}
	case reflect.Ptr:
		if v.IsNil() {
			v.Set(reflect.New(v.Type().Elem()))
		}
		return d.setPrimitiveValue(v.Elem(), s)
	default:
		return fmt.Errorf("unsupported type: %v", v.Kind())
	}

	return nil
}

func getFieldName(field reflect.StructField) string {
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
