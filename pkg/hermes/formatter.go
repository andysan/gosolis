/*
 * Copyright (c) 2019 Andreas Sandberg
 * All rights reserved.
 *
 * SPDX-License-Identifier: BSD-3-Clause
 */

package hermes

import (
	"fmt"
	"reflect"

	"encoding/json"
)

type MqttMessageType uint

type FormattedMessage struct {
	Path    []string
	Message []byte
}

type MessageFormatter interface {
	FormatMessage(m *Message) ([]FormattedMessage, error)
}

type JSONFormatter struct{}

type ValueFormatter struct {
	// Include UNIX time stamp in messages
	IncludeUnixTime bool
}

func createFormatter(name string) (MessageFormatter, error) {
	switch name {
	case "json":
		return &JSONFormatter{}, nil
	case "value":
		return &ValueFormatter{}, nil
	case "time-value":
		return &ValueFormatter{
			IncludeUnixTime: true,
		}, nil
	default:
		return nil, fmt.Errorf("Illegal formatter type '%s'", name)
	}
}

func (jf *JSONFormatter) FormatMessage(msg *Message) ([]FormattedMessage, error) {
	if msg, err := json.Marshal(msg.Message); err == nil {
		return []FormattedMessage{
			FormattedMessage{
				Path:    nil,
				Message: msg,
			},
		}, nil
	} else {
		return nil, err
	}
}

func (vf *ValueFormatter) formatValue(msg *Message, v interface{}) []byte {
	st := reflect.TypeOf(v)

	prefix := ""
	if vf.IncludeUnixTime {
		prefix = fmt.Sprintf("%s%d ", prefix, msg.When.Unix())
	}

	switch st.Kind() {
	case reflect.Bool,
		reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64,
		reflect.Float32, reflect.Float64,
		reflect.Complex64, reflect.Complex128,
		reflect.String:
		return []byte(fmt.Sprintf("%s%v", prefix, v))

	case reflect.Slice, reflect.Array, reflect.Map:
		return nil

	default:
		return nil
	}
}

func (vf *ValueFormatter) FormatMessage(msg *Message) ([]FormattedMessage, error) {
	fms := make([]FormattedMessage, 0, len(msg.Message))

	for k, v := range msg.Message {
		if v := vf.formatValue(msg, v); v != nil {
			fm := FormattedMessage{
				Path:    []string{k},
				Message: v,
			}
			fms = append(fms, fm)
		}
	}

	return fms, nil
}
