package main

var storageInterfaces = map[string]func() StorageInterface{}

type StorageInterface interface {
	Open(*shortieConfiguration) error
	Close() error
	Store(string, string) error
	Fetch(string) (string, error)
}

type StorageAlias interface {
	StoreAlias(string, string) error
	StoreWithAlias(string, string, string) error
	FetchAlias(string) (string, error)
}

func RegisterStorageInterface(name string, f func() StorageInterface) {
	storageInterfaces[name] = f
}

func GetStorageInterface(name string) (storageInterface StorageInterface) {
	if storageInterfacef, ok := storageInterfaces[name]; ok {
		storageInterface = storageInterfacef()
	}
	return
}
