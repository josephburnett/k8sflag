package k8sflag

import (
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
)

type flag interface {
	set([]byte)
	setDefault()
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
					log.Printf("No binding for %v.", event.Name)
					continue
				}
				b, err := ioutil.ReadFile(event.Name)
				if err != nil {
					if os.IsNotExist(err) {
						f.setDefault()
					} else {
						log.Printf("Error reading file: %v", err)
					}
					continue
				}
				f.set(b)
			case err := <-c.watcher.Errors:
				log.Printf("Error event: %v", err)
			}
		}
	}()
	c.watcher.Add(path)
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
		if os.IsNotExist(err) {
			f.setDefault()
		} else {
			log.Printf("Error reading file: %v", err)
		}
	} else {
		f.set(b)
	}
	c.watches[filename] = f
	c.watcher.Add(filename)
}

var defaultConfigMap = NewConfigMap("")

type StringFlag struct {
	value atomic.Value
	def   string
}

func (c *ConfigMap) String(path string, def string) *StringFlag {
	s := &StringFlag{
		def: def,
	}
	s.value.Store(def)
	c.register(path, flag(s))
	return s
}

func String(path, value string) *StringFlag {
	return defaultConfigMap.String(path, value)
}

func (f *StringFlag) set(b []byte) {
	s := string(b)
	f.value.Store(s)
	log.Printf("Set config to %v.", s)
}

func (f *StringFlag) setDefault() {
	f.value.Store(f.def)
	log.Printf("Set to default: %v.", f.def)
}

func (f *StringFlag) Get() string {
	return f.value.Load().(string)
}

type BoolFlag struct {
	value atomic.Value
	def   bool
}

func (c *ConfigMap) Bool(path string, def bool) *BoolFlag {
	b := &BoolFlag{
		def: def,
	}
	b.value.Store(def)
	c.register(path, flag(b))
	return b
}

func Bool(path string, value bool) *BoolFlag {
	return defaultConfigMap.Bool(path, value)
}

func (f *BoolFlag) set(bytes []byte) {
	s := string(bytes)
	b, err := strconv.ParseBool(s)
	if err != nil {
		log.Printf("Error parsing bool %v: %v", s, err)
		return
	}
	f.value.Store(b)
	log.Printf("Set value to %v.", b)
}

func (f *BoolFlag) setDefault() {
	f.value.Store(f.def)
}

func (f *BoolFlag) Get() bool {
	return f.value.Load().(bool)
}
