package generate

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/dragstor/awesome-reliability-tools/internal/schema"
)

func testCatalog() *schema.Catalog {
	return &schema.Catalog{
		Features: []schema.FeatureDef{
			{ID: "http_monitoring", Name: "HTTP", Group: "monitoring", Description: "d"},
			{ID: "incident_management", Name: "Incident", Group: "incident_management", Description: "d"},
			{ID: "api", Name: "API", Group: "general", Description: "d"},
		},
	}
}

func TestNormalizeSortsByName(t *testing.T) {
	tools := []schema.Tool{
		{ID: "z", Name: "Zeta", Homepage: "https://z.com", Description: "z", Categories: []string{"monitoring"}, Deployment: schema.Deployment{Hosted: true}},
		{ID: "a", Name: "Alpha", Homepage: "https://a.com", Description: "a", Categories: []string{"monitoring"}, Deployment: schema.Deployment{Hosted: true}},
	}
	nc := Normalize(testCatalog(), tools)
	if len(nc.Tools) != 2 {
		t.Fatalf("expected 2 tools, got %d", len(nc.Tools))
	}
	if nc.Tools[0].Name != "Alpha" || nc.Tools[1].Name != "Zeta" {
		t.Fatalf("expected alphabetical order, got %q then %q", nc.Tools[0].Name, nc.Tools[1].Name)
	}
}

func TestNormalizeCategoryAwareDefaults(t *testing.T) {
	tools := []schema.Tool{
		{ID: "x", Name: "X", Homepage: "https://x.com", Description: "x", Categories: []string{"monitoring"}, Deployment: schema.Deployment{Hosted: true}},
	}
	nc := Normalize(testCatalog(), tools)
	got := nc.Tools[0].Features
	if got["http_monitoring"].Value != schema.Unknown {
		t.Fatalf("in-category missing claim should be unknown, got %q", got["http_monitoring"].Value)
	}
	if got["incident_management"].Value != schema.NotApplicable {
		t.Fatalf("out-of-category missing claim should be not_applicable, got %q", got["incident_management"].Value)
	}
	if got["api"].Value != schema.Unknown {
		t.Fatalf("general missing claim should be unknown, got %q", got["api"].Value)
	}
}

func TestMarshalCatalogDeterministic(t *testing.T) {
	nc := Normalize(testCatalog(), []schema.Tool{
		{ID: "b", Name: "Beta", Homepage: "https://b.com", Description: "b", Categories: []string{"monitoring"}, Deployment: schema.Deployment{Hosted: true}},
	})
	first, err := marshalCatalog(nc)
	if err != nil {
		t.Fatalf("first marshal: %v", err)
	}
	second, err := marshalCatalog(nc)
	if err != nil {
		t.Fatalf("second marshal: %v", err)
	}
	if !bytes.Equal(first, second) {
		t.Fatal("JSON output is not deterministic")
	}
	var decoded schema.NormalizedCatalog
	if err := json.Unmarshal(first, &decoded); err != nil {
		t.Fatalf("decoded JSON invalid: %v", err)
	}
}

func TestWriteJSONCreatesNestedDirectories(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "deep", "nested", "catalog.json")
	nc := Normalize(testCatalog(), nil)
	if err := WriteJSON(nc, path); err != nil {
		t.Fatalf("WriteJSON: %v", err)
	}
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("expected file at %s: %v", path, err)
	}
}

