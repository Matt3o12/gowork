package gowork

import (
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/op/go-logging"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(m *testing.M) {
	backend := logging.InitForTesting(logging.DEBUG)
	logging.SetBackend(backend)
	code := m.Run()
	logging.Reset()
	os.Exit(code)
}

func patchEnv(key, value string) func() {
	bck := os.Getenv(key)
	deferFunc := func() {
		os.Setenv(key, bck)
	}

	os.Setenv(key, value)

	return deferFunc
}

func makeProjectTree(t *testing.T) (string, func()) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err)

	projects := []string{
		path.Join("github.com", "matt3o12", "gowork"),
		path.Join("github.com", "matt3o12", "termui-widgets"),
		path.Join("github.com", "stretchr", "testify"),
		path.Join("code.google.com", "p", "cascadia"),
		path.Join(".hidden"),
	}

	for _, theProject := range projects {
		os.MkdirAll(path.Join(dir, "src", theProject), 0777)
	}

	files := []string{
		"test.go",
	}

	for _, theFile := range files {
		os.Create(path.Join(dir, "src", theFile))
	}

	return dir, func() {
		os.RemoveAll(dir)
	}
}

func TestMakeProjectTree(t *testing.T) {
	baseDir, removeAll := makeProjectTree(t)

	getDir := func(paths ...string) string {
		paths = append([]string{baseDir, "src"}, paths...)
		return path.Join(paths...)
	}

	assertDir := func(paths ...string) {
		name := getDir(paths...)
		if stat, err := os.Stat(name); err != nil {
			t.Errorf("Directory: %v does not exist. Expected to exist.", name)
		} else if !stat.IsDir() {
			t.Errorf("Expected %v to be a directory.", name)
		}
	}

	assertFile := func(paths ...string) {
		name := getDir(paths...)
		if stat, err := os.Stat(name); err != nil {
			t.Errorf("File: %v does not exist. Expected to exist.", name)
		} else if stat.IsDir() {
			t.Errorf("Expected %v to be a file.", name)
		}
	}

	assertDir("github.com", "matt3o12", "gowork")
	assertDir("github.com", "matt3o12", "termui-widgets")
	assertDir("github.com", "stretchr", "testify")
	assertDir("code.google.com", "p", "cascadia")
	assertDir(".hidden")

	assertFile("test.go")
	removeAll()

	if _, err := os.Stat(baseDir); !os.IsNotExist(err) {
		t.Error("Temp dir was not removed.")
	}
}

func TestDistributor(t *testing.T) {
	defer patchEnv("GOPATH", "/foo/bar")()

	distro := Distributor("github.com")
	assert.Equal(t, "/foo/bar/src/github.com", distro.AbsPath())
}

func TestAllDistributors(t *testing.T) {
	defer patchEnv("GOPATH", "/foo/bar")()
	distros, err := AllDistributors()
	assert.True(t, os.IsNotExist(err), "Expected gopath not to exist")
	assert.Nil(t, distros, "No distros exepcted.")

	dir, deferF := makeProjectTree(t)
	defer deferF()
	defer patchEnv("GOPATH", dir)()
	distros, err = AllDistributors()
	expected := []Distributor{"code.google.com", "github.com"}
	assert.Equal(t, expected, distros)
}

func TestAuthor(t *testing.T) {
	distro := Distributor("github.com")
	author := NewAuthor(distro, "foo")
	assert.Equal(t, Author("github.com/foo"), author)

	rDistro, rAuthor := author.Split()
	assert.Equal(t, rDistro, distro)
	assert.Equal(t, "foo", rAuthor)

	defer patchEnv("GOPATH", "/gopath/")()
	assert.Equal(t, "/gopath/src/github.com/foo", author.AbsPath())
}
