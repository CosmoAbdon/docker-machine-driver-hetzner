package driver

import (
	"testing"

	"github.com/docker/machine/commands/commandstest"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

func TestIsDefaultImageName(t *testing.T) {
	tests := []struct {
		name      string
		imageName string
		expected  bool
	}{
		{"current default image", defaultImage, true},
		{"legacy ubuntu-18.04", "ubuntu-18.04", true},
		{"legacy ubuntu-16.04", "ubuntu-16.04", true},
		{"legacy debian-9", "debian-9", true},
		{"custom image", "my-custom-image", false},
		{"empty string", "", false},
		{"similar but not default", "ubuntu-24.04-custom", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDefaultImageName(tt.imageName)
			if result != tt.expected {
				t.Errorf("isDefaultImageName(%q) = %v, want %v", tt.imageName, result, tt.expected)
			}
		})
	}
}

func TestSetImageArch(t *testing.T) {
	tests := []struct {
		name        string
		arch        string
		expected    hcloud.Architecture
		expectError bool
	}{
		{"empty string", "", emptyImageArchitecture, false},
		{"ARM architecture", string(hcloud.ArchitectureARM), hcloud.ArchitectureARM, false},
		{"x86 architecture", string(hcloud.ArchitectureX86), hcloud.ArchitectureX86, false},
		{"invalid architecture", "invalid-arch", "", true},
		{"random string", "foobar", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDriver("test")
			err := d.setImageArch(tt.arch)

			if tt.expectError {
				if err == nil {
					t.Errorf("setImageArch(%q) expected error, got nil", tt.arch)
				}
			} else {
				if err != nil {
					t.Errorf("setImageArch(%q) unexpected error: %v", tt.arch, err)
				}
				if d.ImageArch != tt.expected {
					t.Errorf("setImageArch(%q) = %v, want %v", tt.arch, d.ImageArch, tt.expected)
				}
			}
		})
	}
}

func TestVerifyImageFlags(t *testing.T) {
	tests := []struct {
		name        string
		imageID     int64
		image       string
		imageArch   hcloud.Architecture
		expectError bool
		errorFlags  []string
	}{
		{
			name:        "no image specified, defaults to defaultImage",
			imageID:     0,
			image:       "",
			imageArch:   "",
			expectError: false,
		},
		{
			name:        "only image name",
			imageID:     0,
			image:       "custom-image",
			imageArch:   "",
			expectError: false,
		},
		{
			name:        "only image ID",
			imageID:     123,
			image:       "",
			imageArch:   "",
			expectError: false,
		},
		{
			name:        "image ID with default image name (legacy support)",
			imageID:     123,
			image:       defaultImage,
			imageArch:   "",
			expectError: false,
		},
		{
			name:        "image ID with legacy ubuntu-18.04 (legacy support)",
			imageID:     123,
			image:       "ubuntu-18.04",
			imageArch:   "",
			expectError: false,
		},
		{
			name:        "both image ID and custom image name",
			imageID:     123,
			image:       "custom-image",
			imageArch:   "",
			expectError: true,
			errorFlags:  []string{flagImage, flagImageID},
		},
		{
			name:        "image ID with arch",
			imageID:     123,
			image:       "",
			imageArch:   hcloud.ArchitectureX86,
			expectError: true,
			errorFlags:  []string{flagImageArch, flagImageID},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDriver("test")
			d.ImageID = tt.imageID
			d.Image = tt.image
			d.ImageArch = tt.imageArch

			err := d.verifyImageFlags()

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				} else {
					for _, flag := range tt.errorFlags {
						assertMutualExclusion(t, err, tt.errorFlags[0], tt.errorFlags[1])
						_ = flag
					}
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if tt.imageID == 0 && tt.image == "" && d.Image != defaultImage {
					t.Errorf("expected default image %q, got %q", defaultImage, d.Image)
				}
			}
		})
	}
}

func TestVerifyNetworkFlags(t *testing.T) {
	tests := []struct {
		name              string
		disablePublic4    bool
		disablePublic6    bool
		usePrivateNetwork bool
		primaryIPv4       string
		primaryIPv6       string
		expectError       bool
		errorFlags        []string
	}{
		{
			name:              "all public enabled",
			disablePublic4:    false,
			disablePublic6:    false,
			usePrivateNetwork: false,
			expectError:       false,
		},
		{
			name:              "public disabled with private network",
			disablePublic4:    true,
			disablePublic6:    true,
			usePrivateNetwork: true,
			expectError:       false,
		},
		{
			name:              "only ipv4 disabled",
			disablePublic4:    true,
			disablePublic6:    false,
			usePrivateNetwork: false,
			expectError:       false,
		},
		{
			name:              "only ipv6 disabled",
			disablePublic4:    false,
			disablePublic6:    true,
			usePrivateNetwork: false,
			expectError:       false,
		},
		{
			name:              "all public disabled without private network",
			disablePublic4:    true,
			disablePublic6:    true,
			usePrivateNetwork: false,
			expectError:       true,
			errorFlags:        []string{flagUsePrivateNetwork, flagDisablePublic},
		},
		{
			name:              "ipv4 disabled with primary ipv4",
			disablePublic4:    true,
			disablePublic6:    false,
			usePrivateNetwork: false,
			primaryIPv4:       "my-primary-ipv4",
			expectError:       true,
			errorFlags:        []string{flagPrimary4, flagDisablePublic4},
		},
		{
			name:              "ipv6 disabled with primary ipv6",
			disablePublic4:    false,
			disablePublic6:    true,
			usePrivateNetwork: false,
			primaryIPv6:       "my-primary-ipv6",
			expectError:       true,
			errorFlags:        []string{flagPrimary6, flagDisablePublic6},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDriver("test")
			d.DisablePublic4 = tt.disablePublic4
			d.DisablePublic6 = tt.disablePublic6
			d.UsePrivateNetwork = tt.usePrivateNetwork
			d.PrimaryIPv4 = tt.primaryIPv4
			d.PrimaryIPv6 = tt.primaryIPv6

			err := d.verifyNetworkFlags()

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				} else if len(tt.errorFlags) >= 2 {
					assertMutualExclusion(t, err, tt.errorFlags[0], tt.errorFlags[1])
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
			}
		})
	}
}

