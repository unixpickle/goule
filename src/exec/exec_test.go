package exec

import (
	"bytes"
	"io/ioutil"
	"os"
	execlib "os/exec"
	"testing"
	"time"
)

func TestBasicStart(t *testing.T) {
	// A simple program that writes "hi" to a file called "test.txt".
	code := "package main\n\nimport \"io/ioutil\"\n\n" +
		"func main() {\n" +
		"\tioutil.WriteFile(\"test.txt\", []byte(\"hi\"), 0777)\n}"
	dir, settings := SetupExecTest(t, code)
	defer dir.Remove(t)

	// Run the executable and wait for it to stop.
	toRun := NewExec(&settings)
	toRun.Start()
	for toRun.GetInfo().Status != HALTED {
		time.Sleep(time.Millisecond)
	}

	// Make sure the file exists and has correct contents.
	data, err := ioutil.ReadFile(dir.FilePath("test.txt"))
	if err != nil {
		t.Fatal("Failed to read test.txt")
	}
	if !bytes.Equal(data, []byte("hi")) {
		t.Fatal("Invalid data in test.txt")
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

func (self *TestDir) FilePath(name string) string {
	return self.path + "/" + name
}
