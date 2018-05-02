package main

import (
	"net"
	"net/http"
	"strings"

	"github.com/BurntSushi/toml"
	gct "github.com/freman/go-commontypes"
)

type shortieConfiguration struct {
	md               toml.MetaData             `toml:"-"`
	RedirectCode     int                       `toml:"code"`
	DisableUI        bool                      `toml:"disable_ui"`
	DisableAlias     bool                      `toml:"disable_alias"`
	RedirectTo       string                    `tonl:"redirect_to"`
	Listen           string                    `toml:"listen"`
	IDInterface      string                    `toml:"id"`
	Metrics          string                    `toml:"metrics"`
	Filters          []string                  `toml:"filters"`
	StorageInterface string                    `toml:"storage"`
	SubConfiguration map[string]toml.Primitive `toml:"config"`
	TrustedUpstreams gct.Networks
}

func loadConfiguration(file string) (*shortieConfiguration, error) {
	var err error
	config := shortieConfiguration{
		Listen:           ":3000",
		RedirectCode:     http.StatusPermanentRedirect,
		StorageInterface: "vedis",
		IDInterface:      "snowflake",
		Metrics:          "noop_analytics",
		TrustedUpstreams: gct.PrivateNetworks,
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

func (c *shortieConfiguration) IP(r *http.Request) string {
	rip, _, _ := net.SplitHostPort(r.RemoteAddr)
	rIP := net.ParseIP(rip)

	if c.TrustedUpstreams.Contains(rIP) {
		for _, header := range []string{"x-real-ip", "x-forwarded-for"} {
			slice := strings.Split(r.Header.Get(header), ",")
			for i := len(slice) - 1; i >= 0; i-- {
				if tmp := strings.TrimSpace(slice[i]); tmp != "" {
					return tmp
				}
			}
		}
	}

	return rip
}
