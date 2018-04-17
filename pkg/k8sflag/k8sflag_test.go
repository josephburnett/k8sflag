package k8sflag

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestFlagWithDefault(t *testing.T) {
	config, dir := tempFlagSet()
	defer os.RemoveAll(dir)

	flag := config.String("name", "nobody")

	name := flag.Get()
	if name != "nobody" {
		t.Fatalf("Incorrect name. Wanted nobody. Got %v.", name)
	}
}

func TestFlagWithConfig(t *testing.T) {
	config, dir := tempFlagSet()
	defer os.RemoveAll(dir)
	writeConfig(dir, "name", "joe")

	flag := config.String("name", "nobody")

	name := flag.Get()
	if name != "joe" {
		t.Fatalf("Incorrect name. Wanted joe. Got %v.", name)
	}
}

func TestFlagConfigCreate(t *testing.T) {
	config, dir := tempFlagSet()
	defer os.RemoveAll(dir)

	flag := config.String("name", "nobody", Dynamic)

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
	config, dir := tempFlagSet()
	defer os.RemoveAll(dir)
	writeConfig(dir, "name", "sally")

	flag := config.String("name", "nobody", Dynamic)

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
	config, dir := tempFlagSet()
	defer os.RemoveAll(dir)
	writeConfig(dir, "name", "joe")

	flag := config.String("name", "nobody", Dynamic)

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

func TestFlagNotDynamic(t *testing.T) {
	config, dir := tempFlagSet()
	defer os.RemoveAll(dir)

	flag := config.String("name", "nobody") // not Dynamic
	writeConfig(dir, "name", "joe")

	name := flag.Get()
	if name != "nobody" {
		t.Fatalf("Incorrect name. Wanted nobody. Got %v.", name)
	}
}

func TestStringEmpty(t *testing.T) {
	config, dir := tempFlagSet()
	defer os.RemoveAll(dir)

	flag := config.String("name", "")

	name := flag.Get()
	if name != "" {
		t.Fatalf("Incorrect name. Wanted empty string. Got %v.", name)
	}
}

func TestBoolTrue(t *testing.T) {
	config, dir := tempFlagSet()
	defer os.RemoveAll(dir)
	writeConfig(dir, "should", "true")

	flag := config.Bool("should", false)

	should := flag.Get()
	if should != true {
		t.Fatalf("Incorrect should. Wanted true. Got %v.", should)
	}
}

func TestBoolFalse(t *testing.T) {
	config, dir := tempFlagSet()
	defer os.RemoveAll(dir)
	writeConfig(dir, "should", "false")

	flag := config.Bool("should", true)

	should := flag.Get()
	if should != false {
		t.Fatalf("Incorrect should. Wanted false. Got %v.", should)
	}
}

func TestBoolEmpty(t *testing.T) {
	config, dir := tempFlagSet()
	defer os.RemoveAll(dir)
	writeConfig(dir, "should", "")

	flag := config.Bool("should", true)

	should := flag.Get()
	if should != true {
		t.Fatalf("Incorrect should. Wanted true. Got %v.", should)
	}
}

func TestIntValid(t *testing.T) {
	config, dir := tempFlagSet()
	defer os.RemoveAll(dir)
	writeConfig(dir, "count", "1")

	flag := config.Int32("count", 0)

	count := flag.Get()
	if count != 1 {
		t.Fatalf("Incorrect count. Wanted 1. Got %v.", count)
	}
}

func TestIntInvalid(t *testing.T) {
	config, dir := tempFlagSet()
	defer os.RemoveAll(dir)
	writeConfig(dir, "count", "wrong")

	flag := config.Int32("count", 0)

	count := flag.Get()
	if count != 0 {
		t.Fatalf("Incorrect count. Wanted 0. Got %v.", count)
	}
}

func TestDurationEmpty(t *testing.T) {
	config, dir := tempFlagSet()
	defer os.RemoveAll(dir)
	writeConfig(dir, "duration", "")
	want := time.Second

	flag := config.Duration("duration", &want)

	got := flag.Get()
	if *got != want {
		t.Fatalf("Incorrect duration. Wanted %v. Got %v.", want, got)
	}
}

func TestDurationMinutes(t *testing.T) {
	config, dir := tempFlagSet()
	defer os.RemoveAll(dir)
	writeConfig(dir, "duration", "5m")
	want := 5 * time.Minute

	flag := config.Duration("duration", nil)

	got := flag.Get()
	if *got != want {
		t.Fatalf("Incorrect duration. Wanted %v. Got %v.", want, got)
	}
}

func TestDurationNil(t *testing.T) {
	config, dir := tempFlagSet()
	defer os.RemoveAll(dir)

	flag := config.Duration("duration", nil)

	got := flag.Get()
	if got != nil {
		t.Fatalf("Incorrect duration. Wanted nil. Got %v.", got)
	}
}

func TestStringRequired(t *testing.T) {
	config, dir := tempFlagSet()
	defer os.RemoveAll(dir)

	defer func() {
		if r := recover(); r != nil {
			// expected
		} else {
			t.Fatalf("Expected panic. Did not panic.")
		}
	}()

	config.String("required", "", Required).Get()
}

func TestBoolRequired(t *testing.T) {
	config, dir := tempFlagSet()
	defer os.RemoveAll(dir)

	defer func() {
		if r := recover(); r != nil {
			// expected
		} else {
			t.Fatalf("Expected panic. Did not panic.")
		}
	}()

	config.Bool("required", false, Required).Get()
}

func TestIntRequired(t *testing.T) {
	config, dir := tempFlagSet()
	defer os.RemoveAll(dir)

	defer func() {
		if r := recover(); r != nil {
			// expected
		} else {
			t.Fatalf("Expected panic. Did not panic.")
		}
	}()

	config.Int32("required", 0, Required).Get()
}

func TestDurationRequired(t *testing.T) {
	config, dir := tempFlagSet()
	defer os.RemoveAll(dir)

	defer func() {
		if r := recover(); r != nil {
			// expected
		} else {
			t.Fatalf("Expected panic. Did not panic.")
		}
	}()

	config.Duration("required", nil, Required).Get()
}

func tempFlagSet() (*FlagSet, string) {
	dir, err := ioutil.TempDir("", "")
	if err != nil {
		panic(err)
	}
	return NewFlagSet(dir), dir
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
