package validate

import (
	"strings"
	"testing"
	"time"

	"github.com/dragstor/awesome-reliability-tools/internal/schema"
)

// testDate tracks "today" so date tests do not break as real time passes.
var testDate = time.Now().UTC().Format("2006-01-02")

func validCatalog() *schema.Catalog {
	return &schema.Catalog{
		Features: []schema.FeatureDef{
			{ID: "http_monitoring", Name: "HTTP monitoring", Group: "monitoring", Description: "d"},
			{ID: "incident_management", Name: "Incident management", Group: "incident_management", Description: "d"},
		},
	}
}

func validTool() schema.Tool {
	return schema.Tool{
		ID:          "example",
		Name:        "Example",
		Homepage:    "https://example.com",
		Description: "An example tool.",
		Categories:  []string{"monitoring"},
		Deployment:  schema.Deployment{Hosted: true},
		OpenSource:  false,
		Features: map[string]schema.FeatureClaim{
			"http_monitoring": {Value: schema.Supported, VerifiedAt: testDate, Sources: []string{"https://example.com/docs"}},
		},
	}
}

func TestCatalogValid(t *testing.T) {
	if res := Catalog(validCatalog()); !res.OK() {
		t.Fatalf("expected valid catalog, got errors:\n%s", res.Error())
	}
}

func TestCatalogDuplicateFeatureID(t *testing.T) {
	c := &schema.Catalog{
		Features: []schema.FeatureDef{
			{ID: "http_monitoring", Name: "HTTP monitoring", Group: "monitoring", Description: "d"},
			{ID: "http_monitoring", Name: "HTTP monitoring 2", Group: "monitoring", Description: "d"},
		},
	}
	res := Catalog(c)
	if res.OK() {
		t.Fatal("expected duplicate feature id error")
	}
	if len(res.Errors) != 1 {
		t.Fatalf("expected 1 error, got %d: %v", len(res.Errors), res.Errors)
	}
}

func TestCatalogUnknownGroup(t *testing.T) {
	c := &schema.Catalog{
		Features: []schema.FeatureDef{{ID: "x", Name: "X", Group: "alerting", Description: "d"}},
	}
	if Catalog(c).OK() {
		t.Fatal("expected unknown group error")
	}
}

func TestCatalogGeneralGroupAllowed(t *testing.T) {
	c := &schema.Catalog{
		Features: []schema.FeatureDef{{ID: "api", Name: "API", Group: "general", Description: "d"}},
	}
	if res := Catalog(c); !res.OK() {
		t.Fatalf("expected general group to be valid, got: %s", res.Error())
	}
}

func TestToolsValid(t *testing.T) {
	if res := Tools([]schema.Tool{validTool()}, validCatalog()); !res.OK() {
		t.Fatalf("expected valid tools, got errors:\n%s", res.Error())
	}
}

func TestToolsDuplicateID(t *testing.T) {
	a := validTool()
	b := validTool()
	b.Name = "Other"
	tools := []schema.Tool{a, b}
	res := Tools(tools, validCatalog())
	if res.OK() {
		t.Fatal("expected duplicate id error")
	}
	if !strings.Contains(res.Error(), "duplicate id") {
		t.Fatalf("expected duplicate id message, got: %s", res.Error())
	}
}

func TestToolsMissingDeployment(t *testing.T) {
	tool := validTool()
	tool.Deployment = schema.Deployment{}
	res := Tools([]schema.Tool{tool}, validCatalog())
	if res.OK() {
		t.Fatal("expected deployment error")
	}
}

func TestToolsDuplicateCategory(t *testing.T) {
	tool := validTool()
	tool.Categories = []string{"monitoring", "monitoring"}
	res := Tools([]schema.Tool{tool}, validCatalog())
	if res.OK() {
		t.Fatal("expected duplicate category error")
	}
}

func TestToolsInvalidStatus(t *testing.T) {
	tool := validTool()
	tool.Status = schema.Lifecycle("zombie")
	res := Tools([]schema.Tool{tool}, validCatalog())
	if res.OK() {
		t.Fatal("expected invalid status error")
	}
}

