package main

import (
	"encoding/binary"
	"os"
	"time"

	gct "github.com/freman/go-commontypes"
)

const defaultCharacterMap = "WAQNGLXCeaw3bHS18EpYvocJki2xh5PB9fj7FsMdTuZr6DmnVy4UKqtRzg"

// Hacky little interface to increment sequentially, because reasons.
type incrementingID struct {
	current      uint32
	dirty        bool
	ch           chan uint32
	StateFile    string       `toml:"state_file"`
	SaveInterval gct.Duration `toml:"save_interval"`
	Skip         int          `toml:"skip"`
	CharacterMap string       `toml:"character_map"`
}

func (i *incrementingID) base58(f uint32) string {
	l := uint32(len(i.CharacterMap))
	if f < l {
		return string(i.CharacterMap[f])
	}

	b := make([]byte, 0, 11)
	for f >= l {
		b = append(b, i.CharacterMap[f%l])
		f /= l
	}
	b = append(b, i.CharacterMap[f])

	for x, y := 0, len(b)-1; x < y; x, y = x+1, y-1 {
		b[x], b[y] = b[y], b[x]
	}

	return string(b)
}

func (i *incrementingID) save() {
	f, err := os.OpenFile(i.StateFile, os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0600)
	if err != nil {
		panic(err)
	}
	defer f.Close()
	binary.Write(f, binary.BigEndian, i.current)
	i.dirty = false
}

func (i *incrementingID) load() {
	f, err := os.Open(i.StateFile)
	if err != nil {
		if !os.IsNotExist(err) {
			panic(err)
		}
		return
	}
	defer f.Close()
	binary.Read(f, binary.BigEndian, &i.current)
	i.dirty = false
}

func (i *incrementingID) run() {
	tmr := time.NewTimer(i.SaveInterval.Duration)
	for {
		select {
		case <-tmr.C:
			if i.dirty {
				i.save()
			}
		case i.ch <- i.current:
			i.current++
			i.dirty = true
		}
		tmr.Reset(i.SaveInterval.Duration)
	}
}

func (i *incrementingID) Setup(c *shortieConfiguration) (err error) {
	i.StateFile = ".increment"
	i.SaveInterval.Duration = 5 * time.Second
	i.Skip = 10
	i.CharacterMap = defaultCharacterMap
	if err = c.UnifySubConfiguration("incrementing", i); err != nil {
		return
	}

	i.load()
	i.ch = make(chan uint32, i.Skip)
	for c := 0; c < i.Skip; c++ {
		i.ch <- i.current
		i.current++
	}

	go i.run()

	return
}

func (i *incrementingID) Close() error {
	i.save()
	return nil
}

func (i *incrementingID) Get() string {
	return i.base58(<-i.ch)
}

func init() {
	RegisterIDInterface("incrementing", func() IDInterface {
		return &incrementingID{}
	})
}
