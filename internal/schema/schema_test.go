package schema

import "testing"

func TestResolveFeatureExplicitWins(t *testing.T) {
	tool := Tool{
		Categories: []string{"monitoring"},
		Features: map[string]FeatureClaim{
			"http_monitoring": {Value: Unsupported, VerifiedAt: "2026-09-24", Sources: []string{"https://x.com"}},
		},
	}
	got := tool.ResolveFeature(FeatureDef{ID: "http_monitoring", Group: "monitoring"})
	if got.Value != Unsupported {
		t.Fatalf("explicit claim should win, got %q", got.Value)
	}
}

func TestResolveFeatureOwnedCategoryMissingIsUnknown(t *testing.T) {
	tool := Tool{Categories: []string{"monitoring"}}
	got := tool.ResolveFeature(FeatureDef{ID: "dns_monitoring", Group: "monitoring"})
	if got.Value != Unknown {
		t.Fatalf("missing claim in owned category should be unknown, got %q", got.Value)
	}
}

func TestResolveFeatureUnownedCategoryIsNotApplicable(t *testing.T) {
	tool := Tool{Categories: []string{"incident_management"}}
	got := tool.ResolveFeature(FeatureDef{ID: "http_monitoring", Group: "monitoring"})
	if got.Value != NotApplicable {
		t.Fatalf("missing claim outside categories should be not_applicable, got %q", got.Value)
	}
}

func TestResolveFeatureGeneralIsUnknown(t *testing.T) {
	tool := Tool{Categories: []string{"monitoring"}}
	got := tool.ResolveFeature(FeatureDef{ID: "api", Group: GeneralFeatureGroup})
	if got.Value != Unknown {
		t.Fatalf("missing general claim should be unknown, got %q", got.Value)
	}
}

func TestLifecycleDisplay(t *testing.T) {
	if got := LifecycleActive.Display(); got != "" {
		t.Fatalf("active should have empty display, got %q", got)
	}
	if got := LifecycleArchived.Display(); got != "Archived" {
		t.Fatalf("archived display = %q", got)
	}
	if !LifecycleDeprecated.IsValid() || Lifecycle("zombie").IsValid() {
		t.Fatal("lifecycle validation is wrong")
	}
}
