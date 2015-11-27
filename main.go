package main

import (
	"encoding/json"
	"flag"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/GeertJohan/go.rice"
	"github.com/NorgannasAddOns/go-uuid"
	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"gopkg.in/tylerb/graceful.v1"
)

type URL struct {
	url.URL
}

type request struct {
	Url URL `json:"url"`
}

type response struct {
	Url   string `json:"url"`
	Short string `json:"short"`
}

func (u *URL) UnmarshalJSON(data []byte) error {
	var rawUrl string
	if err := json.Unmarshal(data, &rawUrl); err != nil {
		return err
	}
	uri, err := url.Parse(rawUrl)

	if uri.Scheme == "" {
		uri.Scheme = "http"
	}

	u.URL = *uri

	return err
}

func main() {
	configFile := flag.String("config", "config.toml", "Path to configuration")
	flag.Parse()

	config, err := loadConfiguration(*configFile)

	if err != nil && !(*configFile == "config.toml" && os.IsNotExist(err)) {
		log.Fatalf("Could not load config file '%s': %s", *configFile, err)
	}

	store := GetStorageInterface(config.StorageInterface)
	if store == nil {
		log.Fatal("Invalid store specified in configuration")
	}

	if err = store.Open(config); err != nil {
		if opError, ok := err.(*net.OpError); ok {
			log.Fatalf("Problem connecting to %s[%s]: %s", config.StorageInterface, opError.Addr.String(), opError.Err.Error())
		}
		log.Fatal(err)
	}
	defer store.Close()

	e := echo.New()

	e.Use(mw.Logger())
	e.Use(mw.Recover())

	assetHandler := http.FileServer(rice.MustFindBox("public").HTTPBox())

	e.Get("/", func(c *echo.Context) error {
		return c.Redirect(http.StatusFound, "/shorten/")
	})

	e.Get("/shorten/*", func(c *echo.Context) error {
		http.StripPrefix("/shorten/", assetHandler).
			ServeHTTP(c.Response().Writer(), c.Request())
		return nil
	})

	e.Get("/favicon.ico", func(c *echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/shorten/favicon/favicon.ico")
	})

	e.Post("/shorten/shrink.json", func(c *echo.Context) error {
		r := request{}
		if err := c.Bind(&r); err != nil {
			return err
		}

		if url := r.Url.String(); url != "" {
			id := uuid.New("U")
			if err = store.Store(id, url); err != nil {
				return err
			}
			return c.JSON(http.StatusOK, response{Url: url, Short: id})
		}

		return nil
	})

	e.Get("/:id", func(c *echo.Context) error {
		url, err := store.Fetch(c.P(0))

		if err != nil {
			return err
		}

		return c.Redirect(http.StatusMovedPermanently, url)
	})

	graceful.ListenAndServe(e.Server(config.Listen), time.Second)
}
