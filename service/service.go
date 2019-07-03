package service

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// Service should describe the service well-enough for it to
// run on Mac, Linux, and Windows.
//
// 	&Service{
// 		// A human-friendy name
// 		Title: "Foobar App",
// 		// A computer-friendly name
// 		Name: "foobar-app",
// 		// A name for OS X plist
// 		ReverseDNS: "com.example.foobar-app",
// 		// A human-friendly description
// 		Desc: "Foobar App",
// 		// The app /service homepage
// 		URL: "https://example.com/foobar-app/",
// 		// The full path of the interpreter, if any (ruby, python, node, etc)
// 		Interpreter: "/opt/node/bin/node",
// 		// The name of the executable (or script)
// 		Exec: "foobar-app.js",
// 		// An array of arguments
// 		Argv: []string{"-c", "/path/to/config.json"},
// 		// A map of Environment variables that should be set
// 		Envs: map[string]string{
// 			PORT: "8080",
// 			ENV: "development",
// 		},
// 		// The user (Linux & Mac only).
// 		// This does not apply to userspace services.
// 		// There may be special considerations
// 		User: "www-data",
// 		// If different from User
// 		Group: "",
// 		// Whether to install as a system or user service
// 		System: false,
// 		// Whether or not the service may need privileged ports
// 		PrivilegedPorts: false,
// 	}
//
// Note that some fields are exported for templating,
// but not intended to be set by you.
// These are documented as omitted from JSON.
// Try to stick to what's outlined above.
type Service struct {
	Title               string            `json:"title"`
	Name                string            `json:"name"`
	Desc                string            `json:"desc"`
	URL                 string            `json:"url"`
	ReverseDNS          string            `json:"reverse_dns"` // i.e. com.example.foo-app
	Interpreter         string            `json:"interpreter"` // i.e. node, python
	Exec                string            `json:"exec"`
	Argv                []string          `json:"argv"`
	Workdir             string            `json:"workdir"`
	Envs                map[string]string `json:"envs"`
	User                string            `json:"user"`
	Group               string            `json:"group"`
	Home                string            `json:"-"`
	Local               string            `json:"-"`
	Logdir              string            `json:"logdir"`
	System              bool              `json:"system"`
	Restart             bool              `json:"restart"`
	Production          bool              `json:"production"`
	PrivilegedPorts     bool              `json:"privileged_ports"`
	MultiuserProtection bool              `json:"multiuser_protection"`
}

func (s *Service) Normalize(force bool) {
	if "" == s.Name {
		ext := filepath.Ext(s.Exec)
		base := filepath.Base(s.Exec[:len(s.Exec)-len(ext)])
		s.Name = strings.ToLower(base)
	}
	if "" == s.Title {
		s.Title = s.Name
	}
	if "" == s.ReverseDNS {
		// technically should be something more like "com.example." + s.Name,
		// but whatever
		s.ReverseDNS = s.Name
	}

	if !s.System {
		home, err := os.UserHomeDir()
		if nil != err {
			fmt.Fprintf(os.Stderr, "Unrecoverable Error: %s", err)
			os.Exit(4)
			return
		}
		s.Local = filepath.Join(home, ".local")
		s.Logdir = filepath.Join(home, ".local", "share", s.Name, "var", "log")
	} else {
		s.Logdir = "/var/log/" + s.Name
	}

	// Check to see if Exec exists
	//   /whatever => must exist exactly
	//   ./whatever => must exist in current or WorkDir(TODO)
	//   whatever => may also exist in {{ .Local }}/opt/{{ .Name }}/{{ .Exec }}
	_, err := os.Stat(s.Exec)
	if nil != err {
		bad := true
		if !strings.Contains(filepath.ToSlash(s.Exec), "/") {
			optpath := filepath.Join(s.Local, "/opt", s.Name, s.Exec)
			_, err := os.Stat(optpath)
			if nil == err {
				bad = false
				fmt.Fprintf(os.Stderr, "Using '%s' for '%s'\n", optpath, s.Exec)
				s.Exec = optpath
			}
		}

		if bad {
			// TODO look for it in WorkDir?
			fmt.Fprintf(os.Stderr, "Error: '%s' could not be found.\n", s.Exec)
			if !force {
				os.Exit(5)
				return
			}
			execpath, err := filepath.Abs(s.Exec)
			if nil == err {
				s.Exec = execpath
			}
			fmt.Fprintf(os.Stderr, "Using '%s' anyway.\n", s.Exec)
		}
	} else {
		execpath, err := filepath.Abs(s.Exec)
		if nil != err {
			fmt.Fprintf(os.Stderr, "Unrecoverable Error: %s", err)
			os.Exit(4)
		} else {
			s.Exec = execpath
		}
	}
}
