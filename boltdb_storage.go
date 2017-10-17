package main

import (
	"github.com/boltdb/bolt"
)

type boltdbStorage struct {
	store *bolt.DB
}

type boltdbConfiguration struct {
	Store string `toml:"store"`
}

func (b *boltdbStorage) Open(c *shortieConfiguration) (err error) {
	boltdbConfig := boltdbConfiguration{
		Store: "shortie.db",
	}
	if err = c.UnifySubConfiguration("boltdb", &boltdbConfig); err != nil {
		return
	}

	b.store, err = bolt.Open(boltdbConfig.Store, 0644, bolt.DefaultOptions)
	return
}

func (b *boltdbStorage) Close() error {
	b.store.Close()
	return nil
}

func (b *boltdbStorage) Store(key, value string) error {
	return b.store.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte("shortie"))
		if err != nil {
			return err
		}

		return bucket.Put([]byte(key), []byte(value))
	})
}

func (b *boltdbStorage) Fetch(key string) (value string, err error) {
	err = b.store.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte("shortie"))
		if bucket == nil {
			// Nothing has been written yet
			return nil
		}

		value = string(bucket.Get([]byte(key)))
		return nil
	})
	return
}

func init() {
	RegisterStorageInterface("boltdb", func() StorageInterface {
		return &boltdbStorage{}
	})
}
