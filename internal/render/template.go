package render

import (
	"bytes"
	"fmt"
	"os"
	"sort"
	"text/template"

	"github.com/dragstor/awesome-reliability-tools/internal/schema"
)

// MarkdownStateIcon returns a table-friendly marker for a feature state.
func MarkdownStateIcon(s schema.FeatureState) string {
	return s.MarkdownIcon()
}

// HasCategory reports whether the tool participates in a category.
func HasCategory(t schema.Tool, cat string) bool {
	return t.HasCategory(cat)
}

// FeaturesByGroup returns feature definitions in a group, sorted by name.
func FeaturesByGroup(catalog *schema.Catalog, group string) []schema.FeatureDef {
	var out []schema.FeatureDef
	for _, f := range catalog.Features {
		if f.Group == group {
			out = append(out, f)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// ToolsByCategory returns tools in a category, sorted by name.
func ToolsByCategory(tools []schema.Tool, cat string) []schema.Tool {
	var out []schema.Tool
	for _, t := range tools {
		if t.HasCategory(cat) {
			out = append(out, t)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Name < out[j].Name })
	return out
}

// Claim returns a tool's resolved claim for a feature.
func Claim(t schema.Tool, f schema.FeatureDef) schema.FeatureClaim {
	return t.ResolveFeature(f)
}

// YesNo returns Yes/No based on a boolean.
func YesNo(v bool) string {
	if v {
		return "Yes"
	}
	return "No"
}

// StatusSuffix returns " (Archived)" / " (Deprecated)" for non-active tools, or "".
func StatusSuffix(s schema.Lifecycle) string {
	if label := s.Display(); label != "" {
		return " (" + label + ")"
	}
	return ""
}

// FuncMap exposes helpers used by templates.
func FuncMap() template.FuncMap {
	return template.FuncMap{
		"stateIcon":       MarkdownStateIcon,
		"hasCategory":     HasCategory,
		"featuresByGroup": FeaturesByGroup,
		"toolsByCategory": ToolsByCategory,
		"claim":           Claim,
		"yesNo":           YesNo,
		"categoryDisplay": schema.CategoryDisplay,
		"statusSuffix":    StatusSuffix,
	}
}

// RenderSite renders the static site template with embedded catalog JSON.
func RenderSite(tmplPath string, catalogJSON []byte) ([]byte, error) {
	text, err := os.ReadFile(tmplPath)
	if err != nil {
		return nil, fmt.Errorf("reading site template: %w", err)
	}
	tmpl, err := template.New("index.html.tmpl").Parse(string(text))
	if err != nil {
		return nil, fmt.Errorf("parsing site template: %w", err)
	}
	data := struct {
		CatalogJSON string
	}{
		CatalogJSON: string(catalogJSON),
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("executing site template: %w", err)
	}
	return buf.Bytes(), nil
}

// RenderREADME renders the README template against the dataset.
func RenderREADME(tmplPath string, catalog *schema.Catalog, tools []schema.Tool) ([]byte, error) {
	text, err := os.ReadFile(tmplPath)
	if err != nil {
		return nil, fmt.Errorf("reading README template: %w", err)
	}
	tmpl, err := template.New("README.md.tmpl").Funcs(FuncMap()).Parse(string(text))
	if err != nil {
		return nil, fmt.Errorf("parsing README template: %w", err)
	}

	data := struct {
		Catalog    *schema.Catalog
		Tools      []schema.Tool
		Categories []string
	}{
		Catalog:    catalog,
		Tools:      tools,
		Categories: schema.KnownCategories,
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return nil, fmt.Errorf("executing README template: %w", err)
	}
	return buf.Bytes(), nil
}
