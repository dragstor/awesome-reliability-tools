package favicons

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

func TestAllowedExt(t *testing.T) {
	for _, ext := range Order {
		if !AllowedExt(ext) {
			t.Errorf("expected %q to be allowed", ext)
		}
	}
	for _, ext := range []string{"", "exe", "html", "txt", "PNG"} {
		if AllowedExt(ext) {
			t.Errorf("expected %q to be rejected", ext)
		}
	}
}

func TestFindUsesDeterministicOrdering(t *testing.T) {
	dir := t.TempDir()
	// Order is svg, png, ... so the svg must win even when a png exists too.
	writeFile(t, filepath.Join(dir, "tool.png"), []byte("png"))
	writeFile(t, filepath.Join(dir, "tool.svg"), []byte("svg"))

	name, ok := Find(dir, "tool")
	if !ok || name != "tool.svg" {
		t.Fatalf("Find = %q, %v; want tool.svg, true", name, ok)
	}
}

func TestFindMissingAndEmpty(t *testing.T) {
	dir := t.TempDir()
	if _, ok := Find(dir, "nope"); ok {
		t.Fatal("expected no favicon for unknown tool")
	}
	writeFile(t, filepath.Join(dir, "empty.png"), nil)
	if _, ok := Find(dir, "empty"); ok {
		t.Fatal("expected empty file to be ignored")
	}
}

func TestSaveWritesAndRejectsBadInput(t *testing.T) {
	dir := t.TempDir()
	if err := Save(dir, "tool", "png", []byte("data")); err != nil {
		t.Fatalf("Save: %v", err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "tool.png"))
	if err != nil || string(got) != "data" {
		t.Fatalf("saved bytes = %q, %v", got, err)
	}
	if err := Save(dir, "tool", "exe", []byte("x")); err == nil {
		t.Fatal("expected unsupported extension to be rejected")
	}
	if err := Save(dir, "tool", "png", nil); err == nil {
		t.Fatal("expected empty data to be rejected")
	}
}

func TestRunSkipsExisting(t *testing.T) {
	dir := t.TempDir()
	writeFile(t, filepath.Join(dir, "a.png"), []byte("cached"))
	called := false
	fetch := func(context.Context, string) (string, []byte, error) {
		called = true
		return "png", []byte("new"), nil
	}

	summary := Run(context.Background(), []Target{{ID: "a", Homepage: "https://a.example"}}, dir, false, fetch, nil)
	if called {
		t.Fatal("fetch should not be called for a cached favicon")
	}
	if summary.Skipped != 1 || summary.Fetched != 0 || summary.Failed != 0 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
}

func TestRunFetchesMissing(t *testing.T) {
	dir := t.TempDir()
	fetch := func(context.Context, string) (string, []byte, error) {
		return "png", []byte("new"), nil
	}
	summary := Run(context.Background(), []Target{{ID: "a", Homepage: "https://a.example"}}, dir, false, fetch, nil)
	if summary.Fetched != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if got, err := os.ReadFile(filepath.Join(dir, "a.png")); err != nil || string(got) != "new" {
		t.Fatalf("saved file = %q, %v", got, err)
	}
}

func TestRunRefreshFailureKeepsExisting(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "a.png")
	writeFile(t, path, []byte("cached"))
	fetch := func(context.Context, string) (string, []byte, error) {
		return "", nil, errors.New("network down")
	}
	summary := Run(context.Background(), []Target{{ID: "a", Homepage: "https://a.example"}}, dir, true, fetch, nil)
	if summary.Failed != 1 || summary.Fetched != 0 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if got, err := os.ReadFile(path); err != nil || string(got) != "cached" {
		t.Fatalf("existing favicon was damaged: %q, %v", got, err)
	}
}

func TestRunOneFailureDoesNotAbortOthers(t *testing.T) {
	dir := t.TempDir()
	fetch := func(_ context.Context, homepage string) (string, []byte, error) {
		if homepage == "https://bad.example" {
			return "", nil, errors.New("boom")
		}
		return "png", []byte("ok"), nil
	}
	targets := []Target{
		{ID: "bad", Homepage: "https://bad.example"},
		{ID: "good", Homepage: "https://good.example"},
	}
	summary := Run(context.Background(), targets, dir, false, fetch, nil)
	if summary.Fetched != 1 || summary.Failed != 1 {
		t.Fatalf("unexpected summary: %+v", summary)
	}
	if _, err := os.Stat(filepath.Join(dir, "good.png")); err != nil {
		t.Fatalf("good favicon missing: %v", err)
	}
	if len(summary.Failures) != 1 {
		t.Fatalf("expected one recorded failure, got %v", summary.Failures)
	}
}

func writeFile(t *testing.T, path string, data []byte) {
	t.Helper()
	if err := os.WriteFile(path, data, 0o644); err != nil {
		t.Fatalf("write %s: %v", path, err)
	}
}
