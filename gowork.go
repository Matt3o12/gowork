package gowork

import (
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

// A Distributor is github, bitbucket or every other hoster for goprojects.
type Distributor string

// AbsPath returns the absolute path for this distributor.
func (d Distributor) AbsPath() string {
	return path.Join(getGopath(), "src", string(d))
}

// AllDistributors returns all `Distributor`s in the gopath.
func AllDistributors() ([]Distributor, error) {
	dirs, err := ioutil.ReadDir(path.Join(getGopath(), "src"))
	if err != nil {
		return nil, err
	}

	var distrbutors []Distributor
	for _, theDir := range dirs {
		if !theDir.IsDir() {
			log.Debug("Not a directory: %v, skipping...", theDir.Name())
			continue
		}

		if theDir.Name()[0] == '.' {
			log.Debug("Invisible file: %v, skipping...", theDir.Name())
			continue
		}

		distrbutors = append(distrbutors, Distributor(theDir.Name()))
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
