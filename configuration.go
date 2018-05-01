package main

import "github.com/BurntSushi/toml"

type shortieConfiguration struct {
	md               toml.MetaData             `toml:"-"`
	DisableUI        bool                      `toml:"disable_ui"`
	RedirectTo       string                    `tonl:"redirect_to"`
	Listen           string                    `toml:"listen"`
	IDInterface      string                    `toml:"id"`
	StorageInterface string                    `toml:"storage"`
	SubConfiguration map[string]toml.Primitive `toml:"config"`
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

func (c *shortieConfiguration) UnifySubConfiguration(name string, v interface{}) (err error) {
	if c.md.IsDefined("config", name) {
		err = c.md.PrimitiveDecode(c.SubConfiguration[name], v)
	}
	return
}
