package schema

// FeatureState represents the four possible claim states for a capability.
type FeatureState string

const (
	Supported     FeatureState = "supported"
	Unsupported   FeatureState = "unsupported"
	Unknown       FeatureState = "unknown"
	NotApplicable FeatureState = "not_applicable"
)

// ValidFeatureStates is the canonical set of feature-state values.
var ValidFeatureStates = []FeatureState{Supported, Unsupported, Unknown, NotApplicable}

// IsValid reports whether s is a recognized feature state.
func (s FeatureState) IsValid() bool {
	switch s {
	case Supported, Unsupported, Unknown, NotApplicable:
		return true
	}
	return false
}

// Display returns a human-readable label for the state.
func (s FeatureState) Display() string {
	switch s {
	case Supported:
		return "Supported"
	case Unsupported:
		return "Unsupported"
	case Unknown:
		return "Unknown / unverified"
	case NotApplicable:
		return "Not applicable"
	}
	return string(s)
}

// MarkdownIcon returns a short visual marker for Markdown tables.
func (s FeatureState) MarkdownIcon() string {
	switch s {
	case Supported:
		return "Yes"
	case Unsupported:
		return "No"
	case Unknown:
		return "?"
	case NotApplicable:
		return "—"
	}
	return string(s)
}

// Lifecycle describes whether a tool is actively developed, deprecated, or archived.
type Lifecycle string

const (
	LifecycleActive     Lifecycle = "active"
	LifecycleDeprecated Lifecycle = "deprecated"
	LifecycleArchived   Lifecycle = "archived"
)

// IsValidLifecycle reports whether l is a recognized lifecycle value.
func (l Lifecycle) IsValid() bool {
	switch l {
	case LifecycleActive, LifecycleDeprecated, LifecycleArchived:
		return true
	}
	return false
}

// Display returns a human-readable label, or "" when the tool is active.
func (l Lifecycle) Display() string {
	switch l {
	case LifecycleDeprecated:
		return "Deprecated"
	case LifecycleArchived:
		return "Archived"
	}
	return ""
}

// KnownCategories lists the lifecycle categories a tool can participate in.
// These are the only values accepted for Tool.Categories and are also the
// feature groups that map one-to-one onto the lifecycle stages. The extra
// "general" feature group holds cross-cutting capabilities.
var KnownCategories = []string{"monitoring", "incident_management", "on_call", "status_page"}

// GeneralFeatureGroup holds features that are not tied to a single lifecycle stage.
const GeneralFeatureGroup = "general"

// IsValidCategory reports whether c is a known category slug.
func IsValidCategory(c string) bool {
	for _, k := range KnownCategories {
		if k == c {
			return true
		}
	}
	return false
}

// IsValidFeatureGroup reports whether g is a known category or the general group.
func IsValidFeatureGroup(g string) bool {
	return g == GeneralFeatureGroup || IsValidCategory(g)
}

// CategoryDisplay returns a human-readable category name.
func CategoryDisplay(c string) string {
	switch c {
	case "monitoring":
		return "Monitoring"
	case "incident_management":
		return "Incident management"
	case "on_call":
		return "On-call"
	case "status_page":
		return "Status pages"
	case GeneralFeatureGroup:
		return "General"
	}
	return c
}

// FeatureDef is a capability that can be claimed by a tool.
type FeatureDef struct {
	ID          string `yaml:"id" json:"id"`
	Name        string `yaml:"name" json:"name"`
	Group       string `yaml:"group" json:"group"`
	Description string `yaml:"description" json:"description"`
}

// FeatureClaim is a tool's claim about a specific capability.
type FeatureClaim struct {
	Value      FeatureState `yaml:"value"`
	VerifiedAt string       `yaml:"verified_at"`
	Sources    []string     `yaml:"sources,omitempty"`
	Notes      string       `yaml:"notes,omitempty"`
}

