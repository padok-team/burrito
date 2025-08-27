package terragrunt

import (
	"reflect"
	"testing"
)

func TestTerragrunt_getDefaultOptions(t *testing.T) {
	tests := []struct {
		name          string
		version       string
		command       string
		workingDir    string
		childExecPath string
		expectedFlags []string
		expectError   bool
	}{
		{
			name:          "Legacy flags for version 0.66.9",
			version:       "0.66.9",
			command:       "plan",
			workingDir:    "/path/to/working/dir",
			childExecPath: "/path/to/terraform",
			expectedFlags: []string{
				"plan",
				"--terragrunt-tfpath",
				"/path/to/terraform",
				"--terragrunt-working-dir",
				"/path/to/working/dir",
				"-no-color",
			},
			expectError: false,
		},
		{
			name:          "Legacy flags for version 0.72.9",
			version:       "0.72.9",
			command:       "apply",
			workingDir:    "/path/to/working/dir",
			childExecPath: "/path/to/terraform",
			expectedFlags: []string{
				"apply",
				"--terragrunt-tfpath",
				"/path/to/terraform",
				"--terragrunt-working-dir",
				"/path/to/working/dir",
				"-no-color",
			},
			expectError: false,
		},
		{
			name:          "New flags for version 0.73.0",
			version:       "0.73.0",
			command:       "plan",
			workingDir:    "/path/to/working/dir",
			childExecPath: "/path/to/terraform",
			expectedFlags: []string{
				"plan",
				"--tf-path",
				"/path/to/terraform",
				"--working-dir",
				"/path/to/working/dir",
				"-no-color",
			},
			expectError: false,
		},
		{
			name:          "New flags for version 0.73.1",
			version:       "0.73.1",
			command:       "init",
			workingDir:    "/path/to/working/dir",
			childExecPath: "/path/to/terraform",
			expectedFlags: []string{
				"init",
				"--tf-path",
				"/path/to/terraform",
				"--working-dir",
				"/path/to/working/dir",
				"-no-color",
			},
			expectError: false,
		},
		{
			name:          "New flags for version 1.0.0",
			version:       "1.0.0",
			command:       "show",
			workingDir:    "/path/to/working/dir",
			childExecPath: "/path/to/terraform",
			expectedFlags: []string{
				"show",
				"--tf-path",
				"/path/to/terraform",
				"--working-dir",
				"/path/to/working/dir",
				"-no-color",
			},
			expectError: false,
		},
		{
			name:          "Invalid version fallback to legacy flags",
			version:       "invalid-version",
			command:       "plan",
			workingDir:    "/path/to/working/dir",
			childExecPath: "/path/to/terraform",
			expectedFlags: []string{
				"plan",
				"--terragrunt-tfpath",
				"/path/to/terraform",
				"--terragrunt-working-dir",
				"/path/to/working/dir",
				"-no-color",
			},
			expectError: false,
		},
		{
			name:          "Empty version fallback to legacy flags",
			version:       "",
			command:       "apply",
			workingDir:    "/path/to/working/dir",
			childExecPath: "/path/to/terraform",
			expectedFlags: []string{
				"apply",
				"--terragrunt-tfpath",
				"/path/to/terraform",
				"--terragrunt-working-dir",
				"/path/to/working/dir",
				"-no-color",
			},
			expectError: false,
		},
		{
			name:          "Pre-release version 0.73.0-rc1 uses legacy flags",
			version:       "0.73.0-rc1",
			command:       "plan",
			workingDir:    "/path/to/working/dir",
			childExecPath: "/path/to/terraform",
			expectedFlags: []string{
				"plan",
				"--terragrunt-tfpath",
				"/path/to/terraform",
				"--terragrunt-working-dir",
				"/path/to/working/dir",
				"-no-color",
			},
			expectError: false,
		},
		{
			name:          "Pre-release version 0.74.0-beta1 uses new flags",
			version:       "0.74.0-beta1",
			command:       "plan",
			workingDir:    "/path/to/working/dir",
			childExecPath: "/path/to/terraform",
			expectedFlags: []string{
				"plan",
				"--tf-path",
				"/path/to/terraform",
				"--working-dir",
				"/path/to/working/dir",
				"-no-color",
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tg := &Terragrunt{
				ExecPath:      "/path/to/terragrunt",
				WorkingDir:    tt.workingDir,
				ChildExecPath: tt.childExecPath,
				Version:       tt.version,
			}

			options, err := tg.getDefaultOptions(tt.command)

			if tt.expectError && err == nil {
				t.Errorf("Expected an error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("Expected no error but got: %v", err)
			}

			if !reflect.DeepEqual(options, tt.expectedFlags) {
				t.Errorf("Expected flags %v, but got %v", tt.expectedFlags, options)
			}
		})
	}
}

func TestTerragrunt_TenvName(t *testing.T) {
	tg := &Terragrunt{}
	expected := "terragrunt"
	result := tg.TenvName()

	if result != expected {
		t.Errorf("Expected TenvName to return %q, but got %q", expected, result)
	}
}

func TestTerragrunt_GetExecPath(t *testing.T) {
	expectedPath := "/path/to/terragrunt"
	tg := &Terragrunt{
		ExecPath: expectedPath,
	}

	result := tg.GetExecPath()

	if result != expectedPath {
		t.Errorf("Expected GetExecPath to return %q, but got %q", expectedPath, result)
	}
}
