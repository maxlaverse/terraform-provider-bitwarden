//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/magefile/mage/mg"
)

const (
	debugBwsVersion = "2.1.0"
	debugBwsRoot    = "tools/bws-debug"
)

// accTestEnv returns env vars common to all acceptance tests: avoid checkpoint (use local
// binary or fixed version), OpenTofu provider registry, and TF_ACC/CHECKPOINT_DISABLE.
func accTestEnv() []string {
	base := []string{"TF_ACC=1", "CHECKPOINT_DISABLE=1", "TF_ACC_PROVIDER_NAMESPACE=hashicorp", "TF_ACC_PROVIDER_HOST=registry.opentofu.org"}
	if path := os.Getenv("TF_ACC_TERRAFORM_PATH"); path != "" {
		return append([]string{"TF_ACC_TERRAFORM_PATH=" + path}, base...)
	}
	for _, name := range []string{"tofu", "terraform"} {
		if path, err := exec.LookPath(name); err == nil {
			return append([]string{"TF_ACC_TERRAFORM_PATH=" + path}, base...)
		}
	}
	return append([]string{"TF_ACC_TERRAFORM_VERSION=1.11.4"}, base...)
}

// overlayEnv copies os.Environ() and applies overrides. Duplicate keys are
// removed first so the override wins: Go's env lookup uses the first value,
// so appending TEST_BACKEND=official after a shell TEST_BACKEND=vaultwarden
// would otherwise keep the mock backend.
//
// Integration targets also pass -count=1: go test's cache ignores TEST_BACKEND,
// so an official run could replay a cached mock-server result.
func overlayEnv(overrides ...string) []string {
	keys := make(map[string]struct{}, len(overrides))
	for _, o := range overrides {
		k, _, _ := strings.Cut(o, "=")
		keys[k] = struct{}{}
	}

	env := os.Environ()
	out := make([]string, 0, len(env)+len(overrides))
	for _, e := range env {
		k, _, _ := strings.Cut(e, "=")
		if _, found := keys[k]; found {
			continue
		}
		out = append(out, e)
	}
	return append(out, overrides...)
}

func goTestEnv(overrides ...string) []string {
	return overlayEnv(append(accTestEnv(), overrides...)...)
}

func debugBwsBin() string {
	return filepath.Join(debugBwsRoot, "bin", "bws")
}

func debugBwsTargetDir() string {
	return filepath.Join(debugBwsRoot, "target")
}

func debugBwsInstalledVersion() string {
	out, err := exec.Command(debugBwsBin(), "--version").Output()
	if err != nil {
		return ""
	}
	fields := strings.Fields(string(out))
	if len(fields) == 0 {
		return ""
	}
	return fields[len(fields)-1]
}

func requireMiseCargo() error {
	if _, err := exec.LookPath("mise"); err != nil {
		return fmt.Errorf("mise is required to build debug bws %s for Secrets Manager CLI tests (install mise and run `mise install`): %w", debugBwsVersion, err)
	}
	cmd := exec.Command("mise", "exec", "--", "cargo", "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("rust/cargo from mise is required to build debug bws %s; run `mise install` to install the toolchain pinned in mise.toml: %w\n%s", debugBwsVersion, err, out)
	}
	return nil
}

func goTestEnvWithDebugBws(overrides ...string) ([]string, error) {
	binDir, err := filepath.Abs(filepath.Join(debugBwsRoot, "bin"))
	if err != nil {
		return nil, fmt.Errorf("debug bws bin dir: %w", err)
	}
	path := binDir + string(os.PathListSeparator) + os.Getenv("PATH")
	return goTestEnv(append([]string{"PATH=" + path}, overrides...)...), nil
}

// EnsureDebugBws installs a cargo --debug bws 2.1.0 into ./tools/bws-debug.
// Release bws sets https_only; the debug build is needed so CLI tests can talk
// to the HTTP mock. Skips if the prefix already contains bws 2.1.0.
func EnsureDebugBws() error {
	if debugBwsInstalledVersion() == debugBwsVersion {
		fmt.Printf("Using existing debug bws %s at %s\n", debugBwsVersion, debugBwsBin())
		return nil
	}

	if err := requireMiseCargo(); err != nil {
		return err
	}

	root, err := filepath.Abs(debugBwsRoot)
	if err != nil {
		return fmt.Errorf("debug bws install root: %w", err)
	}
	targetDir, err := filepath.Abs(debugBwsTargetDir())
	if err != nil {
		return fmt.Errorf("debug bws cargo target dir: %w", err)
	}
	if err := os.MkdirAll(targetDir, 0o755); err != nil {
		return fmt.Errorf("creating %s: %w", targetDir, err)
	}

	fmt.Printf("Installing debug bws %s into %s (CARGO_TARGET_DIR=%s)...\n", debugBwsVersion, root, targetDir)
	cmd := exec.Command("mise", "exec", "--", "cargo", "install", "--debug", "--locked", "--version", debugBwsVersion, "--root", root, "bws")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = overlayEnv("CARGO_TARGET_DIR=" + targetDir)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cargo install --debug bws %s: %w", debugBwsVersion, err)
	}

	if got := debugBwsInstalledVersion(); got != debugBwsVersion {
		return fmt.Errorf("installed debug bws version %q, want %s", got, debugBwsVersion)
	}
	fmt.Printf("Installed debug bws %s at %s\n", debugBwsVersion, debugBwsBin())
	return nil
}

