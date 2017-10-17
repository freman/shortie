package main

import (
	"strconv"

	"github.com/bwmarrin/snowflake"
)

type snowflakeID struct {
	Node     int
	Encoding string
	node     *snowflake.Node
}

func (i *snowflakeID) Setup(c *shortieConfiguration) (err error) {
	i.Encoding = "base58"
	if err = c.UnifySubConfiguration("snowflake", i); err != nil {
		return
	}
	i.node, err = snowflake.NewNode(int64(i.Node))
	return
}

func (i *snowflakeID) Get() string {
	id := i.node.Generate()
	switch i.Encoding {
	case "base2":
		return id.Base2()
	case "base36":
		return id.Base36()
	case "base58":
		return id.Base58()
	case "base64":
		return id.Base64()
	}

	return strconv.FormatInt(id.Int64(), 10)
}

func init() {
	RegisterIDInterface("snowflake", func() IDInterface {
		return &snowflakeID{}
	})
}
