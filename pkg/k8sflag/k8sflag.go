package k8sflag

import (
	"path/filepath"
	"sync/atomic"

	"github.com/fsnotify/fsnotify"
)

type flag interface {
	Load() interface{}
	Store(interface{})
}

type ConfigMap struct {
	path    []string
	watcher *fsnotify.Watcher
	flags   []flag
}

func NewConfigMap(path string) *ConfigMap {
	w, err := fsnotify.NewWatcher()
	if err != nil {
		panic(err)
	}
	return &ConfigMap{
		path:    filepath.SplitList(path),
		watcher: w,
		flags:   make([]flag, 0),
	}
}

var defaultConfigMap = NewConfigMap("")

type StringFlag atomic.Value

func (c *ConfigMap) String(path string, value string) *StringFlag {
	var v atomic.Value
	v.Store(value)
	s := StringFlag(v)
	// TOOD: add watcher for path to update String
	return &s
}

func String(path, value string) *StringFlag {
	return defaultConfigMap.String(path, value)
}

func (f *StringFlag) Get() string {
	v := atomic.Value(*f)
	if s, ok := v.Load().(string); ok {
		return s
	}
	return ""
}

type BoolFlag atomic.Value

func (c *ConfigMap) Bool(path string, value bool) *BoolFlag {
	var v atomic.Value
	v.Store(value)
	b := BoolFlag(v)
	// TODO: add watcher for path to update Bool
	return &b
}

func Bool(path string, value bool) *BoolFlag {
	return defaultConfigMap.Bool(path, value)
}

func (f *BoolFlag) Get() bool {
	v := atomic.Value(*f)
	if b, ok := v.Load().(bool); ok {
		return b
	}
	return false
}
