package updateTracker

import (
	"encoding/json"
	"io"
	"os"
)

type (
	Cache struct {
		file string

		Data map[string]*CacheData
	}

	CacheData struct {
		Version int
	}
)

var cache *Cache

func newCache(file string) (c *Cache, err error) {
	_, err = os.Stat(file)
	exists := err == nil
	var f *os.File
	if exists {
		f, err = os.Open(file)
	} else {
		f, err = os.Create(file)
	}
	if err != nil {
		return
	}
	c = &Cache{}
	var data map[string]*CacheData
	if exists {
		err = json.NewDecoder(f).Decode(&data)
		if err == io.EOF {
			err = nil
			data = map[string]*CacheData{}
		}
	} else {
		data = map[string]*CacheData{}
	}
	c.file = file
	c.Data = data
	f.Close()
	return
}

func (c *Cache) Persist() (err error) {
	f, err := os.OpenFile(c.file, os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(c.Data)
}
