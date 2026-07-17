package runner

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/padok-team/burrito/internal/burrito/config"
)

type fakeExec struct {
	initErrs  []error
	initCalls []string
}

func (f *fakeExec) Init(workingDir string) error {
	f.initCalls = append(f.initCalls, workingDir)
	if len(f.initErrs) == 0 {
		return nil
	}
	err := f.initErrs[0]
	f.initErrs = f.initErrs[1:]
	return err
}

func (f *fakeExec) Plan(string) error {
	return nil
}

func (f *fakeExec) Apply(string) error {
	return nil
}

func (f *fakeExec) Show(string, string) ([]byte, error) {
	return []byte{}, nil
}

func (f *fakeExec) TenvName() string {
	return "terraform"
}

func (f *fakeExec) GetExecPath() string {
	return "/tmp/terraform"
}

func testRunnerConfig(t *testing.T) *config.Config {
	t.Helper()

	conf := config.TestConfig()
	conf.Runner.RepositoryPath = t.TempDir()
	conf.Hermitcrab.URL = "http://hermitcrab.local"
	return conf
}

func TestExecInitWithHermitcrabKeepsMirrorConfigWhenInitSucceeds(t *testing.T) {
	conf := testRunnerConfig(t)
	conf.Hermitcrab.Enabled = true
	_ = os.Unsetenv("TF_CLI_CONFIG_FILE")
	t.Cleanup(func() {
		_ = os.Unsetenv("TF_CLI_CONFIG_FILE")
	})

	runnerInstance := New(conf)
	if err := runnerInstance.EnableHermitcrab(); err != nil {
		t.Fatalf("EnableHermitcrab() error = %v", err)
	}

	exec := &fakeExec{}
	runnerInstance.exec = exec
	runnerInstance.workingDir = filepath.Join(conf.Runner.RepositoryPath, "content")

	if err := runnerInstance.ExecInit(); err != nil {
		t.Fatalf("ExecInit() error = %v", err)
	}
	expectedCalls := []string{runnerInstance.workingDir}
	if !reflect.DeepEqual(exec.initCalls, expectedCalls) {
		t.Fatalf("Init() calls = %v, want %v", exec.initCalls, expectedCalls)
	}
	configPath := filepath.Join(conf.Runner.RepositoryPath, "config.tfrc")
	if got := os.Getenv("TF_CLI_CONFIG_FILE"); got != configPath {
		t.Fatalf("TF_CLI_CONFIG_FILE = %q, want %q", got, configPath)
	}
	if _, err := os.Stat(configPath); err != nil {
		t.Fatalf("network mirror config should exist: %v", err)
	}
}

func TestExecInitWithHermitcrabRetriesWithoutMirrorConfigWhenInitFails(t *testing.T) {
	conf := testRunnerConfig(t)
	conf.Hermitcrab.Enabled = true
	_ = os.Unsetenv("TF_CLI_CONFIG_FILE")
	t.Cleanup(func() {
		_ = os.Unsetenv("TF_CLI_CONFIG_FILE")
	})

	runnerInstance := New(conf)
	if err := runnerInstance.EnableHermitcrab(); err != nil {
		t.Fatalf("EnableHermitcrab() error = %v", err)
	}

	exec := &fakeExec{initErrs: []error{errors.New("mirror init failed"), nil}}
	runnerInstance.exec = exec
	runnerInstance.workingDir = filepath.Join(conf.Runner.RepositoryPath, "content")

	if err := runnerInstance.ExecInit(); err != nil {
		t.Fatalf("ExecInit() error = %v", err)
	}
	expectedCalls := []string{runnerInstance.workingDir, runnerInstance.workingDir}
	if !reflect.DeepEqual(exec.initCalls, expectedCalls) {
		t.Fatalf("Init() calls = %v, want %v", exec.initCalls, expectedCalls)
	}
	if got := os.Getenv("TF_CLI_CONFIG_FILE"); got != "" {
		t.Fatalf("TF_CLI_CONFIG_FILE = %q, want empty", got)
	}
	if _, err := os.Stat(filepath.Join(conf.Runner.RepositoryPath, "config.tfrc")); !os.IsNotExist(err) {
		t.Fatalf("network mirror config should be removed, stat error = %v", err)
	}
}

func TestExecInitWithHermitcrabReturnsRetryErrorWhenDirectInitFails(t *testing.T) {
	conf := testRunnerConfig(t)
	conf.Hermitcrab.Enabled = true
	_ = os.Unsetenv("TF_CLI_CONFIG_FILE")
	t.Cleanup(func() {
		_ = os.Unsetenv("TF_CLI_CONFIG_FILE")
	})

	runnerInstance := New(conf)
	if err := runnerInstance.EnableHermitcrab(); err != nil {
		t.Fatalf("EnableHermitcrab() error = %v", err)
	}

	retryErr := errors.New("direct init failed")
	exec := &fakeExec{initErrs: []error{errors.New("mirror init failed"), retryErr}}
	runnerInstance.exec = exec
	runnerInstance.workingDir = filepath.Join(conf.Runner.RepositoryPath, "content")

	if err := runnerInstance.ExecInit(); !errors.Is(err, retryErr) {
		t.Fatalf("ExecInit() error = %v, want %v", err, retryErr)
	}
	expectedCalls := []string{runnerInstance.workingDir, runnerInstance.workingDir}
	if !reflect.DeepEqual(exec.initCalls, expectedCalls) {
		t.Fatalf("Init() calls = %v, want %v", exec.initCalls, expectedCalls)
	}
	if got := os.Getenv("TF_CLI_CONFIG_FILE"); got != "" {
		t.Fatalf("TF_CLI_CONFIG_FILE = %q, want empty", got)
	}
	if _, err := os.Stat(filepath.Join(conf.Runner.RepositoryPath, "config.tfrc")); !os.IsNotExist(err) {
		t.Fatalf("network mirror config should be removed, stat error = %v", err)
	}
}

func TestExecInitWithoutHermitcrabReturnsInitErrorWithoutRetrying(t *testing.T) {
	conf := testRunnerConfig(t)
	initErr := errors.New("init failed")
	exec := &fakeExec{initErrs: []error{initErr}}
	runnerInstance := New(conf)
	runnerInstance.exec = exec
	runnerInstance.workingDir = filepath.Join(conf.Runner.RepositoryPath, "content")

	if err := runnerInstance.ExecInit(); !errors.Is(err, initErr) {
		t.Fatalf("ExecInit() error = %v, want %v", err, initErr)
	}
	expectedCalls := []string{runnerInstance.workingDir}
	if !reflect.DeepEqual(exec.initCalls, expectedCalls) {
		t.Fatalf("Init() calls = %v, want %v", exec.initCalls, expectedCalls)
	}
}
