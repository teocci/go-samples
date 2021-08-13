// Package service
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-13
package service

import (
	"github.com/teocci/go-samples/internal/common"
)

// ServiceType is an encryption policy.
type ServiceType int

// encryption policies.
const (
	TelemetryService ServiceType = iota
	StreamService
)

const (
	TelemetryServiceName = "telemetry"
	StreamServiceName    = "stream"
)

type Base struct {
	serviceType ServiceType
	protocol    common.Protocol
	host        string
	port        int
	enable		bool
}

type Service interface {
	// InitService is a service
	InitService() error

	// GetName return browser name
	GetName() string

	// GetHost return browser secret key
	GetHost() string

	// GetAllItems return all items (password|bookmark|cookie|history)
	GetAllItems() ([]data.Item, error)

	// GetItem return single one from the password|bookmark|cookie|history
	GetItem(itemName string) (data.Item, error)
}

func (b *Base) InitBase(serviceType ServiceType, protocol common.Protocol, host string, port int) {
	b.serviceType = serviceType
	b.protocol = protocol
	b.host = host
	b.port = port
}

func (b *Base) GetServiceType(serviceType ServiceType, protocol common.Protocol, host string, port int) {
	b.serviceType = serviceType
	b.protocol = protocol
	b.host = host
	b.port = port
}
