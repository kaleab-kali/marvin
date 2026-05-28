package report

import (
	"encoding/json"
	"io"

	"github.com/kaleab-kali/marvin/internal/cost"
)

func WriteJSON(w io.Writer, records []cost.Record, rules cost.WarningRules) error {
	return WriteJSONSummary(w, BuildSummary(records, rules))
}

func WriteJSONSummary(w io.Writer, summary Summary) error {
	encoder := json.NewEncoder(w)
	encoder.SetIndent("", "  ")
	return encoder.Encode(summary)
}
