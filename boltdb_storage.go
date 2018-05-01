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

func (b *boltdbStorage) storein(bucket, key, value string) error {
	return b.store.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(bucket))
		if err != nil {
			return err
		}

		return bucket.Put([]byte(key), []byte(value))
	})
}

func (b *boltdbStorage) fetchfrom(bucket, key string) (value string, err error) {
	err = b.store.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(bucket))
		if bucket == nil {
			// Nothing has been written yet
			return nil
		}

		value = string(bucket.Get([]byte(key)))
		return nil
	})
	return
}

func (b *boltdbStorage) Store(key, value string) error {
	return b.storein("shortie", key, value)
}

func (b *boltdbStorage) Fetch(key string) (string, error) {
	return b.fetchfrom("shortie", key)
}

func (b *boltdbStorage) StoreAlias(key, value string) error {
	return b.storein("alias", key, value)
}

func (b *boltdbStorage) StoreWithAlias(key, value, alias string) error {
	return b.store.Update(func(tx *bolt.Tx) error {
		shortie, err := tx.CreateBucketIfNotExists([]byte("shortie"))
		if err != nil {
			return err
		}
		aliases, err := tx.CreateBucketIfNotExists([]byte("alias"))
		if err != nil {
			return err
		}

		if err := shortie.Put([]byte(key), []byte(value)); err != nil {
			return err
		}

		return aliases.Put([]byte(alias), []byte(key))
	})
}

func (b *boltdbStorage) FetchAlias(key string) (string, error) {
	return b.fetchfrom("alias", key)
}

func init() {
	RegisterStorageInterface("boltdb", func() StorageInterface {
		return &boltdbStorage{}
	})
}
