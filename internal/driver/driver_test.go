package driver

import (
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/docker/machine/commands/commandstest"
	"github.com/docker/machine/libmachine/drivers"
	"github.com/hetznercloud/hcloud-go/v2/hcloud"
)

var defaultFlags = map[string]interface{}{
	flagAPIToken: "foo",
}

func makeFlags(args map[string]interface{}) drivers.DriverOptions {
	combined := make(map[string]interface{}, len(defaultFlags)+len(args))
	for k, v := range defaultFlags {
		combined[k] = v
	}
	for k, v := range args {
		combined[k] = v
	}

	return &commandstest.FakeFlagger{Data: combined}
}

func TestUserData(t *testing.T) {
	const fileContents = "User data from file"
	const inlineContents = "User data"

	file := t.TempDir() + string(os.PathSeparator) + "userData"
	err := os.WriteFile(file, []byte(fileContents), 0644)
	if err != nil {
		t.Fatal(err)
	}

	// mutual exclusion data <=> data file
	d := NewDriver("test")
	err = d.setConfigFromFlagsImpl(makeFlags(map[string]interface{}{
		flagUserData:     inlineContents,
		flagUserDataFile: file,
	}))
	assertMutualExclusion(t, err, flagUserData, flagUserDataFile)

	// mutual exclusion data file <=> legacy flag
	d = NewDriver("test")
	err = d.setConfigFromFlagsImpl(&commandstest.FakeFlagger{
		Data: map[string]interface{}{
			legacyFlagUserDataFromFile: true,
			flagUserDataFile:           file,
		},
	})
	assertMutualExclusion(t, err, legacyFlagUserDataFromFile, flagUserDataFile)

	// inline user data
	d = NewDriver("test")
	err = d.setConfigFromFlagsImpl(makeFlags(map[string]interface{}{
		flagUserData: inlineContents,
	}))
	if err != nil {
		t.Fatalf("unexpected error, %v", err)
	}

	data, err := d.getUserData()
	if err != nil {
		t.Fatalf("unexpected error, %v", err)
	}
	if data != inlineContents {
		t.Error("content did not match (inline)")
	}

	// file user data
	d = NewDriver("test")
	err = d.setConfigFromFlagsImpl(makeFlags(map[string]interface{}{
		flagUserDataFile: file,
	}))
	if err != nil {
		t.Fatalf("unexpected error, %v", err)
	}

	data, err = d.getUserData()
	if err != nil {
		t.Fatalf("unexpected error, %v", err)
	}
	if data != fileContents {
		t.Error("content did not match (file)")
	}

	// legacy file user data
	d = NewDriver("test")
	err = d.setConfigFromFlagsImpl(makeFlags(map[string]interface{}{
		flagUserData:               file,
		legacyFlagUserDataFromFile: true,
	}))
	if err != nil {
		t.Fatalf("unexpected error, %v", err)
	}

	data, err = d.getUserData()
	if err != nil {
		t.Fatalf("unexpected error, %v", err)
	}
	if data != fileContents {
		t.Error("content did not match (legacy-file)")
	}
}

func TestDisablePublic(t *testing.T) {
	d := NewDriver("test")
	err := d.setConfigFromFlagsImpl(makeFlags(map[string]interface{}{
		flagDisablePublic: true,
	}))
	if err != nil {
		t.Fatalf("unexpected error, %v", err)
	}

	if !d.DisablePublic4 {
		t.Error("expected public ipv4 to be disabled")
	}
	if !d.DisablePublic6 {
		t.Error("expected public ipv6 to be disabled")
	}
	if !d.UsePrivateNetwork {
		t.Error("expected private network to be enabled")
	}
}

