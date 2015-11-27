package main

import (
	"labix.org/v2/mgo"
	"labix.org/v2/mgo/bson"
)

type mongoStorage struct {
	session    *mgo.Session
	collection *mgo.Collection
}

type mongoConfiguration struct {
	Url        string `toml:"url"`
	Collection string `toml:"collection"`
}

type mongoCollectionStruct struct {
	Key   string `bson:"_id"`
	Value string `bson:"value"`
}

func (m *mongoStorage) Open(c *shortieConfiguration) (err error) {
	mongoConfig := mongoConfiguration{
		Url:        "mongodb://localhost",
		Collection: "shortie",
	}
	if err = c.UnifyStorageConfiguration("mongo", &mongoConfig); err != nil {
		return
	}

	m.session, err = mgo.Dial(mongoConfig.Url)
	if err != nil {
		return
	}

	m.collection = m.session.DB("").C(mongoConfig.Collection)
	return
}

func (m *mongoStorage) Close() error {
	m.session.Close()
	return nil
}

func (m *mongoStorage) Store(key, value string) error {
	return m.collection.Insert(&mongoCollectionStruct{Key: key, Value: value})
}

func (m *mongoStorage) Fetch(key string) (value string, err error) {
	result := mongoCollectionStruct{}
	if err := m.collection.Find(bson.M{"_id": key}).One(&result); err == nil {
		value = result.Value
	}
	return
}

func init() {
	RegisterStorageInterface("mongo", func() StorageInterface {
		return &mongoStorage{}
	})
}
