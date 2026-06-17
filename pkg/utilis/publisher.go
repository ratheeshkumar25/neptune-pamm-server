// Author: Ratheesh G Kumar — Backend Engineer (Golang)
// Destination: Back-End/neptune-pamm-server/pkg/utilis/publisher.go
// Role: Frameworks & drivers — NATS event publisher
// Description: A thin adapter that satisfies the application's EventPublisher port using core
// NATS (at-most-once). Notification events are non-critical, so at-most-once is acceptable; swap
// to PublishJSONJS here if a subject ever needs durable delivery.

package utilis

import "github.com/nats-io/nats.go"

// NatsPublisher publishes events on core NATS.
type NatsPublisher struct {
	nc *nats.Conn
}

// NewNatsPublisher wraps a NATS connection as a publisher.
func NewNatsPublisher(nc *nats.Conn) *NatsPublisher { return &NatsPublisher{nc: nc} }

// Publish marshals payload to JSON and publishes it on subject.
func (p *NatsPublisher) Publish(subject Subject, payload any) error {
	return PublishJSON(p.nc, subject, payload)
}
