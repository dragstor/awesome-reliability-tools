// Package favicons holds the offline cache logic for tool favicons: locating an
// existing cached icon and writing a new one atomically.
//
// It deliberately has no network dependency. Fetching lives in the
// favicons/fetch subpackage, so the generator binary never links the favicon
// discovery library and `make generate` stays deterministic and offline.
package favicons

import (
	"fmt"
	"os"
	"path/filepath"
)

// Order is the deterministic lookup order for cached favicon extensions.
// It is used both when locating an existing file and when naming new ones, so a
// tool always resolves to the same asset.
var Order = []string{"svg", "png", "webp", "ico", "jpg", "gif"}

// AllowedExt reports whether ext is a supported cached favicon extension.
func AllowedExt(ext string) bool {
	for _, e := range Order {
		if e == ext {
			return true
		}
	}
	return false
}

// Find returns the cached favicon filename for a tool ID, if one exists.
func Find(dir, toolID string) (string, bool) {
	for _, ext := range Order {
		name := toolID + "." + ext
		info, err := os.Stat(filepath.Join(dir, name))
		if err == nil && !info.IsDir() && info.Size() > 0 {
			return name, true
		}
	}
	return "", false
}

// Save writes favicon bytes atomically as <toolID>.<ext>.
//
// It writes to a temporary file in the same directory and renames it into
// place, so an interrupted or failed download can never corrupt an existing
// favicon. Failed lookups never reach this function.
func Save(dir, toolID, ext string, data []byte) error {
	if !AllowedExt(ext) {
		return fmt.Errorf("unsupported favicon extension %q", ext)
	}
	if len(data) == 0 {
		return fmt.Errorf("refusing to write empty favicon")
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("creating favicon directory: %w", err)
	}
	tmp, err := os.CreateTemp(dir, toolID+".*.tmp")
	if err != nil {
		return fmt.Errorf("creating temp file: %w", err)
	}
	tmpName := tmp.Name()
	if _, err := tmp.Write(data); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("writing temp file: %w", err)
	}
	if err := tmp.Chmod(0o644); err != nil {
		tmp.Close()
		os.Remove(tmpName)
		return fmt.Errorf("setting temp file mode: %w", err)
	}
	if err := tmp.Close(); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("closing temp file: %w", err)
	}
	if err := os.Rename(tmpName, filepath.Join(dir, toolID+"."+ext)); err != nil {
		os.Remove(tmpName)
		return fmt.Errorf("renaming temp file: %w", err)
	}
	return nil
}
