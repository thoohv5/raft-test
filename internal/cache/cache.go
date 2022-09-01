package cache

import (
	"encoding/json"
	"io"
	"sync"
)

type (
	cache struct {
		data map[string]string
		rwm  sync.RWMutex
	}
	ICache interface {
		Set(key, value string)
		Get(key string) string
		Marshal() ([]byte, error)
		UnMarshal(rc io.ReadCloser) error
	}
)

func New() ICache {
	return &cache{
		data: make(map[string]string),
	}
}

func (c *cache) Set(key, value string) {
	c.rwm.Lock()
	defer c.rwm.Unlock()
	c.data[key] = value
}

func (c *cache) Get(key string) string {
	c.rwm.RLock()
	defer c.rwm.RUnlock()
	return c.data[key]
}

func (c *cache) Marshal() ([]byte, error) {
	return json.Marshal(c.data)
}

func (c *cache) UnMarshal(rc io.ReadCloser) (err error) {
	return json.NewDecoder(rc).Decode(&c.data)
	// body, err := ioutil.ReadAll(rc)
	// if err != nil {
	// 	return
	// }
	// err = json.Unmarshal(body, &c.data)
	// if err != nil {
	// 	return
	// }
	// return
}
