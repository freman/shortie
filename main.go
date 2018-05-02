package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"html/template"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"os"
	"path"
	"strings"
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

	metrics := GetMetricsInterface(config.Metrics)
	if metrics == nil {
		log.Fatal("Invalid metrics specified in configuration")
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

	if err = metrics.Setup(config); err != nil {
		log.Print(err)
		log.Fatal("Unable to configure metrics interface")
	}

	filters := make([]FilterInterface, len(config.Filters))
	for i, v := range config.Filters {
		filter := GetFilterInterface(v)
		if filter == nil {
			log.Fatalf("Invalid filter '%s' specified in configuration", v)
		}
		if err := filter.Setup(config); err != nil {
			log.Print(err)
			log.Fatal("Unable to configure filter interface '%s'", v)
		}
		filters[i] = filter
	}

	e := echo.New()

	e.Use(mw.Logger())
	e.Use(mw.Recover())

	assets := rice.MustFindBox("public")
	assetHandler := http.FileServer(assets.HTTPBox())

	var indexTemplate *template.Template
	fmt.Println(assets.String("index.tmpl"))
	if ts, err := assets.String("index.tmpl"); ts != "" && err == nil {
		log.Println("Rendering index from index.tmpl")
		indexTemplate, err = template.New("index").Parse(ts)
		if err != nil {
			log.Fatal(err)
		}
	}

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

		filename := path.Base(c.Request().URL.Path)
		if (filename == "shorten" || strings.HasPrefix(filename, "index.htm")) && indexTemplate != nil {
			return indexTemplate.Execute(c.Response().Writer, config)
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
			var id string
			for id == "" {
				id = identifier.Get()
				for _, filter := range filters {
					if filter.Filter(id) {
						id = ""
						continue
					}
				}
			}

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

	e.GET("/:id", func(c echo.Context) (err error) {
		var alias string
		id := c.Param("id")

		s, isa := store.(StorageAlias)
		if isa {
			alias, err = s.FetchAlias(id)
			if err != nil {
				return err
			}
			if alias != "" {
				id = alias
			}
		}

		url, err := store.Fetch(id)

		if err != nil {
			return err
		}

		if err := metrics.Record(c.Request(), url, id, alias); err != nil {
			log.Printf("Unable to record metrics: %v", err)
		}

		return c.Redirect(http.StatusMovedPermanently, url)
	})

	e.Server.Addr = config.Listen

	err = graceful.ListenAndServe(e.Server, time.Second)
	if err != nil {
		panic(err)
	}

	if closer, isa := identifier.(io.Closer); isa {
		log.Println("Closing identifier")
		closer.Close()
	}
	log.Println("Closing store")
	store.Close()

}
