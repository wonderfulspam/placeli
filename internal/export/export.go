package export

import (
	"fmt"
	"io"
	"strings"

	"github.com/user/placeli/internal/models"
)

type Format string

const (
	FormatCSV      Format = "csv"
	FormatGeoJSON  Format = "geojson"
	FormatJSON     Format = "json"
	FormatMarkdown Format = "markdown"
)

func Export(places []*models.Place, format Format, writer io.Writer) error {
	switch strings.ToLower(string(format)) {
	case string(FormatCSV):
		return ExportCSV(places, writer)
	case string(FormatGeoJSON):
		return ExportGeoJSON(places, writer)
	case string(FormatJSON):
		return ExportJSON(places, writer)
	case string(FormatMarkdown), "md":
		return ExportMarkdown(places, writer)
	default:
		return fmt.Errorf("unsupported export format: %s", format)
	}
}

func ValidateFormat(format string) error {
	switch strings.ToLower(format) {
	case string(FormatCSV), string(FormatGeoJSON), string(FormatJSON), string(FormatMarkdown), "md":
		return nil
	default:
		return fmt.Errorf("unsupported format '%s'. Supported formats: csv, geojson, json, markdown", format)
	}
}

func GetSupportedFormats() []string {
	return []string{
		string(FormatCSV),
		string(FormatGeoJSON),
		string(FormatJSON),
		string(FormatMarkdown),
	}
}
