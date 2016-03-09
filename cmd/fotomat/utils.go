package main

import (
	"net"
	"net/url"
	"os"
)

func exists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return err == nil
}

func isTimeout(err error) bool {
	if err == nil {
		return false
	}
	switch err := err.(type) {
	case net.Error:
		return err.Timeout()
	case *url.Error:
		// Only necessary for Go < 1.6.
		if err, ok := err.Err.(net.Error); ok {
			return err.Timeout()
		}
	}
	return false
}
