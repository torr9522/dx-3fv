package contract_test

import (
	"encoding/json"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestNativeAgentReleaseBootstrapsEmptyRuntimeConfig(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve contract test path")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "../../.."))

	assetPath := filepath.Join(repoRoot, "release/native/agent-gost.json")
	asset, err := os.ReadFile(assetPath)
	if err != nil {
		t.Fatalf("read runtime bootstrap asset: %v", err)
	}
	var runtimeConfig map[string]any
	if err := json.Unmarshal(asset, &runtimeConfig); err != nil {
		t.Fatalf("runtime bootstrap asset must be valid JSON: %v", err)
	}
	if len(runtimeConfig) != 0 {
		t.Fatalf("fresh runtime bootstrap must not define services: %#v", runtimeConfig)
	}

	buildScript := readReleaseContractFile(t, repoRoot, "scripts/build-native-release.sh")
	if !strings.Contains(buildScript, "cp \"$repo_root/release/native/agent-gost.json\" \"$stage/deploy/\"") {
		t.Fatal("native release does not package the runtime bootstrap asset")
	}

	installScript := readReleaseContractFile(t, repoRoot, "release/native/install-agent-native.sh")
	installLine := "install -m 0600 \"$release_dir/deploy/agent-gost.json\" \"$target/gost.json\""
	installAt := strings.Index(installScript, installLine)
	startAt := strings.Index(installScript, "systemctl enable --now dx-3fv-agent.service")
	caCheckAt := strings.Index(installScript, "if [[ ! -s \"$ca_bundle\" ]]")
	if caCheckAt < 0 {
		t.Fatal("native Agent installer does not fail closed when the CA bundle is missing")
	}
	if installAt < 0 {
		t.Fatal("native Agent installer does not install gost.json with mode 0600")
	}
	if startAt < 0 || caCheckAt > startAt || installAt > startAt {
		t.Fatal("native Agent installer must validate CA and create gost.json before starting the service")
	}
	for _, contract := range []string{
		`active_state=$(systemctl is-active dx-3fv-agent.service 2>/dev/null || true)`,
		`sub_state=$(systemctl show -p SubState --value dx-3fv-agent.service 2>/dev/null || true)`,
		`test "$ready_checks" -ge 3`,
	} {
		if !strings.Contains(installScript, contract) {
			t.Fatalf("native Agent installer is missing stable readiness contract %q", contract)
		}
	}
}

func TestNativeReleaseTemplatesDoNotContainDeploymentIdentity(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve contract test path")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "../../.."))

	panelEnv := readReleaseContractFile(t, repoRoot, "release/native/panel.env.example")
	if !strings.Contains(panelEnv, "FLVX_WEBAUTHN_ORIGIN=\n") {
		t.Fatal("passkey origin must be opt-in in the release environment template")
	}

	caddy := readReleaseContractFile(t, repoRoot, "release/native/dx-3fv.caddy")
	if !strings.Contains(caddy, "panel.example.com {") {
		t.Fatal("Caddy template must use the documented example hostname")
	}

	for name, contents := range map[string]string{
		"panel.env.example": panelEnv,
		"dx-3-fv.caddy":     caddy,
	} {
		for _, forbidden := range []string{"ixcs.88988588.xyz", "185.126.158.33"} {
			if strings.Contains(contents, forbidden) {
				t.Fatalf("%s contains deployment-specific value %q", name, forbidden)
			}
		}
	}
}

func TestNativeReleasePackagesDocumentationAndPortableMetadata(t *testing.T) {
	_, currentFile, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve contract test path")
	}
	repoRoot := filepath.Clean(filepath.Join(filepath.Dir(currentFile), "../../.."))
	buildScript := readReleaseContractFile(t, repoRoot, "scripts/build-native-release.sh")

	for _, contract := range []string{
		`"$repo_root/README.md"`,
		`"$repo_root/THIRD_PARTY_NOTICES.md"`,
		`"$repo_root/LICENSES/."`,
		`frontend-production-licenses.json`,
		`$release_name-source.tar.gz`,
	} {
		if !strings.Contains(buildScript, contract) {
			t.Fatalf("native release is missing publication contract %q", contract)
		}
	}

	sanitizer := readReleaseContractFile(t, repoRoot, "scripts/sanitize-node-metadata.mjs")
	for _, key := range []string{`key !== "path"`, `key !== "paths"`} {
		if !strings.Contains(sanitizer, key) {
			t.Fatalf("Node metadata sanitizer is missing %q", key)
		}
	}

	goSanitizerPath := filepath.Join(repoRoot, "scripts/sanitize-go-module-metadata.mjs")
	inputPath := filepath.Join(t.TempDir(), "modules.ndjson")
	outputPath := filepath.Join(t.TempDir(), "modules.json")
	moduleInput := strings.Join([]string{
		`{
  "Path": "example.invalid/main",
  "Main": true,
  "Dir": "/root/source",
  "GoMod": "/root/source/go.mod",
  "GoModSum": "h1:main"
}`,
		`{
  "Path": "example.invalid/dependency",
  "Version": "v1.2.3",
  "Sum": "h1:module",
  "GoModSum": "h1:gomod",
  "Dir": "/root/go/pkg/mod/dependency",
  "Replace": {
    "Path": "example.invalid/fork",
    "Version": "v1.2.4",
    "Dir": "/root/fork",
    "GoMod": "/root/fork/go.mod"
  }
}`,
	}, "\n") + "\n"
	if err := os.WriteFile(inputPath, []byte(moduleInput), 0o600); err != nil {
		t.Fatalf("write Go module sanitizer fixture: %v", err)
	}
	command := exec.Command("node", goSanitizerPath, inputPath, outputPath)
	if output, err := command.CombinedOutput(); err != nil {
		t.Fatalf("run Go module metadata sanitizer: %v\n%s", err, output)
	}
	sanitized, err := os.ReadFile(outputPath)
	if err != nil {
		t.Fatalf("read sanitized Go module metadata: %v", err)
	}
	var modules []map[string]any
	if err := json.Unmarshal(sanitized, &modules); err != nil {
		t.Fatalf("sanitized Go module metadata must be a JSON array: %v", err)
	}
	if len(modules) != 2 {
		t.Fatalf("sanitized Go module metadata has %d modules, want 2", len(modules))
	}
	if modules[1]["Path"] != "example.invalid/dependency" ||
		modules[1]["Version"] != "v1.2.3" || modules[1]["Sum"] != "h1:module" {
		t.Fatalf("sanitizer removed module identity: %#v", modules[1])
	}
	replacement, ok := modules[1]["Replace"].(map[string]any)
	if !ok || replacement["Path"] != "example.invalid/fork" || replacement["Version"] != "v1.2.4" {
		t.Fatalf("sanitizer removed replacement identity: %#v", modules[1]["Replace"])
	}
	for _, forbidden := range []string{`"Dir"`, `"GoMod"`, `"GoModSum"`, "/root/"} {
		if strings.Contains(string(sanitized), forbidden) {
			t.Fatalf("sanitized Go module metadata contains forbidden value %q", forbidden)
		}
	}
	for _, contract := range []string{
		`go list -m -json all`,
		`sanitize-go-module-metadata.mjs`,
		`go-backend-modules.json`,
		`go-agent-modules.json`,
	} {
		if !strings.Contains(buildScript, contract) {
			t.Fatalf("native release is missing portable Go SBOM contract %q", contract)
		}
	}

	actualModulesPath := filepath.Join(t.TempDir(), "actual-modules.json")
	actualCommand := exec.Command("bash", "-o", "pipefail", "-c",
		`GOTOOLCHAIN=go1.26.11 go list -m -json all >"$1" && node "$2" "$1" "$3"`,
		"release-contract", filepath.Join(t.TempDir(), "actual-modules.stream"),
		goSanitizerPath, actualModulesPath)
	actualCommand.Dir = filepath.Join(repoRoot, "go-backend")
	if output, err := actualCommand.CombinedOutput(); err != nil {
		t.Fatalf("sanitize actual Go module stream: %v\n%s", err, output)
	}
	actualSanitized, err := os.ReadFile(actualModulesPath)
	if err != nil {
		t.Fatalf("read actual sanitized Go module metadata: %v", err)
	}
	var actualModules []map[string]any
	if err := json.Unmarshal(actualSanitized, &actualModules); err != nil {
		t.Fatalf("actual sanitized Go module metadata must be a JSON array: %v", err)
	}
	if len(actualModules) == 0 || actualModules[0]["Path"] != "go-backend" {
		t.Fatalf("actual sanitized Go module metadata is incomplete: %#v", actualModules)
	}
	for _, forbidden := range []string{`"Dir"`, `"GoMod"`, `"GoModSum"`, "/root/"} {
		if strings.Contains(string(actualSanitized), forbidden) {
			t.Fatalf("actual sanitized Go module metadata contains forbidden value %q", forbidden)
		}
	}
}

func readReleaseContractFile(t *testing.T, repoRoot, name string) string {
	t.Helper()
	contents, err := os.ReadFile(filepath.Join(repoRoot, name))
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	return string(contents)
}
