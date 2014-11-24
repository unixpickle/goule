package exec

import (
	"bytes"
	"io/ioutil"
	"os"
	execlib "os/exec"
	"testing"
	"time"
)

// A simple program that writes "hi" to a file called "test.txt".
const FILE_PROGRAM = "package main\n\nimport \"io/ioutil\"\n\n" +
	"func main() {\n" +
	"\tioutil.WriteFile(\"test.txt\", []byte(\"hi\"), 0777)\n}"

func TestBasicStart(t *testing.T) {
	dir, settings := SetupExecTest(t, FILE_PROGRAM)
	defer dir.Remove(t)

	// Run the executable and wait one second for it to stop.
	toRun := NewExec(&settings)
	toRun.Start()
	WaitStatus(t, toRun, HALTED, 1)

	// Make sure the file exists and has correct contents.
	if data, err := ioutil.ReadFile(dir.FilePath("test.txt")); err != nil {
		t.Fatal("Failed to read test.txt")
	} else if !bytes.Equal(data, []byte("hi")) {
		t.Fatal("Invalid data in test.txt")
	}
}

func TestRelaunch(t *testing.T) {
	if testing.Short() {
		t.Skip("Test takes at least one second to run.")
	}

	dir, settings := SetupExecTest(t, FILE_PROGRAM)
	defer dir.Remove(t)

	// It will relaunch after 1 second. This should give us enough time to
	// remove the file and check if it gets created again.
	settings.Relaunch = true
	settings.RelaunchInterval = 1

	toRun := NewExec(&settings)

	// Run the executable and wait for it to begin relaunching.
	toRun.Start()
	WaitStatus(t, toRun, RESTARTING, 2)

	// Make sure the file exists and has correct contents.
	if data, err := ioutil.ReadFile(dir.FilePath("test.txt")); err != nil {
		t.Fatal("Failed to read test.txt")
	} else if !bytes.Equal(data, []byte("hi")) {
		t.Fatal("Invalid data in test.txt")
	}
	dir.RemoveFile(t, "test.txt")

	// Wait until the file exists again.
	for timeout := 2000; timeout > 0; timeout-- {
		// If the file exists, the test is done!
		if data, err := ioutil.ReadFile(dir.FilePath("test.txt")); err == nil {
			if !bytes.Equal(data, []byte("hi")) {
				t.Fatal("Invalid data in test.txt")
			}
			// Done successfully! Stop the Exec since it will restart again.
			toRun.Stop()
			return
		}
		time.Sleep(time.Millisecond)
	}
	t.Fatal("Executable did not relaunch after 2 seconds.")
}

func TestStopRelaunch(t *testing.T) {
	if testing.Short() {
		t.Skip("Test takes at least two seconds to run.")
	}

	dir, settings := SetupExecTest(t, FILE_PROGRAM)
	defer dir.Remove(t)

	// It will relaunch after 1 second. This should give us enough time to
	// remove the file and check if it gets created again.
	settings.Relaunch = true
	settings.RelaunchInterval = 1

	toRun := NewExec(&settings)

	// Run the executable and wait for it to begin relaunching.
	toRun.Start()
	WaitStatus(t, toRun, RESTARTING, 2)

	// Stop the executable mid-restart
	toRun.Stop()
	if toRun.GetInfo().Status != HALTED {
		t.Fatal("Invalid status after stopping executable.")
	}

	// Make sure the file exists and has correct contents.
	if data, err := ioutil.ReadFile(dir.FilePath("test.txt")); err != nil {
		t.Fatal("Failed to read test.txt")
	} else if !bytes.Equal(data, []byte("hi")) {
		t.Fatal("Invalid data in test.txt")
	}
	dir.RemoveFile(t, "test.txt")

	// Make sure the file does not get created again
	time.Sleep(time.Second * 2)
	if _, err := ioutil.ReadFile(dir.FilePath("test.txt")); err == nil {
		t.Fatal("Executable still ran after being stopped.")
	}
}

func SetupExecTest(t *testing.T, code string) (*TestDir, Settings) {
	dir := NewTestDir(t)

	sourceFile := dir.FilePath("test.go")
	if ioutil.WriteFile(sourceFile, []byte(code), 0777) != nil {
		dir.Remove(t)
		t.Fatal("Failed to write source file")
	}

	logs := LogSettings{false, false, 0}
	goPath, err := execlib.LookPath("go")
	if err != nil {
		dir.Remove(t)
		t.Fatal("Cannot find `go` command.")
	}
	args := []string{goPath, "run", sourceFile}
	env := map[string]string{}
	settings := Settings{dir.path, "test_exc", logs, logs,
		UserIdentity{false, false, 0, 0}, args, env, false, false, 0}

	return dir, settings
}

func WaitStatus(t *testing.T, exc *Exec, status Status, timeout int) {
	threshold := timeout * 1000
	for exc.GetInfo().Status != status {
		time.Sleep(time.Millisecond)
		threshold--
		if threshold == 0 {
			t.Fatal("Did not reach executable status after timeout.")
		}
	}
}

type TestDir struct {
	path string
}

func NewTestDir(t *testing.T) *TestDir {
	if path, err := ioutil.TempDir("", "exec_test_"); err != nil {
		t.Fatal("Failed to create temporary directory")
		return nil
	} else {
		return &TestDir{path}
	}
}

func (self *TestDir) Remove(t *testing.T) {
	if os.RemoveAll(self.path) != nil {
		t.Fatal("Failed to remove temporary directory")
	}
}

func (self *TestDir) RemoveFile(t *testing.T, name string) {
	if err := os.Remove(self.FilePath(name)); err != nil {
		t.Fatal("Failed to remove file " + name + " with error " + err.Error())
	}
}

func (self *TestDir) FilePath(name string) string {
	return self.path + "/" + name
}
