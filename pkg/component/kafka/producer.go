package kafka

import (
	"nezha/pkg/component"
	"nezha/pkg/common/config"
	"reflect"

	"github.com/Shopify/sarama"
	"github.com/shima-park/inject"
)

var _ component.Component = &Producer{}

func init() {
	if err := component.Register("kafka_producer", func(config string) (component.Component, error) {
		return NewProducer(config)
	}); err != nil {
		panic(err)
	}
}

type ProducerConfig struct {
	Name  string
	Addrs []string
}

type Producer struct {
	config   ProducerConfig
	producer sarama.SyncProducer
}

func NewProducer(rawConfig string) (*Producer, error) {
	var conf ProducerConfig
	err := config.Unmarshal([]byte(rawConfig), &conf)
	if err != nil {
		return nil, err
	}

	producer, err := sarama.NewSyncProducer(conf.Addrs, nil)
	if err != nil {
		return nil, err
	}

	return &Producer{
		config:   conf,
		producer: producer,
	}, nil
}

func (c *Producer) SampleConfig() string {
	conf := ProducerConfig{
		Addrs: []string{"localhost:9092"},
	}

	b, _ := config.Marshal(&conf)
	return string(b)
}

// Description returns a one-sentence description on the Input
func (c *Producer) Description() string {
	return "kafka producer factory"
}

func (c *Producer) Instance() component.Instance {
	return component.Instance{
		Name:      c.config.Name,
		Type:      inject.InterfaceOf((*sarama.SyncProducer)(nil)),
		Value:     reflect.ValueOf(c.producer),
		Interface: c.producer,
	}
}

func (c *Producer) Start() error {
	return nil
}

func (c *Producer) Stop() error {
	return c.producer.Close()
}
