// +build !appengine,!go1.5

package kvconfig

import "syscall"

var lookupEnv = syscall.Getenv
