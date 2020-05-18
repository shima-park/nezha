package kafka

import (
	"reflect"
	"time"

	"github.com/shima-park/nezha/pkg/common/config"
	"github.com/shima-park/nezha/pkg/common/log"
	"github.com/shima-park/nezha/pkg/component"

	"github.com/Shopify/sarama"
	cluster "github.com/bsm/sarama-cluster"
)

var _ component.Component = &Consumer{}

func init() {
	if err := component.Register("kafka_consumer", func(config string) (component.Component, error) {
		return NewConsumer(config)
	}); err != nil {
		panic(err)
	}
}

type ConsumerConfig struct {
	Name              string   `yaml:"name"`
	Addrs             []string `yaml:"addrs"`
	ConsumerGroup     string   `yaml:"consumer_group"`
	Topics            []string `yaml:"topics"`
	OffsetsInitial    int64    `yaml:"offsets_initial"`
	OffsetsAutoCommit bool     `yaml:"offsets_auto_commit"`
}

type Consumer struct {
	config   ConsumerConfig
	consumer *cluster.Consumer
	done     chan struct{}
	instance component.Instance
}

func NewConsumer(rawConfig string) (*Consumer, error) {
	var conf ConsumerConfig
	err := config.Unmarshal([]byte(rawConfig), &conf)
	if err != nil {
		return nil, err
	}

	log.Info("Kafka consumer config: %+v", conf)

	kafkaConf := cluster.NewConfig()
	kafkaConf.Consumer.Return.Errors = true
	kafkaConf.Group.Return.Notifications = true
	kafkaConf.Consumer.Offsets.Initial = conf.OffsetsInitial

	// sarama >= 1.26 https://github.com/Shopify/sarama/issues/1638
	// fix it: panic: non-positive interval for NewTicker
	kafkaConf.Consumer.Offsets.CommitInterval = time.Second
	kafkaConf.Consumer.Offsets.AutoCommit.Enable = conf.OffsetsAutoCommit

	consumer, err := cluster.NewConsumer(
		conf.Addrs, conf.ConsumerGroup, conf.Topics, kafkaConf)
	if err != nil {
		return nil, err
	}

	return &Consumer{
		config:   conf,
		consumer: consumer,
		done:     make(chan struct{}),
		instance: component.NewInstance(
			conf.Name,
			reflect.TypeOf(consumer),
			reflect.ValueOf(consumer),
			consumer,
		),
	}, nil
}

// SampleConfig returns the default configuration of the Input
func (c *Consumer) SampleConfig() string {
	conf := ConsumerConfig{
		Name:              "MyKafkaConsumer",
		Addrs:             []string{"localhost:9092"},
		ConsumerGroup:     "my_consumer_group",
		Topics:            []string{"my_topics"},
		OffsetsInitial:    sarama.OffsetNewest,
		OffsetsAutoCommit: true,
	}

	b, _ := config.Marshal(&conf)

	return string(b)
}

// Description returns a one-sentence description on the Input
func (c *Consumer) Description() string {
	return "kafka consumer factory"
}

func (c *Consumer) Instance() component.Instance {
	return c.instance
}

func (c *Consumer) Start() error {
	go func() {
		for {
			select {
			case ntf := <-c.consumer.Notifications():
				log.Info("Rebalanced: %+v\n", ntf)
			case err := <-c.consumer.Errors():
				log.Error("Error: %s\n", err.Error())
			case <-c.done:
				return
			}

		}
	}()
	return nil
}

func (c *Consumer) Stop() error {
	select {
	case <-c.done:
		return nil
	default:
		close(c.done)

		return c.consumer.Close()
	}
}
