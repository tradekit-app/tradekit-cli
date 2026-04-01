package output

import (
	"encoding/json"
	"io"
)

type JSONFormatter struct {
	Pretty bool
}

func (f *JSONFormatter) Format(w io.Writer, data any) error {
	enc := json.NewEncoder(w)
	if f.Pretty {
		enc.SetIndent("", "  ")
	}
	return enc.Encode(data)
}