func TestDisablePublic46(t *testing.T) {
	d := NewDriver("test")
	err := d.setConfigFromFlagsImpl(makeFlags(map[string]interface{}{
		flagDisablePublic4: true,
	}))
	if err != nil {
		t.Fatalf("unexpected error, %v", err)
	}

	if !d.DisablePublic4 {
		t.Error("expected public ipv4 to be disabled")
	}
	if d.DisablePublic6 {
		t.Error("public ipv6 disabled unexpectedly")
	}
	if d.UsePrivateNetwork {
		t.Error("network enabled unexpectedly")
	}

	// 6
	d = NewDriver("test")
	err = d.setConfigFromFlagsImpl(makeFlags(map[string]interface{}{
		flagDisablePublic6: true,
	}))
	if err != nil {
		t.Fatalf("unexpected error, %v", err)
	}

	if d.DisablePublic4 {
		t.Error("public ipv4 disabled unexpectedly")
	}
	if !d.DisablePublic6 {
		t.Error("expected public ipv6 to be disabled")
	}
	if d.UsePrivateNetwork {
		t.Error("network enabled unexpectedly")
	}
}

func TestDisablePublic46Legacy(t *testing.T) {
	d := NewDriver("test")
	err := d.setConfigFromFlagsImpl(makeFlags(map[string]interface{}{
		legacyFlagDisablePublic4: true,
		// any truthy flag should take precedence
		flagDisablePublic4: false,
	}))
	if err != nil {
		t.Fatalf("unexpected error, %v", err)
	}

	if !d.DisablePublic4 {
		t.Error("expected public ipv4 to be disabled")
	}
	if d.DisablePublic6 {
		t.Error("public ipv6 disabled unexpectedly")
	}
	if d.UsePrivateNetwork {
		t.Error("network enabled unexpectedly")
	}

	// 6
	d = NewDriver("test")
	err = d.setConfigFromFlagsImpl(makeFlags(map[string]interface{}{
		legacyFlagDisablePublic6: true,
		// any truthy flag should take precedence
		flagDisablePublic6: false,
	}))
	if err != nil {
		t.Fatalf("unexpected error, %v", err)
	}

	if d.DisablePublic4 {
		t.Error("public ipv4 disabled unexpectedly")
	}
	if !d.DisablePublic6 {
		t.Error("expected public ipv6 to be disabled")
	}
	if d.UsePrivateNetwork {
		t.Error("network enabled unexpectedly")
	}
}

func TestImageFlagExclusions(t *testing.T) {
	// both id and name given
	d := NewDriver("test")
	err := d.setConfigFromFlagsImpl(makeFlags(map[string]interface{}{
		flagImageID: "42",
		flagImage:   "answer",
	}))
	assertMutualExclusion(t, err, flagImageID, flagImage)

	// both id and arch given
	d = NewDriver("test")
	err = d.setConfigFromFlagsImpl(makeFlags(map[string]interface{}{
		flagImageID:   "42",
		flagImageArch: string(hcloud.ArchitectureX86),
	}))
	assertMutualExclusion(t, err, flagImageID, flagImageArch)
}

func TestImageArch(t *testing.T) {
	// no explicit arch
	d := NewDriver("test")
	err := d.setConfigFromFlagsImpl(makeFlags(map[string]interface{}{
		flagImage: "answer",
	}))
	if err != nil {
		t.Fatalf("unexpected error, %v", err)
	}

	if d.ImageArch != emptyImageArchitecture {
		t.Errorf("expected empty architecture, but got %v", d.ImageArch)
	}

	// existing architectures
	testArchFlag(t, hcloud.ArchitectureARM)
	testArchFlag(t, hcloud.ArchitectureX86)

	// invalid
	d = NewDriver("test")
	err = d.setConfigFromFlagsImpl(makeFlags(map[string]interface{}{
		flagImage:     "answer",
		flagImageArch: "hal9000",
	}))
	if err == nil {
		t.Fatal("expected error, but invalid arch was accepted")
	}
}

func TestBogusId(t *testing.T) {
	d := NewDriver("test")
	err := d.setConfigFromFlagsImpl(makeFlags(map[string]interface{}{
		flagImageID: "answer",
	}))
	if err == nil {
		t.Fatal("expected error, but invalid arch was accepted")
	}
}

func TestLongId(t *testing.T) {
	var testId int64 = 79871865169581

	d := NewDriver("test")
	err := d.setConfigFromFlagsImpl(makeFlags(map[string]interface{}{
		flagImageID: strconv.FormatInt(testId, 10),
	}))
	if err != nil {
		t.Fatalf("unexpected error, %v", err)
	}

	if d.ImageID != testId {
		t.Errorf("expected %v id, but got %v", testId, d.ImageArch)
	}
}

