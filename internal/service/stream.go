// Package service
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-13
package service

import (
	"github.com/teocci/go-samples/internal/common"
)

type Stream struct {
	Base
	mavlinkVersion common.MavLinkVersion
}

// NewStream return Telemetry service interface
func NewStream() (Service, error) {
	return &Stream{
		Base{},
		common.MavLinkV3,
	}, nil
}

func (t Stream) Init(serviceType ServiceType, protocol common.Protocol, host string, port int) {
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

