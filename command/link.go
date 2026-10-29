package command

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/Jabba-Team/jabba/cfg"
	"github.com/Jabba-Team/jabba/semver"
	log "github.com/sirupsen/logrus"
)

func Link(selector string, dir string) error {
	// name has to be <vendor>-<version> or <vendor>@<version>, <version> has to be valid per semver
	if v, err := semver.ParseVersion(selector); err != nil || v.Qualifier() == "" {
		return fmt.Errorf("Name must be <vendor>-<version> (e.g. 'openjdk-11.0.2'), got '%s'", selector)
	}
	if dir == "" {
		ver, err := LsBestMatch(selector)
		if err != nil {
			return err
		}
		if !IsLinked(ver) {
			return fmt.Errorf("'%s' is not a link (use 'jabba uninstall %s' instead)", ver, ver)
		}
		return os.Remove(filepath.Join(cfg.Dir(), "jdk", ver))
	} else {
		if err := assertJavaDistribution(dir, runtime.GOOS); err != nil {
			return err
		}
		if err := os.MkdirAll(filepath.Join(cfg.Dir(), "jdk"), 0755); err != nil {
			return err
		}
		target := filepath.Join(cfg.Dir(), "jdk", selector)
		if _, err := os.Lstat(target); err == nil {
			return fmt.Errorf("'%s' already exists", selector)
		}
		return os.Symlink(dir, target)
	}
}

// IsLinked reports whether jdk/<name> is a symlink to a JDK outside of $JABBA_HOME/jdk (i.e. created by `jabba link`)
func IsLinked(name string) bool {
	jdkDir := filepath.Join(cfg.Dir(), "jdk")
	target, err := os.Readlink(filepath.Join(jdkDir, name))
	if err != nil {
		return false
	}
	if !filepath.IsAbs(target) {
		target = filepath.Join(jdkDir, target)
	}
	rel, err := filepath.Rel(jdkDir, filepath.Clean(target))
	return err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

func LinkLatest() error {
	files, _ := readDir(filepath.Join(cfg.Dir(), "jdk"))
	var vs, err = Ls()
	if err != nil {
		return err
	}
	cache := make(map[string]string)
	for _, f := range files {
		if f.IsDir() || f.Mode()&os.ModeSymlink == os.ModeSymlink {
			sourceVersion := f.Name()
			if strings.Count(sourceVersion, ".") == 1 && !IsLinked(sourceVersion) {
				target := GetLink(sourceVersion)
				_, err := LsBestMatchWithVersionSlice(vs, sourceVersion)
				if err != nil {
					err := os.Remove(filepath.Join(cfg.Dir(), "jdk", sourceVersion))
					if err == nil {
						log.Info(sourceVersion + " -/> " + target)
					}
					if !os.IsNotExist(err) {
						return err
					}
				} else {
					cache[sourceVersion] = target
				}
			}
		}
	}
	for _, v := range semver.VersionSlice(vs).TrimTo(semver.VPMinor) {
		sourceVersion := v.TrimTo(semver.VPMinor)
		target := filepath.Join(cfg.Dir(), "jdk", v.String())
		if v.Prerelease() == "" && cache[sourceVersion] != target && !IsLinked(v.String()) {
			source := filepath.Join(cfg.Dir(), "jdk", sourceVersion)
			log.Info(sourceVersion + " -> " + target)
			os.Remove(source)
			if err := os.Symlink(target, source); err != nil {
				return err
			}
		}
	}
	return linkAlias("default", vs)
}

func LinkAlias(name string) error {
	var vs, err = Ls()
	if err != nil {
		return err
	}
	return linkAlias(name, vs)
}

func linkAlias(name string, vs []*semver.Version) error {
	defaultAlias := GetAlias(name)
	if defaultAlias != "" {
		defaultAlias, _ = LsBestMatchWithVersionSlice(vs, defaultAlias)
	}
	sourceRef := /*"alias@" + */ name
	source := filepath.Join(cfg.Dir(), "jdk", sourceRef)
	sourceTarget := GetLink(sourceRef)
	if defaultAlias != "" {
		target := filepath.Join(cfg.Dir(), "jdk", defaultAlias)
		if sourceTarget != target {
			log.Info(sourceRef + " -> " + target)
			os.Remove(source)
			if err := os.Symlink(target, source); err != nil {
				return err
			}
		}
	} else {
		err := os.Remove(source)
		if err == nil {
			log.Info(sourceRef + " -/> " + sourceTarget)
		}
		if !os.IsNotExist(err) {
			return err
		}
	}
	return nil
}

func GetLink(name string) string {
	res, err := filepath.EvalSymlinks(filepath.Join(cfg.Dir(), "jdk", name))
	if err != nil {
		return ""
	}
	return res
}
