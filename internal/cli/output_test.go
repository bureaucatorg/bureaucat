package cli

import "testing"

func TestParseTaskID(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		wantProject string
		wantNum     string
		wantErr     bool
	}{
		{name: "simple", input: "PROJ-42", wantProject: "PROJ", wantNum: "42"},
		{name: "hyphenated project", input: "MY-PROJ-7", wantProject: "MY-PROJ", wantNum: "7"},
		{name: "missing dash", input: "PROJ42", wantErr: true},
		{name: "missing number", input: "PROJ-", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotProject, gotNum, err := parseTaskID(tt.input)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseTaskID(%q) error = nil, want error", tt.input)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseTaskID(%q) error = %v", tt.input, err)
			}
			if gotProject != tt.wantProject || gotNum != tt.wantNum {
				t.Fatalf("parseTaskID(%q) = (%q, %q), want (%q, %q)", tt.input, gotProject, gotNum, tt.wantProject, tt.wantNum)
			}
		})
	}
}
