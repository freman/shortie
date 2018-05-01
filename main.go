package main

import (
	"encoding/json"
	"errors"
	"flag"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/GeertJohan/go.rice"
	"github.com/labstack/echo"
	mw "github.com/labstack/echo/middleware"
	"gopkg.in/tylerb/graceful.v1"
)

type URL struct {
	url.URL
}

type shrinkRequest struct {
	Url   URL    `json:"url"`
	Alias string `json:"alias"`
}

type aliasRequest struct {
	Short string `json:"short"`
	Name  string `json:"name"`
}

type response struct {
	Url   string `json:"url"`
	Short string `json:"short"`
	Alias string `json:"alias,omitempty"`
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

	identifier := GetIDInterface(config.IDInterface)
	if identifier == nil {
		log.Fatal("Invalid identifier specified in configuration")
	}

	if err = store.Open(config); err != nil {
		if opError, ok := err.(*net.OpError); ok {
			log.Fatalf("Problem connecting to %s[%s]: %s", config.StorageInterface, opError.Addr.String(), opError.Err.Error())
		}
		log.Print(err)
		log.Fatal("Unable to configure storage interface")
	}
	defer store.Close()

	if err = identifier.Setup(config); err != nil {
		log.Print(err)
		log.Fatal("Unable to configure identifier interface")
	}

	e := echo.New()

	e.Use(mw.Logger())
	e.Use(mw.Recover())

	assetHandler := http.FileServer(rice.MustFindBox("public").HTTPBox())

	e.GET("/", func(c echo.Context) error {
		if config.DisableUI {
			if config.RedirectTo != "" {
				return c.Redirect(http.StatusFound, config.RedirectTo)
			}
			return echo.ErrNotFound
		}
		return c.Redirect(http.StatusFound, "/shorten/")
	})

	e.GET("/shorten/*", func(c echo.Context) error {
		if config.DisableUI {
			if config.RedirectTo != "" {
				return c.Redirect(http.StatusFound, config.RedirectTo)
			}
			return echo.ErrNotFound
		}

		http.StripPrefix("/shorten/", assetHandler).
			ServeHTTP(c.Response().Writer, c.Request())
		return nil
	})

	e.GET("/favicon.ico", func(c echo.Context) error {
		return c.Redirect(http.StatusMovedPermanently, "/shorten/favicon/favicon.ico")
	})

	e.POST("/shorten/shrink.json", func(c echo.Context) error {
		r := shrinkRequest{}
		if err := c.Bind(&r); err != nil {
			return err
		}

		if url := r.Url.String(); url != "" {
			id := identifier.Get()
			if r.Alias != "" {
				s, isa := store.(StorageAlias)
				if !isa {
					return errors.New("Storage interface doesn't support aliases")
				}

				if err = s.StoreWithAlias(id, url, r.Alias); err != nil {
					return err
				}

				return c.JSON(http.StatusOK, response{Url: url, Short: id, Alias: r.Alias})
			}

			if err = store.Store(id, url); err != nil {
				return err

			}

			return c.JSON(http.StatusOK, response{Url: url, Short: id})
		}

		return nil
	})

	e.POST("/shorten/alias.json", func(c echo.Context) error {
		r := aliasRequest{}
		if err := c.Bind(&r); err != nil {
			return err
		}

		s, isa := store.(StorageAlias)
		if !isa {
			return errors.New("Storage interface doesn't support aliases")
		}

		if r.Name == "" || r.Short == "" {
			return errors.New("Both an alias Name and a Short id are required")
		}

		url, err := store.Fetch(r.Short)
		if err != nil {
			return err
		}
		if url == "" {
			return errors.New("No such alias exists")
		}

		if err = s.StoreAlias(r.Name, r.Short); err != nil {
			return err
		}

		return c.JSON(http.StatusOK, response{Url: url, Short: r.Short, Alias: r.Name})
	})

	e.GET("/:id", func(c echo.Context) error {
		id := c.Param("id")

		s, isa := store.(StorageAlias)
		if isa {
			aliased, err := s.FetchAlias(id)
			if err != nil {
				return err
			}
			if aliased != "" {
				id = aliased
			}
		}

		url, err := store.Fetch(id)

		if err != nil {
			return err
		}

		return c.Redirect(http.StatusMovedPermanently, url)
	})

	e.Server.Addr = config.Listen

	err = graceful.ListenAndServe(e.Server, time.Second)
	if err != nil {
		panic(err)
	}
}
