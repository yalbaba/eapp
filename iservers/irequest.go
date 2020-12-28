package iservers

type IRequest interface {
	GetService() string
	GetMethod() string
	GetForm() map[string]interface{}
}
