package goule

import (
	"io"
	"io/ioutil"
)

// createLogStdout takes the info for an executable and returns a writer to which
// standard output should be written.
func createLogStdout(info ExecutableInfo) (io.Writer, error) {
	// TODO: here, open a file etc.
	return ioutil.Discard, nil
}

// createLogStderr takes the info for an executable and returns a writer to which
// standard error should be written.
func createLogStderr(info ExecutableInfo) (io.Writer, error) {
	// TODO: here, open a file etc.
	return ioutil.Discard, nil
}
