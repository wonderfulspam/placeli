package export

import (
	"encoding/json"
	"fmt"
	"io"

	"github.com/user/placeli/internal/models"
)

func ExportJSON(places []*models.Place, writer io.Writer) error {
	encoder := json.NewEncoder(writer)
	encoder.SetIndent("", "  ")

	if err := encoder.Encode(places); err != nil {
		return fmt.Errorf("failed to encode JSON: %w", err)
	}

	return nil
}
