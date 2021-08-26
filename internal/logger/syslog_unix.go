// Package logger
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-11
//go:build !windows && !plan9 && !nacl
// +build !windows,!plan9,!nacl

package logger

import (
	"io"
	native "log/syslog"
)

type syslog struct {
	inner *native.Writer
}

func newSyslog(prefix string) (io.WriteCloser, error) {
	inner, err := native.New(native.LOG_INFO|native.LOG_DAEMON, prefix)
	if err != nil {
		return nil, err
	}

	return &syslog{
		inner: inner,
	}, nil
}

func (ls *syslog) Close() error {
	return ls.inner.Close()
}

func (ls *syslog) Write(p []byte) (int, error) {
	return ls.inner.Write(p)
}
