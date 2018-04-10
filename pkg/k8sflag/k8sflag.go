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

var Verbose = false

func verbose(msg string, params ...interface{}) {
	if Verbose {
		log.Printf(msg, params...)
	}
}

func info(msg string, params ...interface{}) {
	log.Printf(msg, params...)
}

type flag interface {
	set([]byte)
	setDefault()
}

type FlagSet struct {
	path    []string
	watcher *fsnotify.Watcher
	watches map[string]flag
}

func NewFlagSet(path string) *FlagSet {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	c := &FlagSet{
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
					verbose("No binding for %v.", event.Name)
					continue
				}
				b, err := ioutil.ReadFile(event.Name)
				if err != nil {
					if os.IsNotExist(err) {
						f.setDefault()
					} else {
						verbose("Error reading file: %v", err)
					}
					continue
				}
				f.set(b)
			case err := <-c.watcher.Errors:
				verbose("Error event: %v", err)
			}
		}
	}()
	c.watcher.Add(path) // Watch for new files
	return c
}

func (c *FlagSet) register(key string, f flag) {
	filename := filepath.Join(append(c.path, filepath.SplitList(key)...)...)
	if _, ok := c.watches[filename]; ok {
		panic("Flag already bound to " + key)
	}
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		if os.IsNotExist(err) {
			f.setDefault()
		} else {
			verbose("Error reading file: %v", err)
		}
	} else {
		f.set(b)
	}
	c.watches[filename] = f
	c.watcher.Add(filename)
}

var defaultFlagSet = NewFlagSet("/etc/config")

type StringFlag struct {
	key   string
	value atomic.Value
	def   string
}

type BoolFlag struct {
	key   string
	value atomic.Value
	def   bool
}

type IntFlag struct {
	key   string
	value atomic.Value
	def   int
}

func (c *FlagSet) String(key string, def string) *StringFlag {
	s := &StringFlag{
		key: key,
		def: def,
	}
	s.value.Store(def)
	c.register(key, flag(s))
	return s
}

func (c *FlagSet) Bool(key string, def bool) *BoolFlag {
	b := &BoolFlag{
		key: key,
		def: def,
	}
	b.value.Store(def)
	c.register(key, flag(b))
	return b
}

func (c *FlagSet) Int(key string, def int) *IntFlag {
	i := &IntFlag{
		key: key,
		def: def,
	}
	i.value.Store(def)
	c.register(key, flag(i))
	return i
}

func String(key, def string) *StringFlag {
	return defaultFlagSet.String(key, def)
}

func Bool(key string, def bool) *BoolFlag {
	return defaultFlagSet.Bool(key, def)
}

func Int(key string, def int) *IntFlag {
	return defaultFlagSet.Int(key, def)
}

func (f *StringFlag) set(b []byte) {
	s := string(b)
	f.value.Store(s)
	info("Set StringFlag %v: %v.", f.key, s)
}

func (f *BoolFlag) set(bytes []byte) {
	s := string(bytes)
	b, err := strconv.ParseBool(s)
	if err != nil {
		verbose("Error parsing BoolFlag %v: %v", f.key, err)
		return
	}
	f.value.Store(b)
	info("Set BoolFlag %v: %v.", f.key, b)
}

func (f *IntFlag) set(bytes []byte) {
	s := string(bytes)
	i, err := strconv.Atoi(s)
	if err != nil {
		verbose("Error parsing InfFlag %v: %v.", f.key, err)
		return
	}
	f.value.Store(i)
	info("Set IntFlag %v: %v.", f.key, i)
}

func (f *StringFlag) setDefault() {
	f.value.Store(f.def)
	info("Set StringFlag %v to default: %v.", f.key, f.def)
}

func (f *BoolFlag) setDefault() {
	f.value.Store(f.def)
	info("Set BoolFlag %v to default: %v.", f.key, f.def)
}

func (f *IntFlag) setDefault() {
	f.value.Store(f.def)
	info("Set IntFlag %v to default: %v.", f.key, f.def)
}

func (f *StringFlag) Get() string {
	return f.value.Load().(string)
}

func (f *BoolFlag) Get() bool {
	return f.value.Load().(bool)
}

func (f *IntFlag) Get() int {
	return f.value.Load().(int)
}
