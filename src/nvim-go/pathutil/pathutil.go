// Copyright 2016 The nvim-go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package pathutil

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/neovim/go-client/nvim"
)

var pkgPathutil = "pathutil"

// Chdir changes the vim current working directory.
// The returned function restores working directory to `getcwd()` result path
// and unlocks the mutex.
func Chdir(v *nvim.Nvim, dir string) func() {
	var (
		m   sync.Mutex
		cwd interface{}
	)
	m.Lock()
	if err := v.Eval("getcwd()", &cwd); err != nil {
		return nil
	}
	v.SetCurrentDirectory(dir)
	return func() {
		v.SetCurrentDirectory(cwd.(string))
		m.Unlock()
	}
}

func TrimGoPath(p string) string {
	p = strings.TrimPrefix(p, os.Getenv("GOPATH")) + string(filepath.Separator)
	switch p[:4] {
	case "/bin", "/pkg", "/src":
		return filepath.Clean(p[5:])
	default:
		return filepath.Clean(p)
	}
}

// JoinGoPath joins the $GOPATH + "src" to p
func JoinGoPath(p string) string {
	return filepath.Join(os.Getenv("GOPATH"), "src", p)
}

// ShortFilePath return the simply trim cwd into p.
func ShortFilePath(p, cwd string) string {
	return strings.Replace(p, cwd, ".", 1)
}

// Rel wrapper of filepath.Rel function that return only one variable.
func Rel(cwd, f string) string {
	rel, err := filepath.Rel(cwd, f)
	if err != nil {
		return f
	}
	return rel
}

// ExpandGoRoot expands the "$GOROOT" include from p.
func ExpandGoRoot(p string) string {
	if strings.Index(p, "$GOROOT") != -1 {
		return strings.Replace(p, "$GOROOT", runtime.GOROOT(), 1)
	}

	return p // Not hit
}

// IsDir returns whether the filename is directory.
func IsDir(filename string) bool {
	fi, err := os.Stat(filename)
	return err == nil && fi.IsDir()
}

// IsExist returns whether the filename is exists.
func IsExist(filename string) bool {
	_, err := os.Stat(filename)
	return !os.IsNotExist(err) || err == nil
}

// IsNotExist returns whether the filename is exists.
func IsNotExist(filename string) bool {
	_, err := os.Stat(filename)
	return os.IsNotExist(err)
}

// IsGoFile returns whether the filename is exists.
func IsGoFile(filename string) bool {
	f, err := os.Stat(filename)
	return err == nil && filepath.Ext(f.Name()) == ".go"
}
