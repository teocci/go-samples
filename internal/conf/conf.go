// Package conf Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-11
package conf

import (
	"encoding/base64"
	"fmt"
	"github.com/teocci/go-samples/internal/common"
	"io/ioutil"
	"os"
	"time"

	"github.com/teocci/go-samples/internal/confenv"
	"github.com/teocci/go-samples/internal/logger"

	"github.com/aler9/gortsplib/pkg/headers"
	"golang.org/x/crypto/nacl/secretbox"
	"gopkg.in/yaml.v3"
)

// Encryption is an encryption policy.
type Encryption int

// encryption policies.
const (
	EncryptionNo Encryption = iota
	EncryptionOptional
	EncryptionStrict
)

func decrypt(key string, bytes []byte) ([]byte, error) {
	enc, err := base64.StdEncoding.DecodeString(string(bytes))
	if err != nil {
		return nil, err
	}

	var secretKey [32]byte
	copy(secretKey[:], key)

	var decryptNonce [24]byte
	copy(decryptNonce[:], enc[:24])
	decrypted, ok := secretbox.Open(nil, enc[24:], &decryptNonce, &secretKey)
	if !ok {
		return nil, fmt.Errorf("decryption error")
	}

	return decrypted, nil
}

// Conf is the main program configuration.
type Conf struct {
	// general
	LogLevel              string                          `yaml:"logLevel" json:"logLevel"`
	LogLevelParsed        logger.Level                    `yaml:"-" json:"-"`
	LogDestinations       []string                        `yaml:"logDestinations" json:"logDestinations"`
	LogDestinationsParsed map[logger.Destination]struct{} `yaml:"-" json:"-"`
	LogFile               string                          `yaml:"logFile" json:"logFile"`
	ReadTimeout           time.Duration                   `yaml:"readTimeout" json:"readTimeout"`
	WriteTimeout          time.Duration                   `yaml:"writeTimeout" json:"writeTimeout"`
	ReadBufferCount       int                             `yaml:"readBufferCount" json:"readBufferCount"`

	// drones

	// rtsp
	RTSPDisable       bool                  `yaml:"rtspDisable" json:"rtspDisable"`
	Protocols         []string              `yaml:"protocols" json:"protocols"`
	ProtocolsParsed   map[common.Protocol]struct{} `yaml:"-" json:"-"`
	Encryption        string                `yaml:"encryption" json:"encryption"`
	EncryptionParsed  Encryption            `yaml:"-" json:"-"`
	RTSPAddress       string                `yaml:"rtspAddress" json:"rtspAddress"`
	RTSPSAddress      string                `yaml:"rtspsAddress" json:"rtspsAddress"`
	RTPAddress        string                `yaml:"rtpAddress" json:"rtpAddress"`
	RTCPAddress       string                `yaml:"rtcpAddress" json:"rtcpAddress"`
	MulticastIPRange  string                `yaml:"multicastIPRange" json:"multicastIPRange"`
	MulticastRTPPort  int                   `yaml:"multicastRTPPort" json:"multicastRTPPort"`
	MulticastRTCPPort int                   `yaml:"multicastRTCPPort" json:"multicastRTCPPort"`
	ServerKey         string                `yaml:"serverKey" json:"serverKey"`
	ServerCert        string                `yaml:"serverCert" json:"serverCert"`
	AuthMethods       []string              `yaml:"authMethods" json:"authMethods"`
	AuthMethodsParsed []headers.AuthMethod  `yaml:"-" json:"-"`
	ReadBufferSize    int                   `yaml:"readBufferSize" json:"readBufferSize"`

	// rtmp
	RTMPDisable bool   `yaml:"rtmpDisable" json:"rtmpDisable"`
	RTMPAddress string `yaml:"rtmpAddress" json:"rtmpAddress"`

	// paths
	Paths map[string]*PathConf `yaml:"paths" json:"paths"`
}

// Load loads a Conf.
func Load(filePath string) (*Conf, bool, error) {
	conf := &Conf{}

	// read from file
	found, err := func() (bool, error) {
		// rtsp-simple-server.yml is optional
		if filePath == "app-receiver.yml" {
			if _, err := os.Stat(filePath); err != nil {
				return false, nil
			}
		}

		fileBytes, err := ioutil.ReadFile(filePath)
		if err != nil {
			return true, err
		}

		if key, ok := os.LookupEnv("RTSP_CONFKEY"); ok {
			fileBytes, err = decrypt(key, fileBytes)
			if err != nil {
				return true, err
			}
		}

		err = yaml.Unmarshal(fileBytes, conf)
		if err != nil {
			return true, err
		}

		return true, nil
	}()
	if err != nil {
		return nil, false, err
	}

	// read from environment
	err = confenv.Load("RTSP", conf)
	if err != nil {
		return nil, false, err
	}

	err = conf.CheckAndFillMissing()
	if err != nil {
		return nil, false, err
	}

	return conf, found, nil
}

func (conf *Conf) parseLogLevel() error {
	switch conf.LogLevel {
	case logger.WarnString:
		conf.LogLevelParsed = logger.Warn
		return nil
	case logger.InfoString:
		conf.LogLevelParsed = logger.Info
		return nil
	case logger.DebugString:
		conf.LogLevelParsed = logger.Debug
		return nil
	default:
		return fmt.Errorf("unsupported log level: %s", conf.LogLevel)
	}
}

