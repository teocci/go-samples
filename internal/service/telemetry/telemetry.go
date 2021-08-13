// Package telemetry
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-13
package telemetry

import (
	"github.com/teocci/go-samples/internal/common"
	"github.com/teocci/go-samples/internal/service"
)

type Telemetry struct {
	service.Base
	mavlinkVersion common.MavLinkVersion
}

// NewTelemetry return Telemetry service interface
func NewTelemetry() (service.Service, error) {
	return &Telemetry{
		service.Base{},
		common.MavLinkV3,
	}, nil
}

func (t Telemetry) Init(serviceType service.ServiceType, protocol common.Protocol, host string, port int) {
	t.Base.InitBase(serviceType, protocol, host, port)
}

func (t Telemetry) InitService() error {
	panic("implement me")
}

func (t Telemetry) GetName() string {
	panic("implement me")
}

func (t Telemetry) GetHost() string {
	panic("implement me")
}

func (t Telemetry) GetAllItems() ([]interface{}, error) {
	panic("implement me")
}

func (t Telemetry) GetItem(itemName string) (interface{}, error) {
	panic("implement me")
}