func testArchFlag(t *testing.T, arch hcloud.Architecture) {
	d := NewDriver("test")
	err := d.setConfigFromFlagsImpl(makeFlags(map[string]interface{}{
		flagImage:     "answer",
		flagImageArch: string(arch),
	}))
	if err != nil {
		t.Fatalf("unexpected error, %v", err)
	}

	if d.ImageArch != arch {
		t.Errorf("expected %v architecture, but got %v", arch, d.ImageArch)
	}
}

func assertMutualExclusion(t *testing.T, err error, flag1, flag2 string) {
	if err == nil {
		t.Errorf("expected mutually exclusive flags to fail, but no error was thrown: %v %v", flag1, flag2)
		return
	}

	errstr := err.Error()
	if !(strings.Contains(errstr, flag1) && strings.Contains(errstr, flag2) && strings.Contains(errstr, "mutually exclusive")) {
		t.Errorf("expected mutually exclusive flags to fail, but message differs: %v %v %v", flag1, flag2, errstr)
	}
}

func TestAdditionalUserData(t *testing.T) {
	tests := []struct {
		name               string
		baseUserData       string
		additionalUserData string
		wantContains       []string
	}{
		{
			name:               "only additional user data",
			baseUserData:       "",
			additionalUserData: "#cloud-config\npackages:\n  - vim",
			wantContains:       []string{"packages:", "vim"},
		},
		{
			name:               "only base user data",
			baseUserData:       "#cloud-config\npackages:\n  - git",
			additionalUserData: "",
			wantContains:       []string{"packages:", "git"},
		},
		{
			name:               "merge packages lists",
			baseUserData:       "#cloud-config\npackages:\n  - git",
			additionalUserData: "#cloud-config\npackages:\n  - vim",
			wantContains:       []string{"packages:", "vim", "git"},
		},
		{
			name:               "merge runcmd lists",
			baseUserData:       "#cloud-config\nruncmd:\n  - echo base",
			additionalUserData: "#cloud-config\nruncmd:\n  - echo additional",
			wantContains:       []string{"runcmd:", "echo base", "echo additional"},
		},
		{
			name:               "different keys",
			baseUserData:       "#cloud-config\npackages:\n  - git",
			additionalUserData: "#cloud-config\nruncmd:\n  - echo hello",
			wantContains:       []string{"packages:", "git", "runcmd:", "echo hello"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			d := NewDriver("test")
			d.userData = tt.baseUserData
			d.additionalUserData = tt.additionalUserData

			result, err := d.getUserData()
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			for _, want := range tt.wantContains {
				if !strings.Contains(result, want) {
					t.Errorf("result should contain %q, got:\n%s", want, result)
				}
			}
		})
	}
}

func TestMergeUserDataPrepends(t *testing.T) {
	base := "#cloud-config\nruncmd:\n  - echo base"
	additional := "#cloud-config\nruncmd:\n  - echo additional"

	result, err := mergeUserData(base, additional)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	additionalIdx := strings.Index(result, "echo additional")
	baseIdx := strings.Index(result, "echo base")

	if additionalIdx == -1 || baseIdx == -1 {
		t.Fatalf("expected both commands in result, got:\n%s", result)
	}

	if additionalIdx > baseIdx {
		t.Errorf("additional data should be prepended, but base appears first in result:\n%s", result)
	}
}

func TestAdditionalUserDataFlag(t *testing.T) {
	d := NewDriver("test")
	err := d.setConfigFromFlagsImpl(makeFlags(map[string]interface{}{
		flagUserData:           "#cloud-config\npackages:\n  - git",
		flagAdditionalUserData: "#cloud-config\npackages:\n  - vim",
	}))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	data, err := d.getUserData()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !strings.Contains(data, "git") {
		t.Error("result should contain git")
	}
	if !strings.Contains(data, "vim") {
		t.Error("result should contain vim")
	}
}
