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

	err := t.Offline()
	if err != nil {
		return err
	}

	err = t.IntegrationVaultwardenWithEmbeddedClient()
	if err != nil {
		return err
	}

	err = t.IntegrationOfficialWithEmbeddedClient()
	if err != nil {
		return err
	}

	err = t.IntegrationVaultwardenWithCLI()
	if err != nil {
		return err
	}

	return nil
}

// Run the integration tests with the embedded client on the official bitwarden.com backend.
func (t Test) IntegrationOfficialWithEmbeddedClient() error {
	return t.IntegrationOfficialWithEmbeddedClientArgs("")
}

// Run certain integration tests with the embedded client on the official bitwarden.com backend.
func (t Test) IntegrationOfficialWithEmbeddedClientArgs(testPattern string) error {
	mg.Deps(InstallDeps)

	fmt.Println("Running integration tests with embedded client on official bitwarden.com backend...")
	args := []string{"test", "./...", "--tags", "integration", "-v", "-race", "-timeout", "20m"}
	if testPattern != "" {
		args = append(args, "-run", testPattern)
	}
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "TF_ACC=1", "TEST_BACKEND=official", "TEST_PROVIDER_SERVER_URL=https://vault.bitwarden.com", "TEST_PROVIDER_EXPERIMENTAL_EMBEDDED_CLIENT=1")
	return cmd.Run()
}

// Run the integration tests with the embedded client on a locally running Vaultwarden instance.
func (t Test) IntegrationVaultwardenWithEmbeddedClient() error {
	return t.IntegrationVaultwardenWithEmbeddedClientArgs("")
}

// Run certain integration tests with the embedded client on a locally running Vaultwarden instance.
func (Test) IntegrationVaultwardenWithEmbeddedClientArgs(testPattern string) error {
	mg.Deps(InstallDeps)

	fmt.Println("Running integration tests with embedded client on locally running Vaultwarden...")
	args := []string{"test", "./...", "--tags", "integration", "-v", "-race", "-timeout", "10m"}
	if testPattern != "" {
		args = append(args, "-run", testPattern)
	}
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "TF_ACC=1", "TEST_PROVIDER_SERVER_URL=http://127.0.0.1:8000", "TEST_BACKEND=vaultwarden", "TEST_PROVIDER_EXPERIMENTAL_EMBEDDED_CLIENT=1")
	return cmd.Run()
}

// Run the integration tests with the CLI on a locally running Vaultwarden instance.
func (t Test) IntegrationVaultwardenWithCLI() error {
	return t.IntegrationVaultwardenWithCLIArgs("")
}

// Run certain integration tests with the CLI on a locally running Vaultwarden instance.
func (Test) IntegrationVaultwardenWithCLIArgs(testPattern string) error {
	mg.Deps(InstallDeps)

	fmt.Println("Running integration tests with CLI on locally running Vaultwarden...")
	args := []string{"test", "./...", "--tags", "integration", "-v", "-race", "-timeout", "60m"}
	if testPattern != "" {
		args = append(args, "-run", testPattern)
	}
	cmd := exec.Command("go", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "TEST_PROVIDER_REVERSE_PROXY_URL=http://127.0.0.1:8080")
	cmd.Env = append(cmd.Env, "TF_ACC=1", "TEST_PROVIDER_SERVER_URL=http://127.0.0.1:8000", "TEST_BACKEND=vaultwarden", "TEST_PROVIDER_EXPERIMENTAL_EMBEDDED_CLIENT=0")
	return cmd.Run()
}

// Run the tests not requiring a running Bitwarden-compatible backend.
func (Test) Offline() error {
	mg.Deps(InstallDeps)

	fmt.Println("Running offline tests...")
	cmd := exec.Command("go", "test", "./...", "--tags", "offline", "-v", "-race", "-timeout", "30s")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = os.Environ()
	cmd.Env = append(cmd.Env, "TF_ACC=1", "TEST_BACKEND=vaultwarden")
	return cmd.Run()
}

// Generate and formats the documentation for the project.
func GenerateDocumentation() error {
	fmt.Println("Generating documentation...")
	cmd := exec.Command("go", "run", "github.com/hashicorp/terraform-plugin-docs/cmd/tfplugindocs@v0.19.0")
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
