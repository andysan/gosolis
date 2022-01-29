/*
 * SPDX-FileCopyrightText: Copyright 2019, 2022 Andreas Sandberg <andreas@sandberg.uk>
 *
 * SPDX-License-Identifier: BSD-3-Clause
 */

package hermes

import (
	"fmt"
	"os"
	"reflect"
	"strings"

	"crypto/tls"
	"crypto/x509"

	"path/filepath"

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
	basePath string
	URL      string
	ClientID string   `mapstructure:"client_id"`
	AuthCert string   `mapstructure:"auth_cert"`
	AuthKey  string   `mapstructure:"auth_key"`
	CaCerts  []string `mapstructure:"ca_certs"`

	Topic []MqttTopic
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

func mqttCreateViper(subv *viper.Viper, basePath string) (Backend, error) {
	cfg := MqttConfig{
		basePath: basePath,
	}

	hooks := viper.DecodeHook(
		mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToSliceHookFunc(","),
			stringToMessageFormatter,
		))

	if err := subv.Unmarshal(&cfg, hooks); err != nil {
		return nil, err
	}

	return cfg.Create()
}

func init() {
	bf := BackendFactory{
		CreateViper: mqttCreateViper,
	}

	Backends["mqtt"] = bf
}

func (mc *MqttConfig) certPath(name string) string {
	if filepath.IsAbs(name) {
		return name
	} else {
		return filepath.Join(mc.basePath, name)
	}
}

func (mc *MqttConfig) tlsConfig() (*tls.Config, error) {
	config := &tls.Config{}

	/* Setup TLS authentication using client ceritficates */
	if mc.AuthCert != "" || mc.AuthKey != "" {
		if mc.AuthCert == "" || mc.AuthKey == "" {
			return nil, fmt.Errorf("Public key authentication requires both a certificate and key")
		}

		cert, err := tls.LoadX509KeyPair(
			mc.certPath(mc.AuthCert),
			mc.certPath(mc.AuthKey))
		if err != nil {
			return nil, fmt.Errorf("Failed to load client certificate: %s", err)
		}

		config.Certificates = []tls.Certificate{cert}
	}

	/* Setup user-provided CA certificates. Use the system's CA
	 * pool if no CA certificates provided.
	 */
	if mc.CaCerts == nil {
		ca_pool, err := x509.SystemCertPool()
		if err != nil {
			return nil, fmt.Errorf("Failed to load system CA pool: %s", err)
		}
		config.RootCAs = ca_pool
	} else {
		config.RootCAs = x509.NewCertPool()
	}

	for _, value := range mc.CaCerts {
		pem, err := os.ReadFile(mc.certPath(value))
		if err != nil {
			return nil, fmt.Errorf("Failed to load CA cert: %s", err)
		}

		if !config.RootCAs.AppendCertsFromPEM(pem) {
			return nil, fmt.Errorf("Failed to add CA '%s' to pool", value)
		}
	}

	return config, nil
}

func (mc *MqttConfig) Create() (*Mqtt, error) {
	cli := Mqtt{
		config: mc,
	}

	// Resolve files relative the the current working directory if
	// basePath hasn't been set by mqttCreateViper
	if mc.basePath == "" {
		if pwd, err := os.Getwd(); err == nil {
			mc.basePath = pwd
		} else {
			return nil, err
		}
	}

	opts := mqtt.NewClientOptions()

	opts.AddBroker(mc.URL)
	opts.SetClientID(mc.ClientID)

	if tls, err := mc.tlsConfig(); err == nil {
		opts.SetTLSConfig(tls)
	} else {
		return nil, err
	}

	cli.client = mqtt.NewClient(opts)

	return &cli, nil
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
