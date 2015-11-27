package main

import "github.com/icholy/vedis"

type vedisStorage struct {
	store *vedis.Store
}

type vedisConfiguration struct {
	Store string `toml:"store"`
}

func (v *vedisStorage) Open(c *shortieConfiguration) (err error) {
	vedisConfig := vedisConfiguration{}
	if err = c.UnifyStorageConfiguration("vedis", &vedisConfig); err != nil {
		return
	}

	if vedisConfig.Store == "" {
		vedisConfig.Store = ":mem:"
	}

	v.store, err = vedis.Open(vedisConfig.Store)
	return
}

func (v *vedisStorage) Close() error {
	v.store.Close()
	return nil
}

func (v *vedisStorage) Store(key, value string) error {
	return v.store.KvStore([]byte(key), []byte(value))
}

func (v *vedisStorage) Fetch(key string) (value string, err error) {
	var bvalue []byte
	if bvalue, err = v.store.KvFetch([]byte(key)); err == nil {
		value = string(bvalue)
	}
	return
}

func init() {
	RegisterStorageInterface("vedis", func() StorageInterface {
		return &vedisStorage{}
	})
}
