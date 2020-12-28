package iservers

type IServer interface {
	Start() error
	Run() error
	Stop()
	RegistService(h IHandler)
}
