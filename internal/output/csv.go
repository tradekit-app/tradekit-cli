package output

import (
	"encoding/csv"
	"fmt"
	"io"
	"reflect"
)

type CSVFormatter struct{}

func (f *CSVFormatter) Format(w io.Writer, data any) error {
	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	// Handle slices
	if val.Kind() == reflect.Slice {
		if val.Len() == 0 {
			return nil
		}
		return f.formatSlice(w, val)
	}

	// Single struct: wrap in slice
	return f.formatSlice(w, reflect.ValueOf([]any{data}))
}

func (f *CSVFormatter) formatSlice(w io.Writer, val reflect.Value) error {
	writer := csv.NewWriter(w)
	defer writer.Flush()

	if val.Len() == 0 {
		return nil
	}

	// Get headers from first element's struct fields
	first := val.Index(0)
	if first.Kind() == reflect.Ptr {
		first = first.Elem()
	}
	if first.Kind() == reflect.Interface {
		first = first.Elem()
	}
	if first.Kind() != reflect.Struct {
		return fmt.Errorf("CSV output requires struct data")
	}

	t := first.Type()
	var headers []string
	for i := 0; i < t.NumField(); i++ {
		headers = append(headers, t.Field(i).Name)
	}
	if err := writer.Write(headers); err != nil {
		return err
	}

	// Write rows
	for i := 0; i < val.Len(); i++ {
		item := val.Index(i)
		if item.Kind() == reflect.Ptr {
			item = item.Elem()
		}
		if item.Kind() == reflect.Interface {
			item = item.Elem()
		}

		var row []string
		for j := 0; j < item.NumField(); j++ {
			row = append(row, fmt.Sprintf("%v", item.Field(j).Interface()))
		}
		if err := writer.Write(row); err != nil {
			return err
		}
	}
	return nil
}