func TestSetLabelsFromFlags(t *testing.T) {
	tests := []struct {
		name           string
		serverLabels   []string
		keyLabels      []string
		expectError    bool
		expectedServer map[string]string
		expectedKey    map[string]string
	}{
		{
			name:           "no labels",
			serverLabels:   []string{},
			keyLabels:      []string{},
			expectError:    false,
			expectedServer: map[string]string{},
			expectedKey:    map[string]string{},
		},
		{
			name:           "single server label",
			serverLabels:   []string{"env=production"},
			keyLabels:      []string{},
			expectError:    false,
			expectedServer: map[string]string{"env": "production"},
			expectedKey:    map[string]string{},
		},
		{
			name:           "multiple server labels",
			serverLabels:   []string{"env=production", "team=backend", "version=1.0"},
			keyLabels:      []string{},
			expectError:    false,
			expectedServer: map[string]string{"env": "production", "team": "backend", "version": "1.0"},
			expectedKey:    map[string]string{},
		},
		{
			name:           "single key label",
			serverLabels:   []string{},
			keyLabels:      []string{"managed-by=docker-machine"},
			expectError:    false,
			expectedServer: map[string]string{},
			expectedKey:    map[string]string{"managed-by": "docker-machine"},
		},
		{
			name:           "both server and key labels",
			serverLabels:   []string{"env=dev"},
			keyLabels:      []string{"owner=test"},
			expectError:    false,
			expectedServer: map[string]string{"env": "dev"},
			expectedKey:    map[string]string{"owner": "test"},
		},
		{
			name:           "label with value containing equals sign",
			serverLabels:   []string{"config=key=value"},
			keyLabels:      []string{},
			expectError:    false,
			expectedServer: map[string]string{"config": "key=value"},
			expectedKey:    map[string]string{},
		},
		{
			name:         "invalid server label format",
			serverLabels: []string{"invalid-label"},
			keyLabels:    []string{},
			expectError:  true,
		},
		{
			name:         "invalid key label format",
			serverLabels: []string{},
			keyLabels:    []string{"no-equals-sign"},
			expectError:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDriver("test")
			opts := &commandstest.FakeFlagger{
				Data: map[string]interface{}{
					flagServerLabel: tt.serverLabels,
					flagKeyLabel:    tt.keyLabels,
				},
			}

			err := d.setLabelsFromFlags(opts)

			if tt.expectError {
				if err == nil {
					t.Error("expected error, got nil")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if len(d.ServerLabels) != len(tt.expectedServer) {
					t.Errorf("server labels count mismatch: got %d, want %d", len(d.ServerLabels), len(tt.expectedServer))
				}
				for k, v := range tt.expectedServer {
					if d.ServerLabels[k] != v {
						t.Errorf("server label %q = %q, want %q", k, d.ServerLabels[k], v)
					}
				}

				if len(d.keyLabels) != len(tt.expectedKey) {
					t.Errorf("key labels count mismatch: got %d, want %d", len(d.keyLabels), len(tt.expectedKey))
				}
				for k, v := range tt.expectedKey {
					if d.keyLabels[k] != v {
						t.Errorf("key label %q = %q, want %q", k, d.keyLabels[k], v)
					}
				}
			}
		})
	}
}

func TestDeprecatedBooleanFlag(t *testing.T) {
	tests := []struct {
		name           string
		flagValue      bool
		deprecatedFlag bool
		expected       bool
		expectDfr      bool
	}{
		{
			name:           "neither flag set",
			flagValue:      false,
			deprecatedFlag: false,
			expected:       false,
			expectDfr:      false,
		},
		{
			name:           "only current flag set",
			flagValue:      true,
			deprecatedFlag: false,
			expected:       true,
			expectDfr:      false,
		},
		{
			name:           "only deprecated flag set",
			flagValue:      false,
			deprecatedFlag: true,
			expected:       true,
			expectDfr:      true,
		},
		{
			name:           "both flags set",
			flagValue:      true,
			deprecatedFlag: true,
			expected:       true,
			expectDfr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDriver("test")
			opts := &commandstest.FakeFlagger{
				Data: map[string]interface{}{
					"current-flag":    tt.flagValue,
					"deprecated-flag": tt.deprecatedFlag,
				},
			}

			result := d.deprecatedBooleanFlag(opts, "current-flag", "deprecated-flag")

			if result != tt.expected {
				t.Errorf("deprecatedBooleanFlag() = %v, want %v", result, tt.expected)
			}
			if d.usesDfr != tt.expectDfr {
				t.Errorf("usesDfr = %v, want %v", d.usesDfr, tt.expectDfr)
			}
		})
	}
}
