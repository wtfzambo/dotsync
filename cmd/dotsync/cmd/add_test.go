package cmd

import (
	"strings"
	"testing"
)

func TestValidateEntryName(t *testing.T) {
	tests := []struct {
		name        string
		entryName   string
		wantErr     bool
		errContains string
	}{
		{
			name:      "valid name",
			entryName: "testapp",
			wantErr:   false,
		},
		{
			name:        "empty name",
			entryName:   "",
			wantErr:     true,
			errContains: "empty",
		},
		{
			name:        "path separator forward slash",
			entryName:   "foo/bar",
			wantErr:     true,
			errContains: "separator",
		},
		{
			name:        "path separator backslash",
			entryName:   "foo\\bar",
			wantErr:     true,
			errContains: "separator",
		},
		{
			name:        "current directory",
			entryName:   ".",
			wantErr:     true,
			errContains: "cannot be",
		},
		{
			name:        "parent directory",
			entryName:   "..",
			wantErr:     true,
			errContains: "cannot be",
		},
		{
			name:        "multiple separators",
			entryName:   "foo/bar/baz",
			wantErr:     true,
			errContains: "separator",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateEntryName(tt.entryName)

			if tt.wantErr {
				if err == nil {
					t.Errorf("expected error for name %q, got nil", tt.entryName)
					return
				}
				if tt.errContains != "" && !strings.Contains(strings.ToLower(err.Error()), tt.errContains) {
					t.Errorf("error should contain %q, got %v", tt.errContains, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error for valid name %q: %v", tt.entryName, err)
				}
			}
		})
	}
}
