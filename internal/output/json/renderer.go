// Package json renders a Document as schema-tagged, pretty-printed JSON.
package json

import (
	"encoding/json"
	"io"

	"github.com/Aakhri-Pastaa/infrasight/internal/output"
)

// Render writes the document as indented JSON.
func Render(w io.Writer, doc output.Document) error {
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	return enc.Encode(doc)
}