type Test mg.Namespace
type Setup mg.Namespace

func Build() error {
	mg.Deps(InstallDeps)

	fmt.Println("Building the provider...")
	cmd := exec.Command("go", "build")
	return cmd.Run()
}

// Install the dependencies for the project.
func InstallDeps() error {
	fmt.Println("Installing dependencies...")
	cmd := exec.Command("go", "mod", "download")
	return cmd.Run()
}

// Run all tests.
func (t Test) All() error {
	mg.Deps(InstallDeps)

	err := t.Docs()
	if err != nil {
		return err
	}

	err = t.Offline()
	if err != nil {
		return err
	}

	err = t.IntegrationPwdVaultwardenWithEmbeddedClient()
	if err != nil {
		return err
	}

	err = t.IntegrationPwdOfficialWithEmbeddedClient()
	if err != nil {
		return err
	}

	err = t.IntegrationBwsOfficialWithEmbeddedClient()
	if err != nil {
		return err
	}

	err = t.IntegrationBwsOfficialWithCLI()
	if err != nil {
		return err
	}

	err = t.IntegrationBwsMockedWithEmbeddedClient()
	if err != nil {
		return err
	}

	err = t.IntegrationBwsMockedWithCLI()
	if err != nil {
		return err
	}

	err = t.IntegrationPwdVaultwardenWithCLI()
	if err != nil {
		return err
	}

	return nil
}

func officialSweepEnv() []string {
	return []string{"TEST_BACKEND=official", "TEST_EXPERIMENTAL_EMBEDDED_CLIENT=1"}
}

// Sweep leftover Password Manager resources from the official Bitwarden test account.
func (t Test) SweepOfficialPasswordManager() error {
	mg.Deps(InstallDeps)

	fmt.Println("Sweeping leftover official Password Manager test resources...")
	cmd := exec.Command("go", "test", "-count=1", "-timeout", "5m", "-tags=integration", "-run", "^$", "./internal/provider/")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = overlayEnv(officialSweepEnv()...)
	return cmd.Run()
}

// Sweep leftover Secrets Manager resources from the official Bitwarden test account.
func (t Test) SweepOfficialSecretsManager() error {
	mg.Deps(InstallDeps)

	fmt.Println("Sweeping leftover official Secrets Manager test resources...")
	cmd := exec.Command("go", "test", "-count=1", "-timeout", "5m", "-tags=integrationBws", "-run", "^$", "./internal/provider/")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = overlayEnv(officialSweepEnv()...)
	return cmd.Run()
}

// Run Password Manager integration tests with embedded client on bitwarden.com.
func (t Test) IntegrationPwdOfficialWithEmbeddedClient() error {
	return t.IntegrationPwdOfficialWithEmbeddedClientArgs("")
}

// Like test:integrationPwdOfficialWithEmbeddedClient but with a test pattern.
func (t Test) IntegrationPwdOfficialWithEmbeddedClientArgs(testPattern string) error {
	mg.Deps(InstallDeps)

	fmt.Println("Running Password Manager integration tests with embedded client on official bitwarden.com instances...")
	args := []string{"test", "-count=1", "-v", "-race", "-coverprofile=profile.cov", "-tags=integration", "-coverpkg=./...", "./..."}
	if testPattern != "" {
		args = append(args, "-run", testPattern, "-timeout", "1m")
	} else {
		args = append(args, "-timeout", "30m")
	}
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = goTestEnv("TEST_BACKEND=official", "TEST_EXPERIMENTAL_EMBEDDED_CLIENT=1")
	return cmd.Run()
}

// Run Bitwarden Secrets integration tests with CLI on bitwarden.com.
func (t Test) IntegrationBwsOfficialWithCLI() error {
	return t.IntegrationBwsOfficialWithCLIArgs("")
}