func TestToolsOpenSourceRequiresLicenseAndRepo(t *testing.T) {
	tool := validTool()
	tool.Deployment = schema.Deployment{SelfHosted: true}
	tool.OpenSource = true
	res := Tools([]schema.Tool{tool}, validCatalog())
	if res.OK() {
		t.Fatal("expected open source errors")
	}
	if len(res.Errors) != 2 {
		t.Fatalf("expected 2 errors, got %d: %v", len(res.Errors), res.Errors)
	}
}

func TestFeatureClaimSupportedRequiresSource(t *testing.T) {
	res := FeatureClaim("tool x", "http_monitoring", schema.FeatureClaim{Value: schema.Supported, VerifiedAt: testDate})
	if res.OK() {
		t.Fatal("expected missing source error for supported claim")
	}
}

func TestFeatureClaimUnsupportedRequiresSource(t *testing.T) {
	res := FeatureClaim("tool x", "http_monitoring", schema.FeatureClaim{Value: schema.Unsupported, VerifiedAt: testDate})
	if res.OK() {
		t.Fatal("expected missing source error for unsupported claim")
	}
}

func TestFeatureClaimUnknownDoesNotRequireSource(t *testing.T) {
	res := FeatureClaim("tool x", "http_monitoring", schema.FeatureClaim{Value: schema.Unknown, VerifiedAt: testDate})
	if !res.OK() {
		t.Fatalf("unknown claim should not require a source, got: %s", res.Error())
	}
}

func TestFeatureClaimMalformedDate(t *testing.T) {
	res := FeatureClaim("tool x", "http_monitoring", schema.FeatureClaim{Value: schema.Unknown, VerifiedAt: "24-09-2026"})
	if res.OK() {
		t.Fatal("expected malformed date error")
	}
}

func TestFeatureClaimFutureDate(t *testing.T) {
	res := FeatureClaim("tool x", "http_monitoring", schema.FeatureClaim{Value: schema.Unknown, VerifiedAt: "2999-01-01"})
	if res.OK() {
		t.Fatal("expected future date error")
	}
}

func TestFeatureClaimDuplicateSource(t *testing.T) {
	res := FeatureClaim("tool x", "http_monitoring", schema.FeatureClaim{
		Value:      schema.Supported,
		VerifiedAt: testDate,
		Sources:    []string{"https://example.com", "https://example.com"},
	})
	if res.OK() {
		t.Fatal("expected duplicate source error")
	}
}

func TestToolsUnknownFeature(t *testing.T) {
	tool := validTool()
	tool.Features = map[string]schema.FeatureClaim{
		"tls_monitoring": {Value: schema.Supported, VerifiedAt: testDate, Sources: []string{"https://x.com"}},
	}
	res := Tools([]schema.Tool{tool}, validCatalog())
	if res.OK() {
		t.Fatal("expected unknown feature error")
	}
}

func TestToolsBadURL(t *testing.T) {
	tool := validTool()
	tool.Homepage = "not-a-url"
	res := Tools([]schema.Tool{tool}, validCatalog())
	if res.OK() {
		t.Fatal("expected bad url error")
	}
}

func TestToolsJavascriptURLRejected(t *testing.T) {
	tool := validTool()
	tool.Homepage = "javascript:alert(1)"
	res := Tools([]schema.Tool{tool}, validCatalog())
	if res.OK() {
		t.Fatal("expected javascript: url to be rejected")
	}
}

func TestToolsURLWhitespaceRejected(t *testing.T) {
	tool := validTool()
	tool.Homepage = "https://example.com/a b"
	res := Tools([]schema.Tool{tool}, validCatalog())
	if res.OK() {
		t.Fatal("expected whitespace in url to be rejected")
	}
}

func TestToolsUnknownCategory(t *testing.T) {
	tool := validTool()
	tool.Categories = []string{"alerting"}
	res := Tools([]schema.Tool{tool}, validCatalog())
	if res.OK() {
		t.Fatal("expected unknown category error")
	}
}
