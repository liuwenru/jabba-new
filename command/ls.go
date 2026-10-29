package command

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"

	"github.com/Jabba-Team/jabba/cfg"
	"github.com/Jabba-Team/jabba/semver"
)

var readDir = ioutil.ReadDir

func Ls() ([]*semver.Version, error) {
	files, _ := readDir(filepath.Join(cfg.Dir(), "jdk"))
	var r []*semver.Version
	for _, f := range files {
		if f.IsDir() {
			v, err := semver.ParseVersion(f.Name())
			if err != nil {
				return nil, err
			}
			r = append(r, v)
		} else if f.Mode()&os.ModeSymlink == os.ModeSymlink && IsLinked(f.Name()) {
			// skip foreign symlinks instead of failing (e.g. stale links left after $JABBA_HOME was moved)
			v, err := semver.ParseVersion(f.Name())
			if err != nil || v.Qualifier() == "" {
				continue
			}
			r = append(r, v)
		}
	}
	sort.Sort(sort.Reverse(semver.VersionSlice(r)))
	return r, nil
}

func LsBestMatch(selector string) (ver string, err error) {
	vs, err := Ls()
	if err != nil {
		return
	}
	return LsBestMatchWithVersionSlice(vs, selector)
}

func LsBestMatchWithVersionSlice(vs []*semver.Version, selector string) (ver string, err error) {
	rng, err := semver.ParseRange(selector)
	if err != nil {
		return
	}
	for _, v := range vs {
		if rng.Contains(v) {
			ver = v.String()
			break
		}
	}
	if ver == "" {
		err = fmt.Errorf("%s isn't installed", rng)
	}
	return
}
