package main

import "github.com/BurntSushi/toml"

type shortieConfiguration struct {
	md                   toml.MetaData             `toml:"-"`
	Listen               string                    `toml:"listen"`
	StorageInterface     string                    `toml:"storage"`
	StorageConfiguration map[string]toml.Primitive `toml:"config"`
}

func loadConfiguration(file string) (*shortieConfiguration, error) {
	var err error
	config := shortieConfiguration{
		Listen:           ":3000",
		StorageInterface: "vedis",
	}
	config.md, err = toml.DecodeFile(file, &config)
	return &config, err
}

func (c *shortieConfiguration) UnifyStorageConfiguration(name string, v interface{}) (err error) {
	if c.md.IsDefined(name) {
		err = c.md.PrimitiveDecode(c.StorageConfiguration[name], v)
	}
	return
}
