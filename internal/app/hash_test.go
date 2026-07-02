package app

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestLoadHashes(t *testing.T) {
	tests := []struct {
		name    string
		content []byte
		create  bool
		want    map[string]bool
	}{
		{
			name:    "normal json",
			content: []byte(`{"abc":true,"def":false}`),
			create:  true,
			want:    map[string]bool{"abc": true, "def": false},
		},
		{
			name:    "json with utf8 bom",
			content: append([]byte{0xEF, 0xBB, 0xBF}, []byte(`{"abc":true}`)...),
			create:  true,
			want:    map[string]bool{"abc": true},
		},
		{
			name:    "empty file",
			content: []byte{0x20, 0x20, 0x0A, 0x09, 0x20, 0x20},
			create:  true,
			want:    map[string]bool{},
		},
		{
			name:   "missing file",
			create: false,
			want:   map[string]bool{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			oldPath := hashesPath
			oldHashes := hashes
			t.Cleanup(func() {
				hashesPath = oldPath
				hashes = oldHashes
			})

			dir := t.TempDir()
			hashesPath = filepath.Join(dir, "import_hashes.json")
			hashes = map[string]bool{"stale": true}
			if tt.create {
				if err := os.WriteFile(hashesPath, tt.content, 0644); err != nil {
					t.Fatalf("write test hash file: %v", err)
				}
			}

			if err := loadHashes(); err != nil {
				t.Fatalf("loadHashes() error = %v", err)
			}
			if len(hashes) != len(tt.want) {
				t.Fatalf("len(hashes) = %d, want %d; hashes=%v", len(hashes), len(tt.want), hashes)
			}
			for key, want := range tt.want {
				if got := hashes[key]; got != want {
					t.Fatalf("hashes[%q] = %v, want %v", key, got, want)
				}
			}
		})
	}
}

func TestLoadHashesInvalidJSONIncludesPath(t *testing.T) {
	oldPath := hashesPath
	t.Cleanup(func() { hashesPath = oldPath })

	dir := t.TempDir()
	hashesPath = filepath.Join(dir, "import_hashes.json")
	if err := os.WriteFile(hashesPath, []byte("not-json"), 0644); err != nil {
		t.Fatalf("write test hash file: %v", err)
	}

	err := loadHashes()
	if err == nil {
		t.Fatal("loadHashes() error = nil, want parse error")
	}
	if !strings.Contains(err.Error(), hashesPath) {
		t.Fatalf("error %q does not include path %q", err.Error(), hashesPath)
	}
}
