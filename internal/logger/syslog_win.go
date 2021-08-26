// Package logger
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-11
// +build windows

package logger

import (
	"fmt"
	"io"
)

func newSyslog(prefix string) (io.WriteCloser, error) {
	return nil, fmt.Errorf("not implemented on windows")
}