// Deployment describes how a tool is consumed.
type Deployment struct {
	Hosted     bool `yaml:"hosted" json:"hosted"`
	SelfHosted bool `yaml:"self_hosted" json:"self_hosted"`
}

// Tool is a single entry in the directory.
type Tool struct {
	ID          string                  `yaml:"id"`
	Name        string                  `yaml:"name"`
	Status      Lifecycle               `yaml:"status,omitempty"`
	Homepage    string                  `yaml:"homepage"`
	Description string                  `yaml:"description"`
	Categories  []string                `yaml:"categories"`
	Deployment  Deployment              `yaml:"deployment"`
	OpenSource  bool                    `yaml:"open_source"`
	Repository  string                  `yaml:"repository,omitempty"`
	License     string                  `yaml:"license,omitempty"`
	Features    map[string]FeatureClaim `yaml:"features,omitempty"`
	Sources     []string                `yaml:"sources,omitempty"`
	Notes       string                  `yaml:"notes,omitempty"`
}

// HasCategory reports whether the tool participates in a category.
func (t *Tool) HasCategory(cat string) bool {
	for _, c := range t.Categories {
		if c == cat {
			return true
		}
	}
	return false
}

// ResolveFeature returns the tool's claim for a feature.
//
// Explicit claims always win. When a feature has no claim:
//   - features in a category the tool does not participate in resolve to
//     not_applicable, so a pure incident-response tool does not show dozens of
//     meaningless "unknown" monitoring cells;
//   - general cross-cutting features and features inside an owned category
//     resolve to unknown (not yet verified).
func (t *Tool) ResolveFeature(f FeatureDef) FeatureClaim {
	if c, ok := t.Features[f.ID]; ok {
		return c
	}
	if f.Group == GeneralFeatureGroup || t.HasCategory(f.Group) {
		return FeatureClaim{Value: Unknown}
	}
	return FeatureClaim{Value: NotApplicable}
}

// LastVerified returns the most recent verified_at date across all claims, or "" if none.
func (t *Tool) LastVerified() string {
	var latest string
	for _, c := range t.Features {
		if c.VerifiedAt > latest {
			latest = c.VerifiedAt
		}
	}
	return latest
}

// Catalog holds the central feature definitions.
type Catalog struct {
	Features []FeatureDef `yaml:"features"`
}

// FeatureMap returns a map of feature ID to definition.
func (c *Catalog) FeatureMap() map[string]FeatureDef {
	m := make(map[string]FeatureDef, len(c.Features))
	for _, f := range c.Features {
		m[f.ID] = f
	}
	return m
}

// NormalizedTool is the shape exported to the static website.
type NormalizedTool struct {
	ID           string                       `json:"id"`
	Name         string                       `json:"name"`
	Status       Lifecycle                    `json:"status,omitempty"`
	Homepage     string                       `json:"homepage"`
	Description  string                       `json:"description"`
	Categories   []string                     `json:"categories"`
	Deployment   Deployment                   `json:"deployment"`
	OpenSource   bool                         `json:"open_source"`
	Repository   string                       `json:"repository,omitempty"`
	License      string                       `json:"license,omitempty"`
	Features     map[string]NormalizedFeature `json:"features"`
	Sources      []string                     `json:"sources,omitempty"`
	Notes        string                       `json:"notes,omitempty"`
	LastVerified string                       `json:"last_verified"`
	Favicon      string                       `json:"favicon,omitempty"`
}

// NormalizedFeature is the claim shape exported to the website.
type NormalizedFeature struct {
	Value      FeatureState `json:"value"`
	VerifiedAt string       `json:"verified_at,omitempty"`
	Sources    []string     `json:"sources,omitempty"`
	Notes      string       `json:"notes,omitempty"`
}

// NormalizedCatalog is the top-level payload consumed by the website.
type NormalizedCatalog struct {
	Note     string           `json:"_note,omitempty"`
	Features []FeatureDef     `json:"features"`
	Tools    []NormalizedTool `json:"tools"`
}
