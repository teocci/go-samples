// Package logger
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-11
package logger

import (
	"errors"
	"github.com/stretchr/testify/require"
	"io/ioutil"
	"os"
	"testing"
)

type Conf struct {
	// general
	LogLevel              string					`yaml:"logLevel" json:"logLevel"`
	LogLevelParsed        Level						`yaml:"-" json:"-"`
	LogDestinations       []string					`yaml:"logDestinations" json:"logDestinations"`
	LogDestinationsParsed map[Destination]struct{}	`yaml:"-" json:"-"`
	LogFile               string                    `yaml:"logFile" json:"logFile"`
}


func writeTempFile(byts []byte) (string, error) {
	tmpf, err := ioutil.TempFile(os.TempDir(), "rtsp-")
	if err != nil {
		return "", err
	}
	defer tmpf.Close()

	_, err = tmpf.Write(byts)
	if err != nil {
		return "", err
	}

	return tmpf.Name(), nil
}

func TestThis(t *testing.T) {
	var (
		logLevel string
		logLevelParsed Level
		logDestinations []string
		logDestinationsParsed map[Destination]struct{}
		logFile string
	)

	os.Setenv("RTSP_PATHS_CAM1_SOURCE", "rtsp://testing")
	defer os.Unsetenv("RTSP_PATHS_CAM1_SOURCE")

	tmpf, err := writeTempFile([]byte("{}"))
	require.NoError(t, err)
	defer os.Remove(tmpf)

	if logLevel == "" {
		logLevel = "info"
	}

	switch logLevel {
	case "warn":
		logLevelParsed = Warn

	case "info":
		logLevelParsed = Info

	case "debug":
		logLevelParsed = Debug

	default:
		err := errors.New("level not defined " + logLevel)
		require.NoError(t, err)
	}

	prefix := levelPrefix(logLevelParsed)
	require.Equal(t, "I ", prefix)

	if len(logDestinations) == 0 {
		logDestinations = []string{"stdout"}
	}
	logDestinationsParsed = make(map[Destination]struct{})
	for _, dest := range logDestinations {
		switch dest {
		case "stdout":
			logDestinationsParsed[DestinationStdout] = struct{}{}

		case "file":
			logDestinationsParsed[DestinationFile] = struct{}{}

		case "syslog":
			logDestinationsParsed[DestinationSyslog] = struct{}{}

		default:
			err := errors.New("unsupported log destination: " + dest)
			require.NoError(t, err)
		}
	}

	if logFile == "" {
		logFile = "logger-test.log"
	}
}
