package command

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"
)

func TestGenerateBashIntegration(t *testing.T) {
	content := generateBashIntegration("/home/user/.jabba")

	if !strings.Contains(content, "JABBA_HOME") {
		t.Error("Bash integration should contain JABBA_HOME")
	}
	if !strings.Contains(content, "jabba()") {
		t.Error("Bash integration should contain jabba function")
	}
	if !strings.Contains(content, "msys*|cygwin*|win32") {
		t.Error("Bash integration should contain Windows detection")
	}
}

func TestEnsureCompletionIntegration(t *testing.T) {
	path := filepath.Join(t.TempDir(), "jabba.sh")
	original := "# custom integration without a trailing newline"
	if err := os.WriteFile(path, []byte(original), 0644); err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 2; i++ {
		if err := ensureCompletionIntegration(path); err != nil {
			t.Fatal(err)
		}
	}
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(content) != original+bashCompletionIntegration {
		t.Fatal("existing integration must be preserved and completion appended exactly once")
	}
}

func TestCompletionIntegrationLoadersMatch(t *testing.T) {
	installer, err := os.ReadFile("../install.sh")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(installer), bashCompletionIntegration) {
		t.Fatal("install.sh and jabba init must load completion identically")
	}
	content := generateBashIntegration("/home/user/jabba home")
	if !strings.HasSuffix(content, bashCompletionIntegration) {
		t.Fatal("new integrations must include completion")
	}
}

func TestGeneratePowerShellIntegration(t *testing.T) {
	content := generatePowerShellIntegration("C:\\Users\\test\\.jabba")

	if !strings.Contains(content, "$env:JABBA_HOME") {
		t.Error("PowerShell integration should contain JABBA_HOME")
	}
	if !strings.Contains(content, "function jabba") {
		t.Error("PowerShell integration should contain jabba function")
	}
}

func TestGenerateFishIntegration(t *testing.T) {
	content := generateFishIntegration("/home/user/.jabba")

	if !strings.Contains(content, "JABBA_HOME") {
		t.Error("Fish integration should contain JABBA_HOME")
	}
	if !strings.Contains(content, "function jabba") {
		t.Error("Fish integration should contain jabba function")
	}
}

func TestGenerateNushellIntegration(t *testing.T) {
	content := generateNushellIntegration("/home/user/.jabba")

	if !strings.Contains(content, "JABBA_HOME") {
		t.Error("Nushell integration should contain JABBA_HOME")
	}
	if !strings.Contains(content, "def --env jabba") {
		t.Error("Nushell integration should contain jabba function")
	}
}

func TestCreateIntegrationFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	tmpDir := t.TempDir()
	originalHome := os.Getenv("JABBA_HOME")
	os.Setenv("JABBA_HOME", tmpDir)
	defer os.Setenv("JABBA_HOME", originalHome)

	integrationFile := filepath.Join(tmpDir, "jabba.sh")

	err := createIntegrationFile("bash", integrationFile)
	if err != nil {
		t.Fatalf("createIntegrationFile() error = %v", err)
	}

	// Check file exists
	if _, err := os.Stat(integrationFile); os.IsNotExist(err) {
		t.Error("Integration file should be created")
	}

	// Check content
	content, err := os.ReadFile(integrationFile)
	if err != nil {
		t.Fatalf("Failed to read integration file: %v", err)
	}

	if !strings.Contains(string(content), "jabba()") {
		t.Error("Integration file should contain jabba function")
	}
}

func TestUpdateRCFile(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	tmpDir := t.TempDir()
	rcFile := filepath.Join(tmpDir, ".bashrc")
	integrationFile := filepath.Join(tmpDir, "jabba.sh")

	// Test creating new RC file
	err := updateRCFile("bash", rcFile, integrationFile)
	if err != nil {
		t.Fatalf("updateRCFile() error = %v", err)
	}

	// Check file exists
	if _, err := os.Stat(rcFile); os.IsNotExist(err) {
		t.Error("RC file should be created")
	}

	// Check content
	content, err := os.ReadFile(rcFile)
	if err != nil {
		t.Fatalf("Failed to read RC file: %v", err)
	}

	if !strings.Contains(string(content), "jabba.sh") {
		t.Error("RC file should contain source line for jabba.sh")
	}
	if !strings.Contains(string(content), "source") {
		t.Error("RC file should contain source command")
	}
}