// Like test:integrationBwsOfficialWithCLI but with a test pattern.
func (t Test) IntegrationBwsOfficialWithCLIArgs(testPattern string) error {
	mg.Deps(InstallDeps, EnsureDebugBws)

	fmt.Println("Running Bitwarden Secrets integration tests with CLI on official bitwarden.com instances...")
	args := []string{"test", "-count=1", "-v", "-race", "-coverprofile=profile.cov", "-tags=integrationBws", "-coverpkg=./...", "./..."}
	if testPattern != "" {
		args = append(args, "-run", testPattern, "-timeout", "1m")
	} else {
		args = append(args, "-timeout", "20m")
	}
	env, err := goTestEnvWithDebugBws("TEST_BACKEND=official", "TEST_EXPERIMENTAL_EMBEDDED_CLIENT=0")
	if err != nil {
		return err
	}
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env
	return cmd.Run()
}

// Run Bitwarden Secrets integration tests with embedded client on bitwarden.com.
func (t Test) IntegrationBwsOfficialWithEmbeddedClient() error {
	return t.IntegrationBwsOfficialWithEmbeddedClientArgs("")
}

// Like test:integrationBwsOfficial but with a test pattern.
func (t Test) IntegrationBwsOfficialWithEmbeddedClientArgs(testPattern string) error {
	mg.Deps(InstallDeps)

	fmt.Println("Running Bitwarden Secrets integration tests with embedded client on official bitwarden.com instances...")
	args := []string{"test", "-count=1", "-v", "-race", "-coverprofile=profile.cov", "-tags=integrationBws", "-coverpkg=./...", "./..."}
	if testPattern != "" {
		args = append(args, "-run", testPattern, "-timeout", "1m")
	} else {
		args = append(args, "-timeout", "20m")
	}
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = goTestEnv("TEST_BACKEND=official", "TEST_EXPERIMENTAL_EMBEDDED_CLIENT=1")
	return cmd.Run()
}

// Run Bitwarden Secrets integration tests with embedded client on mocked backend.
func (t Test) IntegrationBwsMockedWithEmbeddedClient() error {
	return t.IntegrationBwsMockedWithEmbeddedClientArgs("")
}

// Run certain Bitwarden Secrets integration tests with embedded client on mocked backend.
func (t Test) IntegrationBwsMockedWithEmbeddedClientArgs(testPattern string) error {
	mg.Deps(InstallDeps)

	fmt.Println("Running Bitwarden Secrets integration tests with embedded client on mocked backend...")
	args := []string{"test", "-count=1", "-v", "-race", "-coverprofile=profile.cov", "-tags=integrationBws", "-coverpkg=./...", "./..."}
	if testPattern != "" {
		args = append(args, "-run", testPattern, "-timeout", "30s")
	} else {
		args = append(args, "-timeout", "20m")
	}
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = goTestEnv("TEST_BACKEND=vaultwarden", "TEST_EXPERIMENTAL_EMBEDDED_CLIENT=1")
	return cmd.Run()
}

// Run Bitwarden Secrets integration tests with CLI on mocked backend.
func (t Test) IntegrationBwsMockedWithCLI() error {
	return t.IntegrationBwsMockedWithCLIArgs("")
}

// Run certain Bitwarden Secrets integration tests with CLI on mocked backend.
func (t Test) IntegrationBwsMockedWithCLIArgs(testPattern string) error {
	mg.Deps(InstallDeps, EnsureDebugBws)

	fmt.Println("Running Bitwarden Secrets integration tests with CLI on mocked backend...")
	args := []string{"test", "-count=1", "-v", "-race", "-coverprofile=profile.cov", "-tags=integrationBws", "-coverpkg=./...", "./..."}
	if testPattern != "" {
		args = append(args, "-run", testPattern, "-timeout", "30s")
	} else {
		args = append(args, "-timeout", "20m")
	}
	env, err := goTestEnvWithDebugBws("TEST_BACKEND=vaultwarden")
	if err != nil {
		return err
	}
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = env
	return cmd.Run()
}

// Run Password Manager integration tests with embedded client on locally running Vaultwarden instance.
func (t Test) IntegrationPwdVaultwardenWithEmbeddedClient() error {
	return t.IntegrationPwdVaultwardenWithEmbeddedClientArgs("")
}

// Like test:integrationPwdVaultwardenWithEmbeddedClient but with a test pattern.
func (t Test) IntegrationPwdVaultwardenWithEmbeddedClientArgs(testPattern string) error {
	mg.Deps(InstallDeps)

	fmt.Println("Running integration tests with embedded client on locally running Vaultwarden...")
	args := []string{"test", "-count=1", "-v", "-race", "-coverprofile=profile.cov", "-tags=integration", "-coverpkg=./...", "./..."}
	if testPattern != "" {
		args = append(args, "-run", testPattern, "-timeout", "60s")
	} else {
		args = append(args, "-timeout", "10m")
	}
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = goTestEnv("TEST_SERVER_URL=http://127.0.0.1:8000", "TEST_BACKEND=vaultwarden", "TEST_EXPERIMENTAL_EMBEDDED_CLIENT=1")
	return cmd.Run()
}

