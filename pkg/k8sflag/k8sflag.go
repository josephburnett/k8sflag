package k8sflag

import (
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
)

type Option int

const (
	Verbose  Option = iota
	Required Option = iota
)

type flag interface {
	name() string
	set([]byte) error
	setDefault()
	isRequired() bool
}

type FlagSet struct {
	path    []string
	watcher *fsnotify.Watcher
	watches map[string]flag
	verbose bool
}

var defaultFlagSet = NewFlagSet("/etc/config")

func NewFlagSet(path string, options ...Option) *FlagSet {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	c := &FlagSet{
		path:    filepath.SplitList(path),
		watcher: w,
		watches: make(map[string]flag),
	}
	c.watcher.Add(path) // Watch for new files
	go func() {
		defer w.Close()
		for {
			select {
			case event := <-c.watcher.Events:
				f, ok := c.watches[event.Name]
				if !ok {
					c.verboseLog("No binding for %v.", event.Name)
					continue
				}
				c.setFromFile(f, event.Name)
			case err := <-c.watcher.Errors:
				c.verboseLog("Error event: %v", err)
			}
		}
	}()
	return c
}

func (c *FlagSet) register(key string, f flag) {
	filename := filepath.Join(append(c.path, filepath.SplitList(key)...)...)
	if _, ok := c.watches[filename]; ok {
		panic("Flag already bound to " + key)
	}
	c.setFromFile(f, filename)
	c.watches[filename] = f
	c.watcher.Add(filename)
}

func (c *FlagSet) setFromFile(f flag, filename string) {
	b, err := ioutil.ReadFile(filename)
	if err != nil {
		if f.isRequired() {
			panic(fmt.Sprintf("Flag %v is required.", f.name()))
		}
		if !os.IsNotExist(err) {
			c.verboseLog("Error reading file: %v", err)
		}
		f.setDefault()
		return
	}
	err = f.set(b)
	if err != nil {
		if f.isRequired() {
			panic(fmt.Sprintf("Error reading %v: %v", f.name(), err))
		}
		f.setDefault()
	}
}

func (c *FlagSet) verboseLog(msg string, params ...interface{}) {
	if c.verbose {
		log.Printf(msg, params...)
	}
}

func info(msg string, params ...interface{}) {
	log.Printf(msg, params...)
}

type flagCommon struct {
	key      string
	value    atomic.Value
	required bool
	verbose  bool
}

func (f *flagCommon) name() string {
	return f.key
}

func (f *flagCommon) isRequired() bool {
	return f.required
}

func (f *flagCommon) verboseLog(msg string, params ...interface{}) {
	if f.verbose {
		log.Printf(msg, params...)
	}
}

type StringFlag struct {
	flagCommon
	def string
}

type BoolFlag struct {
	flagCommon
	def bool
}

type IntFlag struct {
	flagCommon
	def int
}

func (c *FlagSet) String(key string, def string, options ...Option) *StringFlag {
	s := &StringFlag{}
	s.key = key
	s.verbose = c.verbose
	s.def = def
	if hasOption(Required, options) {
		s.required = true
	}
	c.register(key, flag(s))
	return s
}

func (c *FlagSet) Bool(key string, def bool, options ...Option) *BoolFlag {
	b := &BoolFlag{}
	b.key = key
	b.def = def
	b.verbose = c.verbose
	if hasOption(Required, options) {
		b.required = true
	}
	c.register(key, flag(b))
	return b
}

func (c *FlagSet) Int(key string, def int, options ...Option) *IntFlag {
	i := &IntFlag{}
	i.key = key
	i.def = def
	i.verbose = c.verbose
	if hasOption(Required, options) {
		i.required = true
	}
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

func (f *StringFlag) set(b []byte) error {
	s := string(b)
	f.value.Store(s)
	info("Set StringFlag %v: %v.", f.key, s)
	return nil
}

func (f *BoolFlag) set(bytes []byte) error {
	s := string(bytes)
	b, err := strconv.ParseBool(s)
	if err != nil {
		return err
	}
	f.value.Store(b)
	info("Set BoolFlag %v: %v.", f.key, b)
	return nil
}

func (f *IntFlag) set(bytes []byte) error {
	s := string(bytes)
	i, err := strconv.Atoi(s)
	if err != nil {
		return err
	}
	f.value.Store(i)
	info("Set IntFlag %v: %v.", f.key, i)
	return nil
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

func hasOption(option Option, options []Option) bool {
	for _, o := range options {
		if o == option {
			return true
		}
	}
	return false
}
