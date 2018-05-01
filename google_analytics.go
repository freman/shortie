package main

import (
	"net/http"

	ga "github.com/jpillora/go-ogle-analytics"
)

type googleAnalytics struct {
	c          *shortieConfiguration
	TrackingID string `toml:"tracking_id"`
}

func (m *googleAnalytics) Setup(c *shortieConfiguration) (err error) {
	if err = c.UnifySubConfiguration("google_analytics", m); err != nil {
		return
	}
	m.c = c
	return
}

func (m *googleAnalytics) Record(r *http.Request, target, short, alias string) error {
	c, err := ga.NewClient(m.TrackingID)
	if err != nil {
		return err
	}

	c.UserAgentOverride(r.Header.Get("User-Agent"))
	c.CampaignID(short)
	if alias != "" {
		c.CampaignKeyword(alias)
		c.CampaignName(alias)
	}
	c.DocumentLocationURL(target)
	ref := r.Header.Get("Referrer")
	if ref == "" {
		ref = r.Header.Get("Origin")
	}
	c.DocumentReferrer(ref)
	c.IPOverride(m.c.IP(r))

	return c.Send(ga.NewPageview())
}

func init() {
	RegisterMetricsInterface("google_analytics", func() MetricsInterface {
		return &googleAnalytics{}
	})
}
