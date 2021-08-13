// Package drone
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-13
package drone

import (
	"github.com/teocci/go-samples/internal/service"
)

//- id: drone-00
//name: Drone Number 1
//services:
//- telemetry:
//# disable support for telemetry protocol.
//telemetryDisable: no
//protocol: tcp
//mavlinkVersion: 3
//host: 106.244.179.242
//port: 20102
//- stream:
//# disable support for the RTSP protocol.
//rtspDisable: no
//protocol: tcp
//host: 106.244.179.242
//port: 8554

// Drone definition
type Drone struct {
	// general
	DroneId   string                           `yaml:"droneId" json:"drone_id"`
	DroneName string                           `yaml:"droneName" json:"drone_name"`
	Services  map[service.ServiceType]struct{} `yaml:"-" json:"-"`
}
