package output

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"github.com/olekukonko/tablewriter"
)

type Format string

const (
	FormatJSON  Format = "json"
	FormatTable Format = "table"
)

type Result struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
	Message string      `json:"message"`
}

type Printer struct {
	format Format
	writer io.Writer
}

func NewPrinter(format Format) *Printer {
	return &Printer{
		format: format,
		writer: os.Stdout,
	}
}

func (p *Printer) PrintJSON(data interface{}) error {
	var output []byte
	var err error

	switch v := data.(type) {
	case json.RawMessage:
		var parsed interface{}
		if err := json.Unmarshal(v, &parsed); err != nil {
			output = v
		} else {
			output, err = json.MarshalIndent(parsed, "", "  ")
			if err != nil {
				return err
			}
		}
	case []byte:
		var parsed interface{}
		if err := json.Unmarshal(v, &parsed); err != nil {
			output = v
		} else {
			output, err = json.MarshalIndent(parsed, "", "  ")
			if err != nil {
				return err
			}
		}
	default:
		output, err = json.MarshalIndent(data, "", "  ")
		if err != nil {
			return err
		}
	}

	fmt.Fprintln(p.writer, string(output))
	return nil
}

func (p *Printer) PrintTable(headers []string, rows [][]string) {
	table := tablewriter.NewWriter(p.writer)
	table.Header(headers)
	table.Bulk(rows)
	table.Render()
}

func (p *Printer) Print(data interface{}, headers []string, toRows func(interface{}) [][]string) error {
	if p.format == FormatTable && headers != nil && toRows != nil {
		rows := toRows(data)
		p.PrintTable(headers, rows)
		return nil
	}
	return p.PrintJSON(data)
}

func PrintResult(data interface{}, err error) {
	var result Result
	if err != nil {
		result = Result{Success: false, Data: nil, Message: err.Error()}
	} else {
		result = Result{Success: true, Data: data, Message: ""}
	}
	out, _ := json.MarshalIndent(result, "", "  ")
	fmt.Fprintln(os.Stdout, string(out))
}

func PrintError(format string, args ...interface{}) {
	fmt.Fprintf(os.Stderr, "Error: "+format+"\n", args...)
}

func PrintSuccess(format string, args ...interface{}) {
	fmt.Fprintf(os.Stdout, format+"\n", args...)
}
