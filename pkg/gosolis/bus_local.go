/*
 * SPDX-FileCopyrightText: Copyright 2022 Andreas Sandberg <andreas@sandberg.uk>
 *
 * SPDX-License-Identifier: BSD-3-Clause
 */

package gosolis

type LocalBusMessage struct {
	Sender BusInterface
	IsAck  bool
	Frame  Frame
}

type LocalBusInterface struct {
	Echo     bool
	device   DeviceId
	fromDist chan LocalBusMessage
	toDist   chan LocalBusMessage
}

// Ensure that we satisfy the BusInterface interface
var _ BusInterface = &LocalBusInterface{}

type LocalBus struct {
	LocalBusInterface
	Interfaces []*LocalBusInterface
	toDist     chan LocalBusMessage
}

func NewLocalBus(ifaces uint) *LocalBus {
	toDist := make(chan LocalBusMessage)
	bus := LocalBus{
		LocalBusInterface: LocalBusInterface{
			toDist:   toDist,
			fromDist: make(chan LocalBusMessage),
		},
		Interfaces: make([]*LocalBusInterface, ifaces),
		toDist:     toDist,
	}

	for idx, _ := range bus.Interfaces {
		bus.Interfaces[idx] = &LocalBusInterface{
			fromDist: make(chan LocalBusMessage),
			toDist:   toDist,
			device:   DeviceId(idx + 1),
		}
	}

	go bus.run()

	return &bus
}

func (b *LocalBus) run() {
	allInterfaces := append(b.Interfaces, &b.LocalBusInterface)
	for msg := range b.toDist {
		for _, iface := range allInterfaces {
			if !iface.Echo && iface == msg.Sender {
				continue
			}

			iface.fromDist <- msg
		}
	}
}

func (b *LocalBusInterface) ReadFrame() (*Frame, error) {
	msg := <-b.fromDist
	if msg.IsAck {
		return &msg.Frame, IllegalFrameError
	} else {
		return &msg.Frame, nil
	}
}

func (b *LocalBusInterface) ReadAckFrame() (*Frame, error) {
	msg := <-b.fromDist
	if !msg.IsAck {
		return &msg.Frame, IllegalFrameError
	} else {
		return &msg.Frame, nil
	}
}

func (b *LocalBusInterface) WriteFrame(frame *Frame) error {
	b.toDist <- LocalBusMessage{
		Sender: b,
		Frame:  *frame,
	}

	return nil
}

func (b *LocalBusInterface) WriteAck(dev DeviceId, cmd Command) error {
	b.toDist <- LocalBusMessage{
		Sender: b,
		IsAck:  true,
		Frame: Frame{
			Device:  dev,
			Command: cmd,
		},
	}

	return nil
}
