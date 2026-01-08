package toon

import (
	"errors"
	"fmt"
	"strings"
)

type Delimiter string

const (
	DelimiterComma Delimiter = ","
	DelimiterTab   Delimiter = "\t"
	DelimiterPipe  Delimiter = "|"
)

type MarshalOptions struct {
	Indent     int
	Delimiter  Delimiter
	UseTabular bool
}

var (
	ErrInvalidSyntax   = errors.New("toon: invalid syntax")
	ErrUnmarshalType   = errors.New("toon: cannot unmarshal into non-pointer value")
	ErrNilPointer      = errors.New("toon: cannot unmarshal into nil pointer")
	ErrUnsupportedType = errors.New("toon: unsupported type")
)

type SyntaxError struct {
	Line    int
	Column  int
	Message string
}

func (e *SyntaxError) Error() string {
	return fmt.Sprintf("toon: syntax error at line %d, column %d: %s", e.Line, e.Column, e.Message)
}

func DefaultMarshalOptions() MarshalOptions {
	return MarshalOptions{
		Indent:     2,
		Delimiter:  DelimiterComma,
		UseTabular: true,
	}
}

func Marshal(v any) ([]byte, error) {
	return MarshalWithOptions(v, DefaultMarshalOptions())
}

func MarshalWithOptions(v any, opts MarshalOptions) ([]byte, error) {
	e := newEncoder(opts)
	return e.encode(v)
}

func Unmarshal(data []byte, v any) error {
	d := newDecoder(data)
	return d.decode(v)
}

func Valid(data []byte) bool {
	input := string(data)
	lines := strings.Split(input, "\n")

	for _, line := range lines {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" || strings.HasPrefix(trimmed, "#") {
			continue
		}
		if !strings.Contains(trimmed, ":") && !strings.Contains(trimmed, "[") {
			return false
		}
	}

	return true
}
