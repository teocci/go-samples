// Package stream
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-13
package stream

import (
	"github.com/teocci/go-samples/internal/common"
	"github.com/teocci/go-samples/internal/service"
)

type Stream struct {
	service.Base
	mavlinkVersion common.MavLinkVersion
}

// NewStream return Telemetry service interface
func NewStream() (service.Service, error) {
	return &Stream{
		service.Base{},
		common.MavLinkV3,
	}, nil
}

func (t Stream) Init(serviceType service.ServiceType, protocol common.Protocol, host string, port int) {
	t.Base.InitBase(serviceType, protocol, host, port)
}

func (t Stream) InitService() error {
	panic("implement me")
}

func (t Stream) GetName() string {
	panic("implement me")
}

func (t Stream) GetHost() string {
	panic("implement me")
}

func (t Stream) GetAllItems() ([]interface{}, error) {
	panic("implement me")
}

func (t Stream) GetItem(itemName string) (interface{}, error) {
	panic("implement me")
}

