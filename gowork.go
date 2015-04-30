package gowork

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"github.com/op/go-logging"
)

var log = logging.MustGetLogger("gowork")

func getGopath() string {
	return os.Getenv("GOPATH")
}

// Checks if dir is a proper directory (i.e. isDir returns true and it is
// visible).
func isProperDirectory(dir os.FileInfo) bool {
	if !dir.IsDir() {
		log.Debug("Not a directory: %v, skipping...", dir.Name())
		return false
	}

	if dir.Name()[0] == '.' {
		log.Debug("Invisible file: %v, skipping...", dir.Name())
		return false
	}

	return true
}

// A Distributor is github, bitbucket or every other hoster for goprojects.
type Distributor string

// AbsPath returns the absolute path for this distributor.
func (d Distributor) AbsPath() string {
	return path.Join(getGopath(), "src", d.Name())
}

// Name returns the name of the distro (e.g. github.com)
func (d Distributor) Name() string {
	return string(d)
}

// AllDistributors returns all `Distributor`s in the gopath.
func AllDistributors() ([]Distributor, error) {
	dirs, err := ioutil.ReadDir(path.Join(getGopath(), "src"))
	if err != nil {
		return nil, err
	}

	var distrbutors []Distributor
	for _, theDir := range dirs {
		if isProperDirectory(theDir) {
			distrbutors = append(distrbutors, Distributor(theDir.Name()))
		}
	}

	return distrbutors, nil
}

// An Author is someone who hosts code on a repo. The format is:
// distro/name (e.g. github.com/matt3o12). If an author has projects on many
// distros, there will be one author instance for each distro (e.g.
// github.com/matt3o12 and bitbucket.org/matt3o12).
// Always use NewAuthor to construct an author.
type Author string

// NewAuthor creates a new author with name `name` who hosts its code on distro.
func NewAuthor(distro Distributor, name string) Author {
	return Author(fmt.Sprintf("%v/%v", string(distro), name))
}

// Split returns the distro and name the author.
func (a Author) Split() (distro Distributor, name string) {
	split := strings.SplitN(string(a), "/", 2)
	distro, name = Distributor(split[0]), split[1]

	return
}

// AbsPath returns the absolute path to all the author's projects.
func (a Author) AbsPath() string {
	distro, name := a.Split()
	return path.Join(distro.AbsPath(), name)
}

// Error returned when author could not be found.
var ErrAuthorCouldNotBeFound = errors.New("Author could not be found")

// FindAuthor tries to find author in all distros. If the author could not be
// found, it returns the `ErrAuthorCouldNotBeFound` error.
// The first match will be returned. That means if the author has a
// reposetory on github.com and bitbucket.com, the author string will most
// likely be: bitbucket.com/name (don't rely on that, though. It might as well
// return github).
// If you need to find the author on a known Distributor, use: `FindAuthorIn`.
func FindAuthor(name string) (Author, error) {
	distros, err := AllDistributors()
	if err != nil {
		return "", err
	}

	for _, theDistro := range distros {
		author, err := FindAuthorIn(name, theDistro)
		if err == nil {
			return author, nil
		}
	}

	return "", ErrAuthorCouldNotBeFound
}

// FindAuthorIn tries to find author in the given distribution.
func FindAuthorIn(name string, distro Distributor) (Author, error) {
	files, err := ioutil.ReadDir(distro.AbsPath())
	if err != nil {
		return "", err
	}

	for _, dir := range files {
		if isProperDirectory(dir) && strings.EqualFold(name, dir.Name()) {
			return NewAuthor(distro, dir.Name()), nil
		}
	}

	return "", ErrAuthorCouldNotBeFound
}

// A Project must include the distributor and the author like so:
// distro/author/project.
// If you want to create a new project object, please use: NewProject instead.
type Project string

// NewProject creates a new project for the given author with the given name.
func NewProject(a Author, name string) Project {
	return Project(fmt.Sprintf("%v/%v", string(a), name))
}

// Split splits the project into its 3 components: the distro, the author,
// and the name
func (p Project) Split() (Distributor, Author, string) {
	split := strings.SplitN(string(p), "/", 3)
	distro := Distributor(split[0])
	return distro, NewAuthor(distro, split[1]), split[2]
}

// Distributor returns the distro of the project.
func (p Project) Distributor() Distributor {
	distro, _, _ := p.Split()
	return distro
}

// Author returns the author of the project.
func (p Project) Author() Author {
	_, author, _ := p.Split()
	return author
}

// Name returns the name of the project.
func (p Project) Name() string {
	_, _, name := p.Split()
	return name
}

// AbsPath returns the absolute path of the project using the GOPATH env
// variable.
func (p Project) AbsPath() string {
	return path.Join(p.Author().AbsPath(), p.Name())
}