func TestAllFromEmptyOutputDir(t *testing.T) {
	root := t.TempDir()
	readmeTmpl := filepath.Join(root, "README.md.tmpl")
	if err := os.WriteFile(readmeTmpl, []byte("# Tools: {{len .Tools}}\n"), 0o644); err != nil {
		t.Fatalf("setup readme tmpl: %v", err)
	}
	siteTmpl := filepath.Join(root, "index.html.tmpl")
	if err := os.WriteFile(siteTmpl, []byte("<html><script id=\"catalog-data\" type=\"application/json\">{{.CatalogJSON}}</script></html>"), 0o644); err != nil {
		t.Fatalf("setup site tmpl: %v", err)
	}
	siteDir := filepath.Join(root, "site")
	if err := os.MkdirAll(filepath.Join(siteDir, "assets", "favicons"), 0o755); err != nil {
		t.Fatalf("setup site dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(siteDir, "main.js"), []byte("// js"), 0o644); err != nil {
		t.Fatalf("setup main.js: %v", err)
	}
	if err := os.WriteFile(filepath.Join(siteDir, "assets", "favicons", "x.png"), []byte("png"), 0o644); err != nil {
		t.Fatalf("setup favicon: %v", err)
	}
	distDir := filepath.Join(root, "fresh", "dist")
	readmeOut := filepath.Join(root, "out", "README.md")

	tools := []schema.Tool{
		{ID: "x", Name: "X", Homepage: "https://x.com", Description: "x", Categories: []string{"monitoring"}, Deployment: schema.Deployment{Hosted: true}},
	}
	opts := Options{
		ReadmeTmpl: readmeTmpl,
		ReadmeOut:  readmeOut,
		SiteTmpl:   siteTmpl,
		DistDir:    distDir,
		SiteDir:    siteDir,
	}
	if _, err := All(testCatalog(), tools, opts); err != nil {
		t.Fatalf("All from empty output dir: %v", err)
	}
	if _, err := os.Stat(readmeOut); err != nil {
		t.Fatalf("expected README at %s: %v", readmeOut, err)
	}
	for _, p := range []string{"index.html", filepath.Join("data", "catalog.json"), "main.js", filepath.Join("assets", "favicons", "x.png")} {
		if _, err := os.Stat(filepath.Join(distDir, p)); err != nil {
			t.Fatalf("expected %s: %v", p, err)
		}
	}

	// The cached favicon must be referenced from the normalized catalog.
	catalogData, err := os.ReadFile(filepath.Join(distDir, "data", "catalog.json"))
	if err != nil {
		t.Fatalf("reading catalog: %v", err)
	}
	if !strings.Contains(string(catalogData), `"favicon": "x.png"`) {
		t.Fatalf("catalog does not reference the cached favicon: %s", catalogData)
	}
}

func TestAttachFavicons(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "x.png"), []byte("png"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	nc := Normalize(testCatalog(), []schema.Tool{
		{ID: "x", Name: "X", Homepage: "https://x.com", Description: "x", Categories: []string{"monitoring"}, Deployment: schema.Deployment{Hosted: true}},
		{ID: "y", Name: "Y", Homepage: "https://y.com", Description: "y", Categories: []string{"monitoring"}, Deployment: schema.Deployment{Hosted: true}},
	})
	AttachFavicons(nc, dir)
	if nc.Tools[0].Favicon != "x.png" {
		t.Fatalf("tool x favicon = %q, want x.png", nc.Tools[0].Favicon)
	}
	if nc.Tools[1].Favicon != "" {
		t.Fatalf("tool y should have no favicon, got %q", nc.Tools[1].Favicon)
	}
}

func TestEmbeddedCatalogEscapesScriptTags(t *testing.T) {
	nc := Normalize(testCatalog(), []schema.Tool{
		{
			ID: "x", Name: "X", Homepage: "https://x.com", Description: "x",
			Categories: []string{"monitoring"}, Deployment: schema.Deployment{Hosted: true},
			Features: map[string]schema.FeatureClaim{
				"http_monitoring": {
					Value:      schema.Supported,
					VerifiedAt: "2026-09-24",
					Sources:    []string{"https://x.com/</script><img src=x onerror=alert(1)>"},
				},
			},
		},
	})
	b, err := marshalCatalog(nc)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if bytes.Contains(b, []byte("</script>")) {
		t.Fatal("embedded catalog JSON must escape </script> to prevent breaking out of the script tag")
	}
	if !strings.Contains(string(b), `\u003c`) {
		t.Fatal("expected HTML-sensitive characters to be escaped by encoding/json")
	}
}

// TestGeneratedSiteIsProjectPathSafe renders the real site template and checks
// that every asset reference is relative, so the site works when served from a
// GitHub Pages project path such as /awesome-reliability-tools/.
func TestGeneratedSiteIsProjectPathSafe(t *testing.T) {
	root := t.TempDir()
	distDir := filepath.Join(root, "dist")
	readmeOut := filepath.Join(root, "README.md")

	tools := []schema.Tool{
		{ID: "x", Name: "X", Homepage: "https://x.com", Description: "x", Categories: []string{"monitoring"}, Deployment: schema.Deployment{Hosted: true}},
	}
	opts := Options{
		ReadmeTmpl: filepath.Join("..", "..", "templates", "README.md.tmpl"),
		ReadmeOut:  readmeOut,
		SiteTmpl:   filepath.Join("..", "..", "templates", "index.html.tmpl"),
		DistDir:    distDir,
		SiteDir:    filepath.Join("..", "..", "site"),
	}
	if _, err := All(testCatalog(), tools, opts); err != nil {
		t.Fatalf("All: %v", err)
	}

	html, err := os.ReadFile(filepath.Join(distDir, "index.html"))
	if err != nil {
		t.Fatalf("reading generated index: %v", err)
	}
	page := string(html)
	for _, want := range []string{`href="./styles.css"`, `src="./main.js"`} {
		if !strings.Contains(page, want) {
			t.Fatalf("generated index missing %q", want)
		}
	}
	rootRelative := regexp.MustCompile(`(?i)(href|src)="/[^/]`)
	if m := rootRelative.FindString(page); m != "" {
		t.Fatalf("generated index has root-relative asset reference %q; must be relative for project-site hosting", m)
	}

	js, err := os.ReadFile(filepath.Join(distDir, "main.js"))
	if err != nil {
		t.Fatalf("reading copied main.js: %v", err)
	}
	if !strings.Contains(string(js), "'./data/catalog.json'") {
		t.Fatal("copied main.js must fetch catalog data via a relative path")
	}
}

func TestCopyStatic(t *testing.T) {
	src := t.TempDir()
	dst := filepath.Join(t.TempDir(), "nested", "dst")
	if err := os.WriteFile(filepath.Join(src, "index.html"), []byte("<html></html>"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.Mkdir(filepath.Join(src, "css"), 0o755); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := os.WriteFile(filepath.Join(src, "css", "style.css"), []byte("body{}"), 0o644); err != nil {
		t.Fatalf("setup: %v", err)
	}
	if err := CopyStatic(src, dst); err != nil {
		t.Fatalf("copy static: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "index.html")); err != nil {
		t.Fatalf("index.html missing: %v", err)
	}
	if _, err := os.Stat(filepath.Join(dst, "css", "style.css")); err != nil {
		t.Fatalf("css/style.css missing: %v", err)
	}
}
