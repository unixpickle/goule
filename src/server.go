package goule

type Server interface {
	Run(port int) error
	Stop() error
	IsRunning() bool
}