// CheckAndFillMissing checks the configuration for errors and fill missing fields.
func (conf *Conf) CheckAndFillMissing() error {
	if conf.LogLevel == "" {
		conf.LogLevel = logger.GetDefaultLevelName()
	}

	err := conf.parseLogLevel()
	if err != nil {
		return err
	}

	if len(conf.LogDestinations) == 0 {
		conf.LogDestinations = []string{"stdout"}
	}
	conf.LogDestinationsParsed = make(map[logger.Destination]struct{})
	for _, dest := range conf.LogDestinations {
		switch dest {
		case "stdout":
			conf.LogDestinationsParsed[logger.DestinationStdout] = struct{}{}

		case "file":
			conf.LogDestinationsParsed[logger.DestinationFile] = struct{}{}

		case "syslog":
			conf.LogDestinationsParsed[logger.DestinationSyslog] = struct{}{}

		default:
			return fmt.Errorf("unsupported log destination: %s", dest)
		}
	}

	if conf.LogFile == "" {
		conf.LogFile = "app-receiver.log"
	}
	if conf.ReadTimeout == 0 {
		conf.ReadTimeout = 10 * time.Second
	}
	if conf.WriteTimeout == 0 {
		conf.WriteTimeout = 10 * time.Second
	}
	if conf.ReadBufferCount == 0 {
		conf.ReadBufferCount = 512
	}

	if conf.APIAddress == "" {
		conf.APIAddress = "127.0.0.1:9997"
	}

	if conf.MetricsAddress == "" {
		conf.MetricsAddress = "127.0.0.1:9998"
	}

	if conf.PPROFAddress == "" {
		conf.PPROFAddress = "127.0.0.1:9999"
	}

	if len(conf.Protocols) == 0 {
		conf.Protocols = []string{"udp", "multicast", "tcp"}
	}
	conf.ProtocolsParsed = make(map[Protocol]struct{})
	for _, proto := range conf.Protocols {
		switch proto {
		case "udp":
			conf.ProtocolsParsed[ProtocolUDP] = struct{}{}

		case "multicast":
			conf.ProtocolsParsed[ProtocolMulticast] = struct{}{}

		case "tcp":
			conf.ProtocolsParsed[ProtocolTCP] = struct{}{}

		default:
			return fmt.Errorf("unsupported protocol: %s", proto)
		}
	}
	if len(conf.ProtocolsParsed) == 0 {
		return fmt.Errorf("no protocols provided")
	}

	if conf.Encryption == "" {
		conf.Encryption = "no"
	}
	switch conf.Encryption {
	case "no", "false":
		conf.EncryptionParsed = EncryptionNo

	case "optional":
		conf.EncryptionParsed = EncryptionOptional

	case "strict", "yes", "true":
		conf.EncryptionParsed = EncryptionStrict

		if _, ok := conf.ProtocolsParsed[ProtocolUDP]; ok {
			return fmt.Errorf("encryption can't be used with the UDP stream protocol")
		}

	default:
		return fmt.Errorf("unsupported encryption value: '%s'", conf.Encryption)
	}

	if conf.RTSPAddress == "" {
		conf.RTSPAddress = ":8554"
	}
	if conf.RTSPSAddress == "" {
		conf.RTSPSAddress = ":8555"
	}
	if conf.RTPAddress == "" {
		conf.RTPAddress = ":8000"
	}
	if conf.RTCPAddress == "" {
		conf.RTCPAddress = ":8001"
	}
	if conf.MulticastIPRange == "" {
		conf.MulticastIPRange = "224.1.0.0/16"
	}
	if conf.MulticastRTPPort == 0 {
		conf.MulticastRTPPort = 8002
	}
	if conf.MulticastRTCPPort == 0 {
		conf.MulticastRTCPPort = 8003
	}

	if conf.ServerKey == "" {
		conf.ServerKey = "server.key"
	}
	if conf.ServerCert == "" {
		conf.ServerCert = "server.crt"
	}

	if len(conf.AuthMethods) == 0 {
		conf.AuthMethods = []string{"basic", "digest"}
	}
	for _, method := range conf.AuthMethods {
		switch method {
		case "basic":
			conf.AuthMethodsParsed = append(conf.AuthMethodsParsed, headers.AuthBasic)

		case "digest":
			conf.AuthMethodsParsed = append(conf.AuthMethodsParsed, headers.AuthDigest)

		default:
			return fmt.Errorf("unsupported authentication method: %s", method)
		}
	}

	if conf.RTMPAddress == "" {
		conf.RTMPAddress = ":1935"
	}

	if len(conf.Paths) == 0 {
		conf.Paths = map[string]*PathConf{
			"all": {},
		}
	}

	// "all" is an alias for "~^.*$"
	if _, ok := conf.Paths["all"]; ok {
		conf.Paths["~^.*$"] = conf.Paths["all"]
		delete(conf.Paths, "all")
	}

	for name, pconf := range conf.Paths {
		if pconf == nil {
			conf.Paths[name] = &PathConf{}
			pconf = conf.Paths[name]
		}

		err := pconf.checkAndFillMissing(name)
		if err != nil {
			return err
		}
	}

	return nil
}
