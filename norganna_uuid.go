package main

import "github.com/NorgannasAddOns/go-uuid"

type norgannaUUID struct {
	Prefix string
}

func (i *norgannaUUID) Setup(c *shortieConfiguration) error {
	i.Prefix = "U"
	return c.UnifySubConfiguration("norganna", i)
}

func (i *norgannaUUID) Get() string {
	return uuid.New(i.Prefix)
}

func init() {
	RegisterIDInterface("norganna", func() IDInterface {
		return &norgannaUUID{}
	})
}
