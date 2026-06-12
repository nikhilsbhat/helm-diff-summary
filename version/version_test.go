package version

import "testing"

func TestGetBuildInfoDefaultsNonProductionEnvironment(t *testing.T) {
	original := BuildInfo{
		Version:     Version,
		Revision:    Revision,
		Environment: Env,
		BuildDate:   BuildDate,
		GoVersion:   GoVersion,
		Platform:    Platform,
	}
	t.Cleanup(func() {
		Version = original.Version
		Revision = original.Revision
		Env = original.Environment
		BuildDate = original.BuildDate
		GoVersion = original.GoVersion
		Platform = original.Platform
	})

	Version = "1.2.3"
	Revision = "abc123"
	Env = "local"
	BuildDate = "today"
	GoVersion = "go1.25"
	Platform = "linux/amd64"

	info := GetBuildInfo()
	if info.Environment != "alfa" {
		t.Fatalf("expected alfa environment, got %q", info.Environment)
	}

	if info.Version != Version || info.Revision != Revision || info.BuildDate != BuildDate ||
		info.GoVersion != GoVersion || info.Platform != Platform {
		t.Fatalf("unexpected build info: %#v", info)
	}
}

func TestGetBuildInfoKeepsProductionEnvironment(t *testing.T) {
	originalEnv := Env
	t.Cleanup(func() { Env = originalEnv })

	Env = "production"

	if info := GetBuildInfo(); info.Environment != "production" {
		t.Fatalf("expected production environment, got %q", info.Environment)
	}
}
