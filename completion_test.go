package main

import (
	"bytes"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

func completionTestCommand() *cobra.Command {
	root := &cobra.Command{Use: "jabba"}
	for _, name := range []string{"use", "which", "uninstall", "unlink", "link", "alias", "unalias", "install", "ls", "ls-remote", "current"} {
		cmd := &cobra.Command{Use: name, Run: func(*cobra.Command, []string) {}}
		root.AddCommand(cmd)
	}
	install, _, _ := root.Find([]string{"install"})
	install.Flags().StringP("output", "o", "", "")
	ls, _, _ := root.Find([]string{"ls"})
	ls.Flags().String("latest", "", "")
	remote, _, _ := root.Find([]string{"ls-remote"})
	remote.Flags().String("os", "linux", "")
	remote.Flags().String("arch", "amd64", "")
	configureCompletion(root)
	return root
}

func TestArgumentCompletion(t *testing.T) {
	home := t.TempDir()
	t.Setenv("JABBA_HOME", home)
	for _, name := range []string{"temurin@17.0.1", "zulu@21.0.1", "system@11.0.1"} {
		if err := os.MkdirAll(filepath.Join(home, "jdk", name), 0755); err != nil {
			t.Fatal(err)
		}
	}
	for _, name := range []string{"default", "work"} {
		if err := os.WriteFile(filepath.Join(home, name+".alias"), []byte("temurin@17.0.1"), 0644); err != nil {
			t.Fatal(err)
		}
	}
	root := completionTestCommand()
	tests := []struct {
		name   string
		args   []string
		prefix string
		want   []string
		dirs   bool
	}{
		{name: "use", prefix: "tem", want: []string{"temurin@17.0.1"}},
		{name: "use", prefix: "w", want: []string{"work"}},
		{name: "which", prefix: "def", want: []string{"default"}},
		{name: "uninstall", prefix: "system@"},
		{name: "uninstall", prefix: "zulu", want: []string{"zulu@21.0.1"}},
		{name: "unlink", want: []string{"system@11.0.1"}},
		{name: "link", want: []string{"system@11.0.1"}},
		{name: "link", args: []string{"system@11"}, dirs: true},
		{name: "link", args: []string{"system@11", "/jdk"}},
		{name: "alias", want: []string{"default", "work"}},
		{name: "alias", args: []string{"default"}, prefix: "tem", want: []string{"temurin@17.0.1"}},
		{name: "alias", args: []string{"default", "17"}},
		{name: "unalias", want: []string{"default", "work"}},
		{name: "unalias", args: []string{"work"}},
		{name: "use", args: []string{"17"}},
		{name: "install"},
		{name: "current"},
	}
	for _, tt := range tests {
		t.Run(tt.name+"/"+tt.prefix+"/"+strings.Join(tt.args, ","), func(t *testing.T) {
			cmd, _, _ := root.Find([]string{tt.name})
			got, directive := cmd.ValidArgsFunction(cmd, tt.args, tt.prefix)
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("candidates = %v, want %v", got, tt.want)
			}
			wantDirective := cobra.ShellCompDirectiveNoFileComp
			if tt.dirs {
				wantDirective = cobra.ShellCompDirectiveFilterDirs
			}
			if directive != wantDirective {
				t.Errorf("directive = %v, want %v", directive, wantDirective)
			}
		})
	}
}

func TestCompletionEmptyHome(t *testing.T) {
	t.Setenv("JABBA_HOME", filepath.Join(t.TempDir(), "not installed"))
	root := completionTestCommand()
	cmd, _, _ := root.Find([]string{"use"})
	got, directive := cmd.ValidArgsFunction(cmd, nil, "")
	if len(got) != 0 || directive != cobra.ShellCompDirectiveNoFileComp {
		t.Fatalf("empty installation completion = %v, %v", got, directive)
	}
	cmd, _, _ = root.Find([]string{"alias"})
	got, _ = cmd.ValidArgsFunction(cmd, nil, "")
	if !reflect.DeepEqual(got, []string{"default"}) {
		t.Fatalf("new alias completion = %v", got)
	}
}

func TestCompletionScripts(t *testing.T) {
	for _, shell := range []string{"bash", "zsh"} {
		t.Run(shell, func(t *testing.T) {
			root := completionTestCommand()
			var out bytes.Buffer
			root.SetOut(&out)
			root.SetArgs([]string{"completion", shell})
			if err := root.Execute(); err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(out.String(), "__complete") {
				t.Fatal("script must use Cobra's dynamic completion protocol")
			}
		})
	}
}

func TestFlagCompletion(t *testing.T) {
	tests := []struct {
		command string
		flag    string
		want    []string
	}{
		{"ls", "latest", []string{"major", "minor", "patch"}},
		{"ls-remote", "os", []string{"darwin", "linux", "windows"}},
		{"ls-remote", "arch", []string{"amd64", "386", "arm", "arm64"}},
	}
	root := completionTestCommand()
	for _, tt := range tests {
		cmd, _, _ := root.Find([]string{tt.command})
		fn, ok := cmd.GetFlagCompletionFunc(tt.flag)
		if !ok {
			t.Fatalf("missing completion for %s --%s", tt.command, tt.flag)
		}
		got, directive := fn(cmd, nil, "")
		if !reflect.DeepEqual(got, tt.want) || directive != cobra.ShellCompDirectiveNoFileComp {
			t.Errorf("%s --%s = %v, %v", tt.command, tt.flag, got, directive)
		}
	}
}
