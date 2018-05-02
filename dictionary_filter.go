package main

import (
	"bufio"
	"os"
	"regexp"
	"strconv"
	"strings"
)

var emptyStruct = struct{}{}

type dictionaryFilter struct {
	DictFile   string `toml:"file"`
	MaxLength  int    `toml:"max_length"`
	dictionary map[string]struct{}
}

func (m *dictionaryFilter) Setup(c *shortieConfiguration) (err error) {
	m.DictFile = "/usr/share/dict/words"
	m.MaxLength = 12
	m.dictionary = make(map[string]struct{})

	if err = c.UnifySubConfiguration("dictionary_exclude", m); err != nil {
		return
	}

	f, err := os.Open(m.DictFile)
	if err != nil {
		return err
	}
	defer f.Close()

	filterExpr := regexp.MustCompile(`^\w{1,` + strconv.Itoa(m.MaxLength) + `}$`)

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		str := strings.ToLower(strings.TrimSpace(scanner.Text()))
		if filterExpr.MatchString(str) {
			m.dictionary[str] = emptyStruct
		}
	}
	err = scanner.Err()

	return
}

func (m *dictionaryFilter) Filter(id string) (found bool) {
	_, found = m.dictionary[strings.ToLower(id)]
	return
}

func init() {
	RegisterFilterInterface("dictionary_exclude", func() FilterInterface {
		return &dictionaryFilter{}
	})
}
