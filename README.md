# shortie

A simple link shortener with support for several storage backends

Works out of the box with no external database or configuration however configuring it is extremely beneficial

## Supported backends

 * vedis - A networkless redis, similar to sqlite
 * redis - Traditional redis
 * mongo - You guessed it, mongo

## Quick start

 1. go get github.com/freman/shortie
 2. rm ./bin/shortie
 3. go get github.com/GeertJohan/go.rice/rice
 4. ./bin/rice --import-path=github.com/freman/shortie embed-go
 5. go install github.com/freman/shortie

That's it, the `shortie` binary in bin directory is now 100% self contained and ready to roll

## Configuration

Configuration is stored in a TOML file, out of the box it stores data with vedis in memory.

There are example configuration files in the example directory but the basic gist is.

	listen = "listen configuration"
	storage = "{drivername}"

	[config.{drivername}]
	$options for driver

## Known Issues

 * At the moment it's using uuids as a method of shortening urls, it's basically cheap and dirty, I'll get around to making a couple of other drivers for generating ids
 * I've never actually used the mongo driver ;)
 * I googled the image used as a backdrop, it was all over the place with no attribution, if it's yours, let me know if you want me to remove it or credit you.
