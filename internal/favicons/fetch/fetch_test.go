package fetch

import (
	"testing"

	"github.com/dragstor/awesome-reliability-tools/internal/favicons"
)

func TestExtByFormatIsSupported(t *testing.T) {
	if len(extByFormat) == 0 {
		t.Fatal("expected the format map to be populated")
	}
	for format, ext := range extByFormat {
		if !favicons.AllowedExt(ext) {
			t.Errorf("format %v maps to unsupported extension %q", format, ext)
		}
	}
}
