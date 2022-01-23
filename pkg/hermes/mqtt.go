/*
 * SPDX-FileCopyrightText: Copyright 2019 Andreas Sandberg <andreas@sandberg.uk>
 *
 * SPDX-License-Identifier: BSD-3-Clause
 */

package hermes

import (
	"fmt"
	"reflect"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type MqttTopic struct {
	Topic    string
	QOS      byte
	Retained bool
	Format   MessageFormatter
}

type MqttConfig struct {
	URL      string
	ClientID string `mapstructure:"client_id"`
	Topic    []MqttTopic
}

type Mqtt struct {
	client mqtt.Client
	config *MqttConfig
	topic  string
}

func stringToMessageFormatter(from reflect.Type, to reflect.Type,
	data interface{}) (interface{}, error) {

	fmtType := reflect.TypeOf((*MessageFormatter)(nil)).Elem()
	if from.Kind() != reflect.String || !to.Implements(fmtType) {
		return data, nil
	}

	raw := data.(string)
	return createFormatter(raw)
}

func mqttCreateViper(subv *viper.Viper) (Backend, error) {
	cfg := MqttConfig{}

	hooks := viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToSliceHookFunc(","),
			stringToMessageFormatter,
		))

	if err := subv.Unmarshal(&cfg, hooks); err != nil {
		return nil, err
	}

	cli := cfg.Create()

	return cli, nil
}

func init() {
	bf := BackendFactory{
		CreateViper: mqttCreateViper,
	}

	Backends["mqtt"] = bf
}

func (mc *MqttConfig) Create() *Mqtt {
	cli := Mqtt{
		config: mc,
	}

	opts := mqtt.NewClientOptions()

	opts.AddBroker(mc.URL)
	opts.SetClientID(mc.ClientID)

	cli.client = mqtt.NewClient(opts)

	return &cli
}

func sync(t mqtt.Token) error {
	if !t.Wait() && t.Error() == nil {
		return fmt.Errorf("MQTT failed with unexpected wait() return value")
	} else {
		return t.Error()
	}
}

func (mc *Mqtt) syncPublish(topic string, qos byte, retained bool, payload interface{}) error {
	return sync(mc.client.Publish(topic, qos, retained, payload))
}

func (mc *Mqtt) Connect() error {
	return sync(mc.client.Connect())
}

func (mt *MqttTopic) getTopic(fm *FormattedMessage) string {
	if len(fm.Path) > 0 {
		return fmt.Sprintf("%s/%s",
			mt.Topic, strings.Join(fm.Path, "/"))
	} else {
		return mt.Topic
	}
}

func (mc *Mqtt) sendMessages(topic *MqttTopic, msgs []FormattedMessage) {
	for _, m := range msgs {
		t := topic.getTopic(&m)
		if err := mc.syncPublish(t, topic.QOS, topic.Retained, m.Message); err != nil {

			Log.Printf("Failed to publish '%s': %s\n", t, err)
		}
	}
}

func (mc *Mqtt) SendMessage(m *Message) error {
	for _, t := range mc.config.Topic {
		if msgs, err := t.Format.FormatMessage(m); err == nil {
			mc.sendMessages(&t, msgs)
		} else {
			Log.Printf("Topic '%s' failed: %s\n", t.Topic, err)
		}
	}
	return nil
}
