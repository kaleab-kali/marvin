package report

import (
	"encoding/json"
	"io"

	"github.com/kaleab-kali/marvin/internal/cost"
)

func WriteJSON(w io.Writer, records []cost.Record, rules cost.WarningRules) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(BuildSummary(records, rules))
}
