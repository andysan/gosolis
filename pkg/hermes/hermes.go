/*
 * Copyright (c) 2019 Andreas Sandberg
 * All rights reserved.
 *
 * SPDX-License-Identifier: BSD-3-Clause
 */

package hermes

import (
	"errors"
	"log"
	"os"
	"time"

	"github.com/spf13/viper"
)

var Log = log.New(os.Stdout, "Hermes: ", log.LstdFlags)

var BackendBlocked = errors.New("Backend blocking")

type Message struct {
	When    time.Time
	Message map[string]interface{}
}

type Backend interface {
	Connect() error
	SendMessage(m *Message) error
}

type BackendFactory struct {
	CreateViper func(viper *viper.Viper) (Backend, error)
}

type Hermes struct {
	mailbox chan *Message
	backend map[string]Backend
}

var Backends map[string]BackendFactory = make(map[string]BackendFactory)

func NewViper(v *viper.Viper) *Hermes {
	settings := v.AllSettings()
	h := Hermes{
		mailbox: make(chan *Message),
		backend: map[string]Backend{},
	}

	// Look for all of the subsections in the current section.
	for name, value := range settings {
		// Subsections are dictionaries, so ignore anything
		// that isn't a dictionary.
		if _, ok := value.(map[string]interface{}); !ok {
			continue
		}

		sub := v.Sub(name)
		t := sub.GetString("type")
		bf, ok := Backends[t]
		if !ok {
			Log.Printf("Illegal backend type '%s'", t)
			return nil
		}

		Log.Printf("Creating backend '%s' of type '%s'...", name, t)
		if b, err := bf.CreateViper(sub); err == nil {
			h.backend[name] = b
		} else {
			Log.Print("Failed to create backend: ", err)
			return nil
		}
	}

	for name, b := range h.backend {
		Log.Printf("Connecting backend %s...", name)
		if err := b.Connect(); err != nil {
			Log.Print("Connection failed: ", err)
			return nil
		}
	}

	go h.distributor()

	return &h
}

func (h *Hermes) Send(message map[string]interface{}) error {
	m := Message{
		When:    time.Now(),
		Message: message,
	}

	select {
	case h.mailbox <- &m:
		return nil
	default:
		return BackendBlocked
	}

	return nil
}

func (h *Hermes) distributor() {
	for m := range h.mailbox {
		h.propagate(m)
	}
}

func (h *Hermes) propagate(m *Message) {
	for i, b := range h.backend {
		err := b.SendMessage(m)
		if err != nil {
			Log.Printf("Backend '%s' failed to send message: %s",
				i, err)
		}
	}
}
