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
		path.Join("aaa", "user", "project"),
		path.Join("bbb", "user", "project"),
		path.Join("ccc", "user", "project"),
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

	assertDir("aaa", "user", "project")
	assertDir("bbb", "user", "project")
	assertDir("ccc", "user", "project")
	assertDir("github.com", "matt3o12", "termui-widgets")
	assertDir("github.com", "matt3o12", "gowork")
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
	assert.Equal(t, "github.com", distro.Name())
}

func TestDistributor_Authors(t *testing.T) {
	dir, deferF := makeProjectTree(t)
	defer deferF()
	defer patchEnv("GOPATH", dir)()

	distro := Distributor("github.com")
	a := func(n ...string) []Author {
		authors := make([]Author, len(n))
		for i, name := range n {
			authors[i] = NewAuthor(distro, name)
		}

		return authors
	}

	authors, err := distro.Authors()
	if assert.NoError(t, err) {
		assert.Equal(t, a("matt3o12", "stretchr"), authors)
	}
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
	expected := []Distributor{
		"aaa", "bbb", "ccc",
		"code.google.com",
		"github.com",
	}
	assert.Equal(t, expected, distros)
}

func TestAuthor(t *testing.T) {
	distro := Distributor("github.com")
	author := NewAuthor(distro, "foo")
	assert.Equal(t, Author("github.com/foo"), author)

	rDistro, rAuthor := author.Split()
	assert.Equal(t, rDistro, distro)
	assert.Equal(t, "foo", rAuthor)

	assert.Equal(t, distro, author.Distributor())
	assert.Equal(t, "foo", author.Name())

	defer patchEnv("GOPATH", "/gopath/")()
	assert.Equal(t, "/gopath/src/github.com/foo", author.AbsPath())
}

func TestAuthor_Projects(t *testing.T) {
	author := Author("github.com/matt3o12")

	defer patchEnv("GOPATH", "not-exist")()
	projects, err := author.Projects()
	msg := "open not-exist/src/github.com/matt3o12: no such file or directory"
	assert.EqualError(t, err, msg)
	assert.Empty(t, projects, "Expected to return no projects, got: %v", projects)

	dir, deferF := makeProjectTree(t)
	defer deferF()
	defer patchEnv("GOPATH", dir)()

	projects, err = author.Projects()
	if assert.NoError(t, err) {
		assert.Equal(t, Project("github.com/matt3o12/gowork"), projects[0])
		assert.Equal(t, Project("github.com/matt3o12/termui-widgets"), projects[1])
	}
}

func TestFindAuthor(t *testing.T) {
	defer patchEnv("GOPATH", "not-exist")()
	author, err := FindAuthor("barfoo")

	msg := "Expected no author to be found, got: %v"
	assert.Equal(t, Author(""), author, msg, author)
	assert.EqualError(t, err, "open not-exist/src: no such file or directory")

	dir, deferF := makeProjectTree(t)
	defer deferF()
	defer patchEnv("GOPATH", dir)()

	author, err = FindAuthor("matt3o12")
	assert.NoError(t, err)
	assert.Equal(t, Author("github.com/matt3o12"), author)

	author, err = FindAuthor("does-not-exist")
	assert.Equal(t, ErrAuthorCouldNotBeFound, err)
	assert.Equal(t, Author(""), author, "No author expected")
}

func TestFindAuthorIn(t *testing.T) {
	dir, deferF := makeProjectTree(t)
	defer deferF()
	defer patchEnv("GOPATH", dir)()

	assertAuthor := func(distro Distributor, name string) {
		expected := NewAuthor(distro, name)
		author, err := FindAuthorIn(name, distro)
		assert.NoError(t, err)
		assert.Equal(t, expected, author)
	}

	assertNotAuthor := func(distro Distributor, name string) {
		author, err := FindAuthorIn(name, distro)
		msg := "Expected author %v in %v not to be found."
		assert.Error(t, err, name, msg, distro.Name())

		msg = "Did not expect to return an author. Got: %v"
		assert.Equal(t, Author(""), author, msg, author)
	}

	assertAuthor("aaa", "user")
	assertAuthor("bbb", "user")
	assertAuthor("ccc", "user")

	assertNotAuthor("github.com", "not-exist")
	assertNotAuthor("not-exist", "user")
}

func TestIsProperDirectory(t *testing.T) {
	dir, err := ioutil.TempDir("", "")
	require.NoError(t, err)
	defer os.RemoveAll(dir)

	file, err := os.Create(path.Join(dir, "test.txt"))
	require.NoError(t, err)

	stat, err := file.Stat()
	require.NoError(t, err)

	assert.False(t, isProperDirectory(stat), "expeted test.txt not to be a dir.")

	mkdir := func(name string) os.FileInfo {
		err := os.Mkdir(path.Join(dir, name), 0777)
		require.NoError(t, err)

		stat, err := os.Stat(path.Join(dir, name))
		require.NoError(t, err)

		return stat
	}

	stat = mkdir(".hidden")

	msg := "expected .hidden not to be a valid dir because it is hidden."
	assert.False(t, isProperDirectory(stat), msg)

	stat = mkdir("real-dir")

	msg = "expected real-dir to be recognized as a proper directory."
	assert.True(t, isProperDirectory(stat), msg)
}

func TestProject(t *testing.T) {
	distro := Distributor("github.com")
	author := NewAuthor(distro, "foo")
	project := NewProject(author, "bar/test")

	rDistro, rAuthor, rProject := project.Split()
	assert.Equal(t, distro, rDistro)
	assert.Equal(t, author, rAuthor)
	assert.Equal(t, "bar/test", rProject)

	assert.Equal(t, distro, project.Distributor())
	assert.Equal(t, author, project.Author())
	assert.Equal(t, "bar/test", project.Name())

	defer patchEnv("GOPATH", "/gopath")()
	assert.Equal(t, "/gopath/src/github.com/foo/bar/test", project.AbsPath())
}
