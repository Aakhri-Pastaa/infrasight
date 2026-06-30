// Package html renders a Document as a single self-contained, interactive HTML
// report.
//
// The UI is a React/Vite/Tailwind app built under web/ and bundled to one file
// (assets/index.html). Fonts (a subset Material Symbols) are self-hosted and
// inlined, so the report makes zero network calls and works air-gapped. This
// package only embeds that built shell, injects the vendored vis-network library
// (base64) for the graph view, and splices in the scan data as a JSON island.
package html

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"io"
	"strings"

	"github.com/Aakhri-Pastaa/infrasight/internal/output"
)

//go:embed assets/index.html
var htmlShell string

//go:embed assets/vis-network.min.js
var visNetworkJS []byte

// Render writes a self-contained interactive HTML report for the document.
func Render(w io.Writer, doc output.Document) error {
	// MarshalIndent escapes <, > and & (SetEscapeHTML on by default), so the
	// payload is safe inside the <script type="application/json"> island.
	data, err := json.MarshalIndent(doc, "", "  ")
	if err != nil {
		return err
	}
	lib := base64.StdEncoding.EncodeToString(visNetworkJS)

	out := strings.Replace(htmlShell, "__VIS_NETWORK_B64__", lib, 1)
	out = strings.Replace(out, "/*__INFRASIGHT_DATA__*/", string(data), 1)

	_, err = io.WriteString(w, out)
	return err
}
