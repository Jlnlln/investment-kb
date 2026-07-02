package markdown

import (
	"fmt"
	"strings"

	"investment-kb/internal/model"
)

// RenderSourceMetaLines renders stable source metadata for validation and traceability.
func RenderSourceMetaLines(meta model.SourceMeta) string {
	var sb strings.Builder
	if meta.SourceFile != "" {
		fmt.Fprintf(&sb, "source_file: %s  \n", meta.SourceFile)
	}
	if meta.RawHash != "" {
		fmt.Fprintf(&sb, "raw_hash: %s  \n", meta.RawHash)
	}
	if meta.CleanedHash != "" {
		fmt.Fprintf(&sb, "cleaned_hash: %s  \n", meta.CleanedHash)
	}
	if meta.RawID != "" {
		fmt.Fprintf(&sb, "raw_id: %s  \n", meta.RawID)
	}
	if meta.MaterialType != "" {
		fmt.Fprintf(&sb, "material_type: %s  \n", meta.MaterialType)
	}
	return sb.String()
}
