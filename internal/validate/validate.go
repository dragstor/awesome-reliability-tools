package validate

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/dragstor/awesome-reliability-tools/internal/schema"
)

var (
	slugRE     = regexp.MustCompile(`^[a-z0-9]+(?:[_-][a-z0-9]+)*$`)
	spaceRE    = regexp.MustCompile(`\s`)
	controlRE  = regexp.MustCompile(`[\x00-\x1f\x7f]`)
	dateLayout = "2006-01-02"
)

func todayUTC() time.Time {
	return time.Now().UTC().Truncate(24 * time.Hour)
}

// Result accumulates validation problems.
type Result struct {
	Errors []string
}

// Addf records a formatted error.
func (r *Result) Addf(format string, args ...interface{}) {
	r.Errors = append(r.Errors, fmt.Sprintf(format, args...))
}

// OK returns true when no errors were recorded.
func (r *Result) OK() bool {
	return len(r.Errors) == 0
}

// Error implements the error interface for a non-OK result.
func (r *Result) Error() string {
	if r.OK() {
		return ""
	}
	return strings.Join(r.Errors, "\n")
}

// Catalog validates the central feature catalog.
func Catalog(c *schema.Catalog) *Result {
	res := &Result{}
	if c == nil {
		res.Addf("catalog is nil")
		return res
	}
	seen := make(map[string]struct{})
	for _, f := range c.Features {
		id := strings.TrimSpace(f.ID)
		if id == "" {
			res.Addf("feature has empty id")
			continue
		}
		if _, ok := seen[f.ID]; ok {
			res.Addf("duplicate feature id %q", f.ID)
		}
		seen[f.ID] = struct{}{}
		if !slugRE.MatchString(f.ID) {
			res.Addf("feature %q: id must be lowercase letters, numbers, hyphens or underscores", f.ID)
		}
		if strings.TrimSpace(f.Name) == "" {
			res.Addf("feature %q has empty name", f.ID)
		}
		if strings.TrimSpace(f.Description) == "" {
			res.Addf("feature %q has empty description", f.ID)
		}
		if !schema.IsValidFeatureGroup(f.Group) {
			res.Addf("feature %q: unknown group %q; must be one of %v or %q", f.ID, f.Group, schema.KnownCategories, schema.GeneralFeatureGroup)
		}
	}
	return res
}

// Tools validates a list of tools against a feature catalog.
func Tools(tools []schema.Tool, catalog *schema.Catalog) *Result {
	res := &Result{}
	catalogRes := Catalog(catalog)
	if !catalogRes.OK() {
		res.Errors = append(res.Errors, catalogRes.Errors...)
	}
	featureMap := catalog.FeatureMap()

	seenIDs := make(map[string]int)
	for i, t := range tools {
		prefix := fmt.Sprintf("tool[%d]", i)
		if t.ID == "" {
			res.Addf("%s: missing id", prefix)
		} else {
			prefix = fmt.Sprintf("tool %q", t.ID)
			if other, ok := seenIDs[t.ID]; ok {
				res.Addf("%s: duplicate id (first seen at index %d)", prefix, other)
			} else {
				seenIDs[t.ID] = i
			}
			if !slugRE.MatchString(t.ID) {
				res.Addf("%s: id %q must be lowercase letters, numbers, hyphens or underscores", prefix, t.ID)
			}
		}
		if strings.TrimSpace(t.Name) == "" {
			res.Addf("%s: missing name", prefix)
		}
		if t.Status != "" && !t.Status.IsValid() {
			res.Addf("%s: invalid status %q; must be one of %q, %q or %q", prefix, t.Status, schema.LifecycleActive, schema.LifecycleDeprecated, schema.LifecycleArchived)
		}
		if err := validateURL(t.Homepage, "homepage"); err != nil {
			res.Addf("%s: %v", prefix, err)
		}
		if strings.TrimSpace(t.Description) == "" {
			res.Addf("%s: missing description", prefix)
		}
		if len(t.Categories) == 0 {
			res.Addf("%s: at least one category is required", prefix)
		}
		seenCats := make(map[string]struct{}, len(t.Categories))
		for _, c := range t.Categories {
			if !schema.IsValidCategory(c) {
				res.Addf("%s: unknown category %q; must be one of %v", prefix, c, schema.KnownCategories)
				continue
			}
			if _, ok := seenCats[c]; ok {
				res.Addf("%s: duplicate category %q", prefix, c)
			}
			seenCats[c] = struct{}{}
		}
		if !t.Deployment.Hosted && !t.Deployment.SelfHosted {
			res.Addf("%s: at least one of deployment.hosted or deployment.self_hosted must be true", prefix)
		}
		if t.OpenSource {
			if t.Repository == "" {
				res.Addf("%s: open_source is true but repository is missing", prefix)
			}
			if t.License == "" {
				res.Addf("%s: open_source is true but license is missing", prefix)
			}
		} else if t.License != "" {
			res.Addf("%s: license %q is set but open_source is false", prefix, t.License)
		}
		if t.Repository != "" {
			if err := validateURL(t.Repository, "repository"); err != nil {
				res.Addf("%s: %v", prefix, err)
			}
		}
		seenSources := make(map[string]struct{}, len(t.Sources))
		for _, s := range t.Sources {
			if err := validateURL(s, "source"); err != nil {
				res.Addf("%s: %v", prefix, err)
				continue
			}
			if _, ok := seenSources[s]; ok {
				res.Addf("%s: duplicate source %q", prefix, s)
			}
			seenSources[s] = struct{}{}
		}
		for fid, c := range t.Features {
			if _, ok := featureMap[fid]; !ok {
				res.Addf("%s: unknown feature %q", prefix, fid)
				continue
			}
			if err := FeatureClaim(prefix, fid, c); err != nil {
				res.Errors = append(res.Errors, err.Errors...)
			}
		}
	}
	return res
}

