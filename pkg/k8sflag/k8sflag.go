package k8sflag

import (
	"io/ioutil"
	"log"
	"path/filepath"
	"strconv"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
)

type flag interface {
	refresh([]byte)
}

type ConfigMap struct {
	path    []string
	watcher *fsnotify.Watcher
	watches map[string]flag
}

func NewConfigMap(path string) *ConfigMap {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	c := &ConfigMap{
		path:    filepath.SplitList(path),
		watcher: w,
		watches: make(map[string]flag),
	}
	go func() {
		for {
			select {
			case event := <-c.watcher.Events:
				f, ok := c.watches[event.Name]
				if !ok {
					log.Printf("Event for unknown flag %v.", event.Name)
					continue
				}
				b, err := ioutil.ReadFile(event.Name)
				if err != nil {
					log.Printf("Error reading file: %v", err)
					continue
				}
				f.refresh(b)
			case err := <-c.watcher.Errors:
				log.Printf("Error event: %v", err)
			}
		}
	}()
	return c
}

func (c *ConfigMap) register(path string, f flag) {
	p := filepath.SplitList(path)
	p = append(c.path, p...)
	filename := filepath.Join(p...)
	if _, ok := c.watches[filename]; ok {
		panic("Flag already bound to " + filename)
	}
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("No configuration for %v: %v", filename, err)
	} else {
		f.refresh(b)
	}
	c.watches[filename] = f
	c.watcher.Add(filename)
}

var defaultConfigMap = NewConfigMap("")

type StringFlag struct {
	value atomic.Value
}

func (c *ConfigMap) String(path string, value string) *StringFlag {
	s := &StringFlag{}
	s.value.Store(value)
	c.register(path, flag(s))
	return s
}

func String(path, value string) *StringFlag {
	return defaultConfigMap.String(path, value)
}

func (f *StringFlag) refresh(b []byte) {
	s := string(b)
	f.value.Store(s)
	log.Printf("Set config ? to %v.", s)
}

func (f *StringFlag) Get() string {
	return f.value.Load().(string)
}

type BoolFlag struct {
	value atomic.Value
}

func (c *ConfigMap) Bool(path string, value bool) *BoolFlag {
	b := &BoolFlag{}
	b.value.Store(value)
	c.register(path, flag(b))
	return b
}

func Bool(path string, value bool) *BoolFlag {
	return defaultConfigMap.Bool(path, value)
}

func (f *BoolFlag) refresh(bytes []byte) {
	s := string(bytes)
	b, err := strconv.ParseBool(s)
	if err != nil {
		log.Printf("Error parsing bool %v: %v", s, err)
		return
	}
	f.value.Store(b)
	log.Printf("Set value to %v.", b)
}

func (f *BoolFlag) Get() bool {
	return f.value.Load().(bool)
}
