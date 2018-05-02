package main

import (
	"net/http"
	"time"

	"github.com/NorgannasAddOns/go-uuid"
	ga "github.com/jpillora/go-ogle-analytics"
)

type googleAnalytics struct {
	c            *shortieConfiguration
	TrackingID   string `toml:"tracking_id"`
	Prefix       string `toml:"prefix"`
	CampaignID   string `toml:"campaign_id"`
	CookieName   string `toml:"cookie_name"`
	CookiePrefix string `toml:"cookie_prefix"`
}

func (m *googleAnalytics) Setup(c *shortieConfiguration) (err error) {
	m.Prefix = "shortie_"
	m.CampaignID = "shortie"
	m.CookieName = "sga"
	m.CookiePrefix = "ga-"
	if err = c.UnifySubConfiguration("google_analytics", m); err != nil {
		return
	}
	m.c = c
	return
}

func (m *googleAnalytics) Record(r *http.Request, target, short, alias string) (*http.Cookie, error) {
	c, err := ga.NewClient(m.TrackingID)
	if err != nil {
		return nil, err
	}

	cookie, _ := r.Cookie(m.CookieName)
	if cookie == nil {
		cookie = &http.Cookie{
			Name:  m.CookieName,
			Value: m.CookiePrefix + uuid.New("g"),
		}
	}
	cookie.Expires = time.Now().Add(time.Hour * 17520)

	c.ClientID(cookie.Value)

	c.UserAgentOverride(r.Header.Get("User-Agent"))
	c.CampaignID(m.CampaignID)
	c.CampaignName(m.Prefix + short)

	if alias != "" {
		c.CampaignKeyword(m.Prefix + alias)
	}

	c.DocumentLocationURL(target)
	ref := r.Header.Get("Referrer")
	if ref == "" {
		ref = r.Header.Get("Origin")
	}
	c.DocumentReferrer(ref)
	c.IPOverride(m.c.IP(r))

	return cookie, c.Send(ga.NewPageview())
}

func init() {
	RegisterMetricsInterface("google_analytics", func() MetricsInterface {
		return &googleAnalytics{}
	})
}