// FeatureClaim validates a single feature claim.
func FeatureClaim(toolPrefix, featureID string, c schema.FeatureClaim) *Result {
	res := &Result{}
	if !c.Value.IsValid() {
		res.Addf("%s: feature %q has invalid value %q; must be one of %v", toolPrefix, featureID, c.Value, schema.ValidFeatureStates)
	}
	if c.VerifiedAt == "" {
		res.Addf("%s: feature %q is missing verified_at", toolPrefix, featureID)
	} else if err := validateDate(c.VerifiedAt); err != nil {
		res.Addf("%s: feature %q %v", toolPrefix, featureID, err)
	}
	// Supported and unsupported are factual assertions and need evidence.
	if (c.Value == schema.Supported || c.Value == schema.Unsupported) && len(c.Sources) == 0 {
		res.Addf("%s: feature %q is marked %s but has no sources", toolPrefix, featureID, c.Value)
	}
	seenSources := make(map[string]struct{}, len(c.Sources))
	for _, s := range c.Sources {
		if err := validateURL(s, "source"); err != nil {
			res.Addf("%s: feature %q %v", toolPrefix, featureID, err)
			continue
		}
		if _, ok := seenSources[s]; ok {
			res.Addf("%s: feature %q has duplicate source %q", toolPrefix, featureID, s)
		}
		seenSources[s] = struct{}{}
	}
	return res
}

func validateDate(raw string) error {
	d, err := time.Parse(dateLayout, raw)
	if err != nil {
		return fmt.Errorf("has malformed verified_at %q (expected YYYY-MM-DD)", raw)
	}
	if d.After(todayUTC()) {
		return fmt.Errorf("has future verified_at %q; use the date you actually verified the claim", raw)
	}
	return nil
}

func validateURL(raw, role string) error {
	if raw == "" {
		return fmt.Errorf("%s URL is empty", role)
	}
	if spaceRE.MatchString(raw) || controlRE.MatchString(raw) {
		return fmt.Errorf("%s URL %q contains whitespace or control characters", role, raw)
	}
	u, err := url.Parse(raw)
	if err != nil {
		return fmt.Errorf("%s URL %q is malformed: %w", role, raw, err)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return fmt.Errorf("%s URL %q must use http or https", role, raw)
	}
	if u.Host == "" {
		return fmt.Errorf("%s URL %q is missing a host", role, raw)
	}
	return nil
}
