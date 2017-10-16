package main

var idInterfaces = map[string]func() IDInterface{}

type IDInterface interface {
	Setup(*shortieConfiguration) error
	Get() string
}

func RegisterIDInterface(name string, f func() IDInterface) {
	idInterfaces[name] = f
}

func GetIDInterface(name string) (idInterface IDInterface) {
	if idInterfacef, ok := idInterfaces[name]; ok {
		idInterface = idInterfacef()
	}
	return
}
