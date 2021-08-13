// Package logger
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-11
package logger

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"sync"
	"time"

	"github.com/gookit/color"
)

const prefix string = "[rtt-jinan]"

const DebugString = "debug"
const WarnString = "warn"
const InfoString = "info"

// Level is a log level.
type Level int

// Log levels.
const (
	Debug Level = iota
	Info
	Warn
)

// Destination is a log destination.
type Destination int

const (
	// DestinationStdout writes logs to the standard output.
	DestinationStdout Destination = iota

	// DestinationFile writes logs to a file.
	DestinationFile

	// DestinationSyslog writes logs to the system logger.
	DestinationSyslog
)

// Logger is a log handler.
type Logger struct {
	level        Level
	destinations map[Destination]struct{}

	mutex   sync.Mutex
	file    *os.File
	syslog  io.WriteCloser
	buffers []bytes.Buffer
}

func GetDefaultLevelName() string {
	return levelNames()[0]
}

func GetLevelName(level Level) string {
	return levelNames()[level]
}

func levelNames() []string {
	return []string{DebugString, InfoString, WarnString}
}

func levelColors() []string {
	return []string{color.Debug.Code(), color.Info.Code(), color.Warn.Code()}
}

func levelColor(level Level) string {
	return levelColors()[level]
}

func levelPrefixes() []string {
	return []string{"D ", "I ", "W "}
}

func levelPrefix(level Level) string {
	return levelPrefixes()[level]
}

// New allocates a log handler.
func New(level Level, destinations map[Destination]struct{}, filePath string) (*Logger, error) {
	lh := &Logger{
		level:        level,
		destinations: destinations,
	}

	lh.buffers = []bytes.Buffer{bytes.Buffer{}, bytes.Buffer{}, bytes.Buffer{}}

	if _, ok := destinations[DestinationFile]; ok {
		var err error
		lh.file, err = os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
		if err != nil {
			lh.Close()
			return nil, err
		}
	}

	if _, ok := destinations[DestinationSyslog]; ok {
		var err error
		lh.syslog, err = newSyslog(prefix)
		if err != nil {
			lh.Close()
			return nil, err
		}
	}

	return lh, nil
}

// Close closes a log handler.
func (lh *Logger) Close() {
	if lh.file != nil {
		lh.file.Close()
	}

	if lh.syslog != nil {
		lh.syslog.Close()
	}
}

// itoa
// Cheap integer to fixed-width decimal ASCII. Give a negative width to avoid zero-padding.
// https://golang.org/src/log/log.go#L78
func itoa(i int, wid int) []byte {
	// Assemble decimal in reverse order.
	var b [20]byte
	bp := len(b) - 1
	for i >= 10 || wid > 1 {
		wid--
		q := i / 10
		b[bp] = byte('0' + i - q*10)
		bp--
		i = q
	}
	// i < 10
	b[bp] = byte('0' + i)
	return b[bp:]
}

func bufferDate(now time.Time, buf *bytes.Buffer) {
	year, month, day := now.Date()
	buf.Write(itoa(year, 4))
	buf.WriteByte('/')
	buf.Write(itoa(int(month), 2))
	buf.WriteByte('/')
	buf.Write(itoa(day, 2))
	buf.WriteByte(' ')
}

func bufferClock(now time.Time, buf *bytes.Buffer) {
	hour, min, sec := now.Clock()
	buf.Write(itoa(hour, 2))
	buf.WriteByte(':')
	buf.Write(itoa(min, 2))
	buf.WriteByte(':')
	buf.Write(itoa(sec, 2))
	buf.WriteByte(' ')
}

func writeTime(buf *bytes.Buffer, doColor bool) {
	intBuffer := &bytes.Buffer{}

	now := time.Now()

	// date
	bufferDate(now, intBuffer)

	// time
	bufferClock(now, intBuffer)

	if doColor {
		buf.WriteString(color.RenderString(color.Gray.Code(), intBuffer.String()))
	} else {
		buf.WriteString(intBuffer.String())
	}
}

func writeLevel(buf *bytes.Buffer, level Level, doColor bool) {
	if doColor {
		buf.WriteString(color.RenderString(levelColor(level), levelPrefix(level)))
	} else {
		buf.WriteString(levelPrefix(level))
	}
}

func writeContent(buf *bytes.Buffer, format string, args []interface{}) {
	buf.Write([]byte(fmt.Sprintf(format, args...)))
	buf.WriteByte('\n')
}

func (lh *Logger) getDestinationBuffer(destination Destination) bytes.Buffer {
	return lh.buffers[destination]
}

func logEntry(buff *bytes.Buffer, level Level, doColor bool, format string, args ...interface{}) {
	buff.Reset()
	writeTime(buff, doColor)
	writeLevel(buff, level, doColor)
	writeContent(buff, format, args)
}

// Log writes a log entry.
func (lh *Logger) Log(level Level, format string, args ...interface{}) {
	if level < lh.level {
		return
	}

	lh.mutex.Lock()
	defer lh.mutex.Unlock()

	if _, ok := lh.destinations[DestinationStdout]; ok {
		buff := lh.getDestinationBuffer(DestinationStdout)
		logEntry(&buff, level, true, format, args)
		print(buff.String())
	}

	if _, ok := lh.destinations[DestinationFile]; ok {
		buff := lh.getDestinationBuffer(DestinationFile)
		logEntry(&buff, level, false, format, args)
		_, _ = lh.file.Write(buff.Bytes())
	}

	if _, ok := lh.destinations[DestinationSyslog]; ok {
		buff := lh.getDestinationBuffer(DestinationSyslog)
		logEntry(&buff, level, false, format, args)
		_, _ = lh.syslog.Write(buff.Bytes())
	}
}
