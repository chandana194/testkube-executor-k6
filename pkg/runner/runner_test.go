package runner

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	"github.com/kubeshop/testkube/pkg/api/v1/testkube"
	"github.com/stretchr/testify/assert"
)

func TestRunFiles(t *testing.T) {
	// setup
	tempDir := os.TempDir()
	os.Setenv("RUNNER_DATADIR", tempDir)

	t.Run("Run k6 with simple script", func(t *testing.T) {
		// given
		runner := NewRunner()
		execution := testkube.NewQueuedExecution()
		execution.Content = testkube.NewStringTestContent("")
		writeTestContent(t, tempDir, "../../examples/k6-test-script.js")

		// when
		result, err := runner.Run(*execution)

		// then
		assert.NoError(t, err)
		assert.Equal(t, result.Status, testkube.ExecutionStatusSuccess)
		assert.Len(t, result.Steps, 1)
	})

	t.Run("Run k6 with arguments and simple script", func(t *testing.T) {
		// given
		runner := NewRunner()
		execution := testkube.NewQueuedExecution()
		execution.Content = testkube.NewStringTestContent("")
		execution.Args = []string{"--vus", "2", "--duration", "1s"}
		writeTestContent(t, tempDir, "../../examples/k6-test-script.js")

		// when
		result, err := runner.Run(*execution)

		// then
		assert.NoError(t, err)
		assert.Equal(t, result.Status, testkube.ExecutionStatusSuccess)
		assert.Len(t, result.Steps, 1)
	})

	t.Run("Run k6 with ENV variables and script", func(t *testing.T) {
		// given
		runner := NewRunner()
		execution := testkube.NewQueuedExecution()
		execution.Content = testkube.NewStringTestContent("")
		execution.Envs = map[string]string{"TARGET_HOSTNAME": "kubeshop.github.io"}
		writeTestContent(t, tempDir, "../../examples/k6-test-environment.js")

		// when
		result, err := runner.Run(*execution)

		// then
		assert.NoError(t, err)
		assert.Equal(t, result.Status, testkube.ExecutionStatusSuccess)
		assert.Len(t, result.Steps, 1)
	})
}

func TestRunAdvanced(t *testing.T) {
	// setup
	tempDir := os.TempDir()
	os.Setenv("RUNNER_DATADIR", tempDir)

	t.Run("Run k6 with checks and thresholds", func(t *testing.T) {
		// given
		runner := NewRunner()
		execution := testkube.NewQueuedExecution()
		execution.Content = testkube.NewStringTestContent("")
		writeTestContent(t, tempDir, "../../examples/k6-test-thresholds.js")

		// when
		result, err := runner.Run(*execution)

		// then
		assert.NoError(t, err)
		assert.Equal(t, result.ErrorMessage, "some thresholds have failed")
		assert.Equal(t, result.Status, testkube.ExecutionStatusError)
		assert.Len(t, result.Steps, 1)
	})
}

func TestRunDirs(t *testing.T) {
	// setup
	tempDir, _ := os.MkdirTemp("", "*")
	os.Setenv("RUNNER_DATADIR", tempDir)

	repoDir := filepath.Join(tempDir, "repo")
	os.Mkdir(repoDir, 0755)

	k6Script, err := ioutil.ReadFile("../../examples/k6-test-script.js")
	if err != nil {
		assert.FailNow(t, "Unable to read k6 test script")
	}

	err = ioutil.WriteFile(filepath.Join(repoDir, "k6-test-script.js"), k6Script, 0644)
	if err != nil {
		assert.FailNow(t, "Unable to write k6 runner test content file")
	}

	t.Run("Run k6 from directory with script argument", func(t *testing.T) {
		// given
		runner := NewRunner()
		execution := testkube.NewQueuedExecution()
		execution.Content = &testkube.TestContent{
			Type_: string(testkube.TestContentTypeGitDir),
			Repository: &testkube.Repository{
				Uri:    "https://github.com/kubeshop/testkube-executor-k6.git",
				Branch: "main",
			},
		}
		execution.Args = []string{"--duration", "1s", "k6-test-script.js"}

		// when
		result, err := runner.Run(*execution)

		// then
		assert.NoError(t, err)
		assert.Equal(t, result.Status, testkube.ExecutionStatusSuccess)
		assert.Len(t, result.Steps, 1)
	})
}

func TestRunErrors(t *testing.T) {

	t.Run("Run k6 with no script", func(t *testing.T) {
		// setup
		os.Setenv("RUNNER_DATADIR", ".")

		// given
		runner := NewRunner()
		execution := testkube.NewQueuedExecution()
		execution.Content = testkube.NewStringTestContent("")

		// when
		result, err := runner.Run(*execution)

		// then
		assert.NoError(t, err)
		assert.Equal(t, result.Status, testkube.ExecutionStatusError)
		assert.Contains(t, result.ErrorMessage, "255")
	})

	t.Run("Run k6 with invalid arguments", func(t *testing.T) {
		// setup
		os.Setenv("RUNNER_DATADIR", ".")

		runner := NewRunner()
		execution := testkube.NewQueuedExecution()
		execution.Content = testkube.NewStringTestContent("")
		execution.Args = []string{"--vues", "2", "--duration", "5"}

		// when
		result, err := runner.Run(*execution)

		// then
		assert.NoError(t, err)
		assert.Equal(t, result.Status, testkube.ExecutionStatusError)
		assert.Contains(t, result.ErrorMessage, "255")
	})

	t.Run("Run k6 from directory with missing script arg", func(t *testing.T) {
		// setup
		os.Setenv("RUNNER_DATADIR", ".")

		// given
		runner := NewRunner()
		execution := testkube.NewQueuedExecution()
		execution.Content = &testkube.TestContent{
			Type_: string(testkube.TestContentTypeGitDir),
			Repository: &testkube.Repository{
				Uri:    "https://github.com/kubeshop/testkube-executor-k6.git",
				Branch: "main",
				Path:   "examples",
			},
		}
		execution.Args = []string{}

		// when
		result, err := runner.Run(*execution)

		// then
		assert.NoError(t, err)
		assert.Equal(t, result.Status, testkube.ExecutionStatusError)
		assert.Contains(t, result.ErrorMessage, "not found")
	})
}

func TestParse(t *testing.T) {
	// setup
	summary, err := ioutil.ReadFile("../../examples/k6-test-summary.txt")
	if err != nil {
		assert.FailNow(t, "Unable to read k6 test summary")
	}

	t.Run("Parse Scenario Name", func(t *testing.T) {
		name := parseScenarioName(string(summary))
		assert.Equal(t, "* default: 1 iterations for each of 1 VUs (maxDuration: 10m0s, gracefulStop: 30s)", name)
	})

	t.Run("Parse Scenario Duration", func(t *testing.T) {
		duration := parseScenarioDuration(string(summary))
		assert.Equal(t, "00m01.1s/10m0s", duration)
	})

	t.Run("Parse Scenario Status", func(t *testing.T) {
		status := parseScenarioStatus(string(summary))
		assert.Equal(t, "success", status)
	})
}

func writeTestContent(t *testing.T, dir string, testScript string) {
	k6Script, err := ioutil.ReadFile(testScript)
	if err != nil {
		assert.FailNow(t, "Unable to read k6 test script")
	}

	err = ioutil.WriteFile(filepath.Join(dir, "test-content"), k6Script, 0644)
	if err != nil {
		assert.FailNow(t, "Unable to write k6 runner test content file")
	}
}
