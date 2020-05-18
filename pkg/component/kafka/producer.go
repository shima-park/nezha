package kafka

import (
	"reflect"

	"github.com/shima-park/nezha/pkg/common/config"
	"github.com/shima-park/nezha/pkg/common/log"
	"github.com/shima-park/nezha/pkg/component"

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
	Name  string   `yaml:"name"`
	Addrs []string `yaml"addrs"`
}

type Producer struct {
	config   ProducerConfig
	producer sarama.SyncProducer
	instance component.Instance
}

func NewProducer(rawConfig string) (*Producer, error) {
	var conf ProducerConfig
	err := config.Unmarshal([]byte(rawConfig), &conf)
	if err != nil {
		return nil, err
	}

	log.Info("Kafka producer config: %+v", conf)

	producer, err := sarama.NewSyncProducer(conf.Addrs, nil)
	if err != nil {
		return nil, err
	}

	return &Producer{
		config:   conf,
		producer: producer,
		instance: component.NewInstance(
			conf.Name,
			inject.InterfaceOf((*sarama.SyncProducer)(nil)),
			reflect.ValueOf(producer),
			producer,
		),
	}, nil
}

func (c *Producer) SampleConfig() string {
	conf := ProducerConfig{
		Name:  "MyKafkaProducer",
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
	return c.instance
}

func (c *Producer) Start() error {
	return nil
}

func (c *Producer) Stop() error {
	return c.producer.Close()
}
