package k8sflag

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFlagWithDefault(t *testing.T) {
	config, dir := tempConfigMap()
	defer os.RemoveAll(dir)

	flag := config.String("name", "nobody")

	name := flag.Get()
	if name != "nobody" {
		t.Fatalf("Incorrect name. Wanted nobody. Got %v.", name)
	}
}

func TestFlagWithConfig(t *testing.T) {
	config, dir := tempConfigMap()
	defer os.RemoveAll(dir)
	writeConfig(dir, "name", "joe")

	flag := config.String("name", "nobody")

	name := flag.Get()
	if name != "joe" {
		t.Fatalf("Incorrect name. Wanted joe. Got %v.", name)
	}
}

func TestFlagConfigCreate(t *testing.T) {
	config, dir := tempConfigMap()
	defer os.RemoveAll(dir)

	flag := config.String("name", "nobody")

	name := flag.Get()
	if name != "nobody" {
		t.Fatalf("Incorrect initial name. Wanted nobody. Got %v.", name)
	}

	writeConfig(dir, "name", "joe")
	time.Sleep(10 * time.Millisecond)

	name = flag.Get()
	if name != "joe" {
		t.Fatalf("Incorrect updated name. Wanted joe. Got %v.", name)
	}
}

func TestFlagConfigChange(t *testing.T) {
	config, dir := tempConfigMap()
	defer os.RemoveAll(dir)
	writeConfig(dir, "name", "sally")

	flag := config.String("name", "nobody")

	name := flag.Get()
	if name != "sally" {
		t.Fatalf("Incorrect initial name. Wanted sally. Got %v.", name)
	}

	writeConfig(dir, "name", "joe")
	time.Sleep(10 * time.Millisecond)

	name = flag.Get()
	if name != "joe" {
		t.Fatalf("Incorrect updated name. Wanted joe. Got %v.", name)
	}
}

func TestFlagConfigRemove(t *testing.T) {
	config, dir := tempConfigMap()
	defer os.RemoveAll(dir)
	writeConfig(dir, "name", "joe")

	flag := config.String("name", "nobody")

	name := flag.Get()
	if name != "joe" {
		t.Fatalf("Incorrect initial name. Wanted joe. Got %v.", name)
	}

	removeConfig(dir, "name")
	time.Sleep(10 * time.Millisecond)

	name = flag.Get()
	if name != "nobody" {
		t.Fatalf("Incorrect defaulted name. Wanted nobody. Got %v.", name)
	}
}

func TestStringEmpty(t *testing.T) {
	config, dir := tempConfigMap()
	defer os.RemoveAll(dir)

	flag := config.String("name", "")

	name := flag.Get()
	if name != "" {
		t.Fatalf("Incorrect name. Wanted empty string. Got %v.", name)
	}
}

func TestBoolTrue(t *testing.T) {
	config, dir := tempConfigMap()
	defer os.RemoveAll(dir)
	writeConfig(dir, "should", "true")

	flag := config.Bool("should", false)

	should := flag.Get()
	if should != true {
		t.Fatalf("Incorrect should. Wanted true. Got %v.", should)
	}
}

func TestBoolFalse(t *testing.T) {
	config, dir := tempConfigMap()
	defer os.RemoveAll(dir)
	writeConfig(dir, "should", "false")

	flag := config.Bool("should", true)

	should := flag.Get()
	if should != false {
		t.Fatalf("Incorrect should. Wanted false. Got %v.", should)
	}
}

func TestBoolEmpty(t *testing.T) {
	config, dir := tempConfigMap()
	defer os.RemoveAll(dir)
	writeConfig(dir, "should", "")

	flag := config.Bool("should", true)

	should := flag.Get()
	if should != true {
		t.Fatalf("Incorrect should. Wanted true. Got %v.", should)
	}
}

func tempConfigMap() (*ConfigMap, string) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}
	return NewConfigMap(dir), dir
}

func writeConfig(dir, key, value string) {
	filename := filepath.Join(dir, key)
	if err := ioutil.WriteFile(filename, []byte(value), 0666); err != nil {
		panic(err)
	}
}

func removeConfig(dir, key string) {
	os.Remove(filepath.Join(dir, key))
}