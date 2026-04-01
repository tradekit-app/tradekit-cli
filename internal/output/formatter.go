package output

import (
	"io"
	"os"
)

type Format string

const (
	FormatTable Format = "table"
	FormatJSON  Format = "json"
	FormatCSV   Format = "csv"
)

type Formatter interface {
	Format(w io.Writer, data any) error
}

func New(format Format) Formatter {
	switch format {
	case FormatJSON:
		return &JSONFormatter{Pretty: isTTY()}
	case FormatCSV:
		return &CSVFormatter{}
	default:
		return &TableFormatter{Color: isTTY()}
	}
}

func isTTY() bool {
	fi, err := os.Stdout.Stat()
	if err != nil {
		return false
	}
	return fi.Mode()&os.ModeCharDevice != 0
}
