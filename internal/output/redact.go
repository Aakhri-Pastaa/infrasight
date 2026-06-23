package output

import "github.com/Aakhri-Pastaa/infrasight/internal/graph"

// RedactedValue replaces sensitive fields in a redacted document.
const RedactedValue = "[redacted]"

// sensitiveMetaKeys are node metadata fields scrubbed by Redact.
var sensitiveMetaKeys = []string{"hostname", "bindAddress", "version", "versionId"}

// Redact returns a copy of doc with host-identifying and version detail removed,
// so the report is safe to share (paste into a ticket, send to a vendor). It
// scrubs the hostname, every node version, and bind addresses, and marks the
// document Redacted. Node metadata maps are cloned, so the original Document is
// left untouched.
//
// Scope is intentionally conservative and documented (hostname, versions, bind
// addresses); package and service names, ports, and topology are preserved so
// the graph remains useful. Widen via sensitiveMetaKeys if needed.
func Redact(doc Document) Document {
	doc.Scan.Hostname = RedactedValue
	doc.Redacted = true

	nodes := make([]graph.Node, len(doc.Nodes))
	for i, n := range doc.Nodes {
		if n.Version != "" {
			n.Version = RedactedValue
		}
		if n.Metadata != nil {
			cloned := make(map[string]any, len(n.Metadata))
			for k, v := range n.Metadata {
				cloned[k] = v
			}
			for _, key := range sensitiveMetaKeys {
				if _, ok := cloned[key]; ok {
					cloned[key] = RedactedValue
				}
			}
			n.Metadata = cloned
		}
		nodes[i] = n
	}
	doc.Nodes = nodes
	return doc
}
