package template

import (
	"bytes"
	"fmt"
	"testing"
)

func TestParseBytes(t *testing.T) {
	type arg struct {
		Bytes  []byte
		Values any
	}
	type output struct {
		Bytes    []byte
		HasError bool
	}
	type test struct {
		input arg
		want  output
	}

	tests := []test{
		{
			input: arg{Bytes: []byte("{{.hello}}"), Values: map[string]any{"hello": "world"}},
			want:  output{Bytes: []byte("world"), HasError: false},
		},
	}

	for _, t := range tests {
		b, err := ParseBytes(t.input.Bytes, t.input.Values)
		got := output{Bytes: b, HasError: err != nil}
		if got.HasError != t.want.HasError {
			fmt.Printf("want: %+v\t got: %+v\n", t.want, got)
		}

		if bytes.Compare(b, t.want.Bytes) != 0 {
			fmt.Printf("want: %+v\t got: %+v\n", t.want, got)
		}
	}
}
