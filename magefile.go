//go:build mage
// +build mage

package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/magefile/mage/mg"
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
	return append([]string{"TF_ACC_TERRAFORM_VERSION=1.9.0"}, base...)
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

// Run Password Manager integration tests with embedded client on bitwarden.com.
func (t Test) IntegrationPwdOfficialWithEmbeddedClient() error {
	return t.IntegrationPwdOfficialWithEmbeddedClientArgs("")
}

// Like test:integrationPwdOfficialWithEmbeddedClient but with a test pattern.
func (t Test) IntegrationPwdOfficialWithEmbeddedClientArgs(testPattern string) error {
	mg.Deps(InstallDeps)

	fmt.Println("Running Password Manager integration tests with embedded client on official bitwarden.com instances...")
	args := []string{"test", "-v", "-race", "-coverprofile=profile.cov", "-tags=integration", "-coverpkg=./...", "./..."}
	if testPattern != "" {
		args = append(args, "-run", testPattern, "-timeout", "1m")
	} else {
		args = append(args, "-timeout", "30m")
	}
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, accTestEnv()...)
	cmd.Env = append(cmd.Env, "TEST_BACKEND=official", "TEST_EXPERIMENTAL_EMBEDDED_CLIENT=1")
	return cmd.Run()
}

// Run Bitwarden Secrets integration tests with CLI on bitwarden.com.
func (t Test) IntegrationBwsOfficialWithCLI() error {
	return t.IntegrationBwsOfficialWithCLIArgs("")
}

// Like test:integrationBwsOfficialWithCLI but with a test pattern.
func (t Test) IntegrationBwsOfficialWithCLIArgs(testPattern string) error {
	mg.Deps(InstallDeps)

	fmt.Println("Running Bitwarden Secrets integration tests with CLI on official bitwarden.com instances...")
	args := []string{"test", "-v", "-race", "-coverprofile=profile.cov", "-tags=integrationBws", "-coverpkg=./...", "./..."}
	if testPattern != "" {
		args = append(args, "-run", testPattern, "-timeout", "1m")
	} else {
		args = append(args, "-timeout", "20m")
	}
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, accTestEnv()...)
	cmd.Env = append(cmd.Env, "TEST_BACKEND=official", "TEST_EXPERIMENTAL_EMBEDDED_CLIENT=0")
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
	args := []string{"test", "-v", "-race", "-coverprofile=profile.cov", "-tags=integrationBws", "-coverpkg=./...", "./..."}
	if testPattern != "" {
		args = append(args, "-run", testPattern, "-timeout", "1m")
	} else {
		args = append(args, "-timeout", "20m")
	}
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, accTestEnv()...)
	cmd.Env = append(cmd.Env, "TEST_BACKEND=official")
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
	args := []string{"test", "-v", "-race", "-coverprofile=profile.cov", "-tags=integrationBws", "-coverpkg=./...", "./..."}
	if testPattern != "" {
		args = append(args, "-run", testPattern, "-timeout", "30s")
	} else {
		args = append(args, "-timeout", "20m")
	}
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, accTestEnv()...)
	cmd.Env = append(cmd.Env, "TEST_BACKEND=vaultwarden", "TEST_EXPERIMENTAL_EMBEDDED_CLIENT=1")
	return cmd.Run()
}

// Run Bitwarden Secrets integration tests with CLI on mocked backend.
func (t Test) IntegrationBwsMockedWithCLI() error {
	return t.IntegrationBwsMockedWithCLIArgs("")
}

// Run certain Bitwarden Secrets integration tests with CLI on mocked backend.
func (t Test) IntegrationBwsMockedWithCLIArgs(testPattern string) error {
	mg.Deps(InstallDeps)

	fmt.Println("Running Bitwarden Secrets integration tests with embedded client on mocked backend...")
	args := []string{"test", "-v", "-race", "-coverprofile=profile.cov", "-tags=integrationBws", "-coverpkg=./...", "./..."}
	if testPattern != "" {
		args = append(args, "-run", testPattern, "-timeout", "30s")
	} else {
		args = append(args, "-timeout", "20m")
	}
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, accTestEnv()...)
	cmd.Env = append(cmd.Env, "TEST_BACKEND=vaultwarden")
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
	args := []string{"test", "-v", "-race", "-coverprofile=profile.cov", "-tags=integration", "-coverpkg=./...", "./..."}
	if testPattern != "" {
		args = append(args, "-run", testPattern, "-timeout", "60s")
	} else {
		args = append(args, "-timeout", "10m")
	}
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, accTestEnv()...)
	cmd.Env = append(cmd.Env, "TEST_SERVER_URL=http://127.0.0.1:8000", "TEST_BACKEND=vaultwarden", "TEST_EXPERIMENTAL_EMBEDDED_CLIENT=1")
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
	args := []string{"test", "-v", "-race", "-coverprofile=profile.cov", "-tags=integration", "-coverpkg=./...", "./..."}
	if testPattern != "" {
		args = append(args, "-run", testPattern, "-timeout", "2m")
	} else {
		args = append(args, "-timeout", "60m")
	}
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "TEST_REVERSE_PROXY_URL=http://127.0.0.1:8080")
	cmd.Env = append(cmd.Env, accTestEnv()...)
	cmd.Env = append(cmd.Env, "TEST_SERVER_URL=http://127.0.0.1:8000", "TEST_BACKEND=vaultwarden", "TEST_EXPERIMENTAL_EMBEDDED_CLIENT=0")
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
	args := []string{"test", "-v", "-race", "-coverprofile=profile.cov", "-tags=offline", "-timeout", "30s", "-coverpkg=./...", "./..."}
	if testPattern != "" {
		args = append(args, "-run", testPattern)
	}
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, accTestEnv()...)
	cmd.Env = append(cmd.Env, "TEST_BACKEND=vaultwarden")
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
