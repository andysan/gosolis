/*
 * SPDX-FileCopyrightText: Copyright 2019 Andreas Sandberg <andreas@sandberg.uk>
 *
 * SPDX-License-Identifier: BSD-3-Clause
 */

package gosolis

import (
	"errors"
	"io"
	"time"
)

var PortTimeoutError = errors.New("Read timeout")

type TimeoutReadWriter struct {
	port io.ReadWriter

	readChannel chan byte
	timeout     time.Duration
	error       error
}

var _ io.ReadWriter = &TimeoutReadWriter{}

func NewTimeoutReadWriter(port io.ReadWriter, timeout time.Duration,
	buffer uint) *TimeoutReadWriter {

	trw := TimeoutReadWriter{
		port:        port,
		readChannel: make(chan byte, buffer),
		timeout:     timeout,
		error:       nil,
	}

	go trw.receive()

	return &trw
}

func (trw *TimeoutReadWriter) receive() {
	for {
		b := []byte{0}
		n, err := trw.port.Read(b)
		if err != nil {
			trw.error = err
			close(trw.readChannel)
			return
		}

		if n > 0 {
			trw.readChannel <- b[0]
		}
	}
}

func (trw *TimeoutReadWriter) Read(p []byte) (n int, err error) {
	timer := time.NewTimer(trw.timeout)
	defer timer.Stop()

	for i := range p {
		select {
		case value, ok := <-trw.readChannel:
			if !ok {
				return i, trw.error
			} else {
				p[i] = value
			}
		case <-timer.C:
			return i, PortTimeoutError
		}

	}

	return len(p), nil
}

func (trw *TimeoutReadWriter) Write(p []byte) (n int, err error) {
	return trw.port.Write(p)
}
