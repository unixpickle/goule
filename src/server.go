package goule

type Server interface {
	func Run(handler Handler, port int)
	func Stop()
}