func TestUpdateRCFileExisting(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	tmpDir := t.TempDir()
	rcFile := filepath.Join(tmpDir, ".bashrc")
	integrationFile := filepath.Join(tmpDir, "jabba.sh")

	// Create existing RC file with content
	existingContent := "export PATH=/usr/bin:$PATH\n"
	os.WriteFile(rcFile, []byte(existingContent), 0644)

	err := updateRCFile("bash", rcFile, integrationFile)
	if err != nil {
		t.Fatalf("updateRCFile() error = %v", err)
	}

	// Check content preserved and source line added
	content, err := os.ReadFile(rcFile)
	if err != nil {
		t.Fatalf("Failed to read RC file: %v", err)
	}

	contentStr := string(content)
	if !strings.Contains(contentStr, existingContent) {
		t.Error("RC file should preserve existing content")
	}
	if !strings.Contains(contentStr, "jabba.sh") {
		t.Error("RC file should contain source line for jabba.sh")
	}
}

func TestInit(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	tmpDir := t.TempDir()
	originalHome := os.Getenv("JABBA_HOME")
	originalShell := os.Getenv("SHELL")
	homeDir := os.Getenv("HOME")

	os.Setenv("JABBA_HOME", tmpDir)
	os.Setenv("SHELL", "/bin/bash")
	os.Setenv("HOME", tmpDir)

	defer func() {
		os.Setenv("JABBA_HOME", originalHome)
		os.Setenv("SHELL", originalShell)
		os.Setenv("HOME", homeDir)
	}()

	err := Init()
	if err != nil {
		t.Fatalf("Init() error = %v", err)
	}

	// Check integration file created
	integrationFile := filepath.Join(tmpDir, "jabba.sh")
	if _, err := os.Stat(integrationFile); os.IsNotExist(err) {
		t.Error("Init() should create integration file")
	}

	// Check RC file updated
	rcFile := filepath.Join(tmpDir, ".bashrc")
	if _, err := os.Stat(rcFile); os.IsNotExist(err) {
		t.Error("Init() should create/update RC file")
	}

	content, _ := os.ReadFile(rcFile)
	if !strings.Contains(string(content), "jabba.sh") {
		t.Error("Init() should add source line to RC file")
	}
}

func TestInitIdempotent(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Skipping Unix-specific test on Windows")
	}

	tmpDir := t.TempDir()
	originalHome := os.Getenv("JABBA_HOME")
	originalShell := os.Getenv("SHELL")
	homeDir := os.Getenv("HOME")

	os.Setenv("JABBA_HOME", tmpDir)
	os.Setenv("SHELL", "/bin/bash")
	os.Setenv("HOME", tmpDir)

	defer func() {
		os.Setenv("JABBA_HOME", originalHome)
		os.Setenv("SHELL", originalShell)
		os.Setenv("HOME", homeDir)
	}()

	// Run init twice
	err := Init()
	if err != nil {
		t.Fatalf("Init() first run error = %v", err)
	}

	rcFile := filepath.Join(tmpDir, ".bashrc")
	contentAfterFirst, _ := os.ReadFile(rcFile)

	err = Init()
	if err != nil {
		t.Fatalf("Init() second run error = %v", err)
	}

	// Check RC file not duplicated
	contentAfterSecond, _ := os.ReadFile(rcFile)

	firstCount := strings.Count(string(contentAfterFirst), "jabba.sh")
	secondCount := strings.Count(string(contentAfterSecond), "jabba.sh")

	if firstCount != secondCount {
		t.Error("Init() should be idempotent - should not add duplicate source lines")
	}
}