// Run Password Manager integration tests with CLI on locally running Vaultwarden instance.
func (t Test) IntegrationPwdVaultwardenWithCLI() error {
	return t.IntegrationPwdVaultwardenWithCLIArgs("")
}

// Like test:integrationPwdVaultwardenWithCLI but with a test pattern.
func (t Test) IntegrationPwdVaultwardenWithCLIArgs(testPattern string) error {
	mg.Deps(InstallDeps)

	fmt.Println("Running integration tests with CLI on locally running Vaultwarden...")
	args := []string{"test", "-count=1", "-v", "-race", "-coverprofile=profile.cov", "-tags=integration", "-coverpkg=./...", "./..."}
	if testPattern != "" {
		args = append(args, "-run", testPattern, "-timeout", "3m")
	} else {
		args = append(args, "-timeout", "60m")
	}
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = goTestEnv("TEST_REVERSE_PROXY_URL=http://127.0.0.1:8080", "TEST_SERVER_URL=http://127.0.0.1:8000", "TEST_BACKEND=vaultwarden", "TEST_EXPERIMENTAL_EMBEDDED_CLIENT=0")
	return cmd.Run()
}

// Run Password Manager integration tests with CLI on official bitwarden.com instance.
func (t Test) IntegrationPwdOfficialWithCLI() error {
	return t.IntegrationPwdOfficialWithCLIArgs("")
}

// Like test:integrationPwdOfficialWithCLI but with a test pattern.
func (t Test) IntegrationPwdOfficialWithCLIArgs(testPattern string) error {
	mg.Deps(InstallDeps)

	fmt.Println("Running integration tests with CLI on official bitwarden.com instance...")
	args := []string{"test", "-count=1", "-v", "-race", "-coverprofile=profile.cov", "-tags=integration", "-coverpkg=./...", "./..."}
	if testPattern != "" {
		args = append(args, "-run", testPattern, "-timeout", "3m")
	} else {
		args = append(args, "-timeout", "60m")
	}
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = goTestEnv("TEST_BACKEND=official", "TEST_EXPERIMENTAL_EMBEDDED_CLIENT=0")
	return cmd.Run()
}

// Run tests not requiring a running Bitwarden-compatible backend.
func (t Test) Offline() error {
	return t.OfflineArgs("")
}

// Like test:offline but with a test pattern.
func (Test) OfflineArgs(testPattern string) error {
	mg.Deps(InstallDeps)

	fmt.Println("Running offline tests...")
	args := []string{"test", "-v", "-race", "-coverprofile=profile.cov", "-tags=offline", "-timeout", "2m", "-coverpkg=./...", "./..."}
	if testPattern != "" {
		args = append(args, "-run", testPattern)
	}
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = goTestEnv("TEST_BACKEND=vaultwarden")
	return cmd.Run()
}

// Generate the documentation.
func Docs() error {
	fmt.Println("Generating documentation...")
	cmd := exec.Command("go", "run", "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	cmd = exec.Command("tofu", "fmt", "-recursive", "examples")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Validate the documentation.
func (t Test) Docs() error {
	fmt.Println("Validating documentation...")
	cmd := exec.Command("go", "run", "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@latest", "validate")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Start a Vaultwarden instance to run tests against.
func Vaultwarden() error {
	fmt.Println("Starting Vaultwarden and Nginx reverse proxy...")
	cmd := exec.Command("docker-compose", "up")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Remove local copies of test Vaults and clears the test results cache.
func Clean() error {
	fmt.Println("Cleaning...")
	os.RemoveAll("internal/provider/.bitwarden/data.json")

	cmd := exec.Command("go", "clean", "-cache")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

// Install Install the development version of the provider to ~/.terraformrc
func (Setup) Install() error {
	fmt.Println("Installing the development version of the provider...")
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	workDir, err := os.Getwd()
	if err != nil {
		return fmt.Errorf("failed to get current directory: %w", err)
	}

	terraformrcPath := filepath.Join(homeDir, ".terraformrc")

	// Check if file exists
	if _, err := os.Stat(terraformrcPath); err == nil {
		fmt.Printf("File %s already exists. Do you want to overwrite it? [y/N]: ", terraformrcPath)
		var response string
		fmt.Scanln(&response)
		if response != "y" && response != "Y" {
			return fmt.Errorf("operation cancelled by user")
		}
	}

	content := fmt.Sprintf(`provider_installation {
	dev_overrides {
		"maxlaverse/bitwarden" = "%s"
	}
	direct {}
}`, workDir)

	return os.WriteFile(terraformrcPath, []byte(content), 0644)
}

// Removes the entire ~/.terraformrc file.
func (Setup) Uninstall() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get home directory: %w", err)
	}

	terraformrcPath := filepath.Join(homeDir, ".terraformrc")
	return os.Remove(terraformrcPath)
}
