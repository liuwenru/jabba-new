package semver

import (
	"reflect"
	"sort"
	"testing"
)

func TestSort(t *testing.T) {
	actual := asVersionSlice(t,
		"0.2.0", "a@1.8.10", "b@1.8.2", "0.1.20", "a@1.8.2", "0.1.10", "0.1.2", "1.9.0-10.1", "1.9.0-9.60")
	sort.Sort(sort.Reverse(VersionSlice(actual)))
	expected := asVersionSlice(t,
		"1.9.0-10.1", "1.9.0-9.60", "0.2.0", "0.1.20", "0.1.10", "0.1.2", "a@1.8.10", "a@1.8.2", "b@1.8.2")
	if !reflect.DeepEqual(actual, expected) {
		t.Fatalf("actual: %v != expected: %v", actual, expected)
	}
}

func TestParseVersionQualifier(t *testing.T) {
	for _, tc := range []struct{ raw, qualifier, ver string }{
		{"openjdk-11.0.2", "openjdk", "11.0.2"},
		{"zulu@1.17.0", "zulu", "1.17.0"},
		{"system@1.8.73", "system", "1.8.73"},
		{"11.0.21-openjdk", "", "11.0.21-openjdk"},
		{"1.8.0", "", "1.8.0"},
	} {
		v, err := ParseVersion(tc.raw)
		if err != nil {
			t.Fatalf("%s: %v", tc.raw, err)
		}
		if v.Qualifier() != tc.qualifier || v.ver.Original() != tc.ver || v.String() != tc.raw {
			t.Fatalf("%s: got qualifier %q version %q", tc.raw, v.Qualifier(), v.ver.Original())
		}
	}
	for _, raw := range []string{"openjdk11", "openjdk-", "default"} {
		if _, err := ParseVersion(raw); err == nil {
			t.Fatalf("%s: expected error", raw)
		}
	}
}

func asVersionSlice(t *testing.T, slice ...string) (r []*Version) {
	for _, value := range slice {
		ver, err := ParseVersion(value)
		if err != nil {
			t.Fatalf("err: %v", err)
		}
		r = append(r, ver)
	}
	return
}
