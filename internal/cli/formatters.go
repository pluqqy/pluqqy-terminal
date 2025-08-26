package cli

import (
	"encoding/json"
	"fmt"
	"io"
	"strings"
	"text/tabwriter"

	"gopkg.in/yaml.v3"
)

// OutputFormat represents the output format type
type OutputFormat string

const (
	FormatText OutputFormat = "text"
	FormatJSON OutputFormat = "json"
	FormatYAML OutputFormat = "yaml"
)

// TableFormatter helps format tabular output
type TableFormatter struct {
	writer *tabwriter.Writer
}

// NewTableFormatter creates a new table formatter
func NewTableFormatter(w io.Writer) *TableFormatter {
	tw := tabwriter.NewWriter(w, 0, 0, 2, ' ', 0)
	return &TableFormatter{writer: tw}
}

// Header writes the table header
func (t *TableFormatter) Header(columns ...string) {
	fmt.Fprintln(t.writer, strings.Join(columns, "\t"))
	fmt.Fprintln(t.writer, strings.Repeat("-", 80))
}

// Row writes a table row
func (t *TableFormatter) Row(values ...string) {
	fmt.Fprintln(t.writer, strings.Join(values, "\t"))
}

// Flush writes the buffered table to output
func (t *TableFormatter) Flush() {
	t.writer.Flush()
}

// OutputResults formats and outputs results based on the specified format
func OutputResults(w io.Writer, format string, data interface{}) error {
	switch OutputFormat(format) {
	case FormatJSON:
		encoder := json.NewEncoder(w)
		encoder.SetIndent("", "  ")
		return encoder.Encode(data)

	case FormatYAML:
		yamlData, err := yaml.Marshal(data)
		if err != nil {
			return err
		}
		fmt.Fprint(w, string(yamlData))
		return nil

	case FormatText:
		// For text format, we expect the caller to have already formatted
		// the data appropriately. This is a fallback.
		fmt.Fprintf(w, "%v\n", data)
		return nil

	default:
		return fmt.Errorf("unsupported output format: %s", format)
	}
}

// FormatBytes formats byte count in human-readable format
func FormatBytes(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// TruncateString truncates a string to the specified length
func TruncateString(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	if maxLen <= 3 {
		return s[:maxLen]
	}
	return s[:maxLen-3] + "..."
}

// PadRight pads a string with spaces to the right
func PadRight(s string, length int) string {
	if len(s) >= length {
		return s
	}
	return s + strings.Repeat(" ", length-len(s))
}

// ColorizeTag returns a colored tag string (when color is enabled)
func ColorizeTag(tag string, color string) string {
	if noColor {
		return tag
	}
	// This is a simplified version - in production you'd use a proper color library
	// For now, just return the tag as-is
	return tag
}