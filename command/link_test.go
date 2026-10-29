package command

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

func TestLink(t *testing.T) {
	home := t.TempDir()
	t.Setenv("JABBA_HOME", home)
	jdk := filepath.Join(home, "jdk")
	external := filepath.Join(t.TempDir(), "jdk-11")
	for _, dir := range []string{filepath.Join(external, "bin"), filepath.Join(jdk, "11.0.21-openjdk", "bin")} {
		if err := os.MkdirAll(dir, 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "java"), nil, 0755); err != nil {
			t.Fatal(err)
		}
	}
	// internal links (minor version, alias) and a stale foreign link must not show up in Ls
	for name, target := range map[string]string{
		"11.0":    filepath.Join(jdk, "11.0.21-openjdk"),
		"default": filepath.Join(jdk, "11.0.21-openjdk"),
		"myjdk":   external,
	} {
		if err := os.Symlink(target, filepath.Join(jdk, name)); err != nil {
			t.Fatal(err)
		}
	}

	if err := Link("openjdk11", external); err == nil {
		t.Fatal("expected invalid name to be rejected")
	}
	if err := Link("openjdk-11.0.2", external); err != nil {
		t.Fatal(err)
	}
	if err := Link("openjdk-11.0.2", external); err == nil {
		t.Fatal("expected existing link to be rejected")
	}
	if err := Link("system@1.8.0", external); err != nil {
		t.Fatal(err)
	}
	if !IsLinked("openjdk-11.0.2") || IsLinked("11.0") || IsLinked("11.0.21-openjdk") {
		t.Fatal("IsLinked misclassified entries")
	}

	assertLs := func(expected ...string) {
		vs, err := Ls()
		if err != nil {
			t.Fatal(err)
		}
		var actual []string
		for _, v := range vs {
			actual = append(actual, v.String())
		}
		if !reflect.DeepEqual(actual, expected) {
			t.Fatalf("actual: %v != expected: %v", actual, expected)
		}
	}
	assertLs("11.0.21-openjdk", "openjdk-11.0.2", "system@1.8.0")

	if ver, err := LsBestMatch("openjdk-11"); err != nil || ver != "openjdk-11.0.2" {
		t.Fatalf("LsBestMatch: %v %v", ver, err)
	}
	if err := Link("openjdk-11", ""); err != nil {
		t.Fatal(err)
	}
	assertLs("11.0.21-openjdk", "system@1.8.0")
}
