package iservers

type IRequestService interface {
	GetService() string
	GetMethod() string
	GetForm() map[string]interface{}
}
