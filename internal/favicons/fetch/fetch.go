// Package fetch discovers and downloads favicons over the network.
//
// It is separate from the parent favicons package so that the generator and its
// tests never link the discovery library or perform network access.
package fetch

import (
	"context"
	"fmt"
	"time"

	"github.com/kodehat/favifetch"
)

// maxImageBytes caps how large a downloaded favicon may be.
const maxImageBytes = 1 << 20 // 1 MiB

// extByFormat maps a detected image format to the extension we store it under.
// Only real image formats are mapped; anything else is rejected.
var extByFormat = map[favifetch.DetectedFormat]string{
	favifetch.FormatSVG:  "svg",
	favifetch.FormatPNG:  "png",
	favifetch.FormatWebP: "webp",
	favifetch.FormatICO:  "ico",
	favifetch.FormatJPEG: "jpg",
	favifetch.FormatGIF:  "gif",
}

// Fetch discovers and downloads the best favicon for a homepage.
//
// Safety: only http/https URLs, private/loopback addresses blocked, a 10s
// timeout, at most 5 redirects, a 1 MiB response cap, and no third-party
// favicon API fallback. The stored extension comes from the detected image
// format, never from a remote URL suffix or header.
func Fetch(ctx context.Context, homepage string) (string, []byte, error) {
	result, err := favifetch.Fetch(ctx, homepage,
		favifetch.WithBlockPrivateIPs(true),
		favifetch.WithFallbackAPI(false),
		favifetch.WithMaxImageSize(maxImageBytes),
		favifetch.WithMaxRedirects(5),
		favifetch.WithTimeout(10*time.Second),
		favifetch.WithUserAgent("awesome-reliability-tools-favicons/1.0 (+https://github.com/dragstor/awesome-reliability-tools)"),
	)
	if err != nil {
		return "", nil, err
	}
	if len(result.Data) == 0 {
		return "", nil, fmt.Errorf("empty favicon data")
	}
	if len(result.Data) > maxImageBytes {
		return "", nil, fmt.Errorf("favicon exceeds %d bytes", maxImageBytes)
	}
	ext, ok := extByFormat[result.Format]
	if !ok {
		return "", nil, fmt.Errorf("unsupported favicon format %q", result.Format)
	}
	return ext, result.Data, nil
}
