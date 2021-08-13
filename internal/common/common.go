// Package common
// Created by RTT.
// Author: teocci@yandex.com on 2021-Aug-13
package common

// Protocol is a RTSP protocol
type Protocol int

// RTSP protocols.
const (
	ProtocolUDP Protocol = iota
	ProtocolMulticast
	ProtocolTCP
)

// MavLinkVersion is a MavLink version
type MavLinkVersion int

// MavLink Versions.
const (
	MavLinkV1 MavLinkVersion = iota
	MavLinkV2
	MavLinkV3
)
