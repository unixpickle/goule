package goule

import (
	"github.com/unixpickle/executor"
	"github.com/unixpickle/spinlog"
	"io"
)

// Log is a configuration object which stores logging configuration.
type Log struct {
	*spinlog.LineConfig
	BufferLines bool `json:"buffer_lines"`
	Enabled     bool `json:"enabled"`
}

// NewLog creates a disabled log with nice defaults.
func NewLog() *Log {
	res := new(Log)

	// Create a Log which is disabled but which has nice default settings.
	res.MaxLineSize = 1024
	res.MaxCount = 2
	res.MaxSize = 1048576
	res.SetPerm = false
	res.BufferLines = true

	return res
}

// Open creates a new io.WriteCloser based on the logging configuration.
// This implementation makes a Log an executor.Log.
func (l *Log) Open() (io.WriteCloser, error) {
	if !l.Enabled {
		return executor.NullLog.Open()
	} else if l.BufferLines {
		return spinlog.NewLineLog(*l.LineConfig)
	} else {
		return spinlog.NewLog(l.LineConfig.Config)
	}
}
