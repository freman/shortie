package main

var filterInterfaces = map[string]func() FilterInterface{}

type FilterInterface interface {
	Setup(*shortieConfiguration) error
	Filter(id string) (found bool)
}

func RegisterFilterInterface(name string, f func() FilterInterface) {
	filterInterfaces[name] = f
}

func GetFilterInterface(name string) (filterInterface FilterInterface) {
	if filterInterfacef, ok := filterInterfaces[name]; ok {
		filterInterface = filterInterfacef()
	}
	return
}
