package goule

const (
	EXECUTABLE_HALTED = iota
	EXECUTABLE_RUNNING = iota
	EXECUTABLE_RESTARTING = iota
)

type ExecutableStatus int

type Executable struct {
	
}

func NewExecutable(info ExecutableInfo) *Executable {
	// TODO: create a new Executable here
	return nil
}

func (self *Executable) Start() error {
	// TODO: launch the executable if it is not already running
	return nil
}

func (self *Executable) Stop() {
	// TODO: stop the executable if it is not already running
}

func (self *Executable) GetInfo() ExecutableInfo {
	// TODO: return the info contained in the executable
	return ExecutableInfo{}
}

func (self *Executable) GetStatus() ExecutableStatus {
	// TODO: return the current status of the executable
	return 0
}
