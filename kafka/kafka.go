package kafka

import (
	"context"
	"encoding/json"
	"sync"

	"github.com/Shopify/sarama"
	"github.com/smallnest/rpcx/log"
)

var (
	Kafka kafka
	err   error
)

type kafka struct {
	client            sarama.Client
	Consumer          sarama.Consumer
	ConsumerGroup     sarama.ConsumerGroup
	AsyncProducer     sarama.AsyncProducer
	PartitionConsumer sarama.PartitionConsumer
	Messages          chan *sarama.ConsumerMessage
	CancelFunc        context.CancelFunc
	wg                *sync.WaitGroup
}

type Config struct {
	Hosts     []string
	Username  string
	Password  string
	Algorithm string
}

func (k *kafka) GenConfig(c Config) *sarama.Config {
	conf := sarama.NewConfig()
	conf.Version = sarama.V2_3_0_0
	//conf.Producer.Retry.Max = 1
	//conf.Producer.RequiredAcks = sarama.WaitForAll
	//conf.Producer.Return.Successes = true
	//conf.Metadata.Full = true
	conf.Producer.Partitioner = sarama.NewRoundRobinPartitioner
	conf.Producer.Return.Errors = false
	conf.ClientID = "sasl_scram_client"
	conf.Net.SASL.Enable = true
	conf.Net.SASL.User = c.Username
	conf.Net.SASL.Password = c.Password
	conf.Net.SASL.Handshake = true
	switch c.Algorithm {
	case "sha512":
		conf.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &XDGSCRAMClient{HashGeneratorFcn: SHA512} }
		conf.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
	case "sha256":
		conf.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &XDGSCRAMClient{HashGeneratorFcn: SHA256} }
		conf.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA256
	default:
		conf.Net.SASL.SCRAMClientGeneratorFunc = func() sarama.SCRAMClient { return &XDGSCRAMClient{HashGeneratorFcn: SHA512} }
		conf.Net.SASL.Mechanism = sarama.SASLTypeSCRAMSHA512
	}
	return conf
}

func (k *kafka) New(conf Config) {
	config := k.GenConfig(conf)
	k.client, err = sarama.NewClient(conf.Hosts, config)
	panicError(err)
	//defer k.client.Close()

	k.Consumer, err = sarama.NewConsumerFromClient(k.client)
	panicError(err)
	//defer k.GroupHandler.Close()
	topics, err := k.Consumer.Topics()
	log.Infof("kafka topics: %+v.", topics)

	k.AsyncProducer, err = sarama.NewAsyncProducerFromClient(k.client)
	panicError(err)
	//defer k.AsyncProducer.Close()
}

func (k *kafka) NewPartitionConsumer(topic string, partition int32, offset int64) {
	partitions, err := k.Consumer.Partitions(topic)
	log.Infof("all partitions of the %q topic: %+v.", topic, partitions)

	k.PartitionConsumer, err = k.Consumer.ConsumePartition(topic, partition, offset)
	if err != nil {
		log.Errorf("unable to create partition consumer: %q.", err)
	}
	//defer k.PartitionConsumer.Close()
}

func (k *kafka) NewConsumerGroup(groupID string, topics []string) {
	k.ConsumerGroup, err = sarama.NewConsumerGroupFromClient(groupID, k.client)
	if err != nil {
		log.Errorf("unable to create consumer group: %q.", err)
	}
	//defer k.GroupHandler.Close()

	consumer := GroupHandler{
		ready: make(chan bool),
	}

	Kafka.Messages = make(chan *sarama.ConsumerMessage, 10000)
	ctx, cancel := context.WithCancel(context.Background())
	k.CancelFunc = cancel
	k.wg = &sync.WaitGroup{}
	k.wg.Add(1)
	go func() {
		defer k.wg.Done()
		for err := range k.ConsumerGroup.Errors() {
			log.Errorf(err.Error())
		}
	}()

	k.wg.Add(1)
	go func() {
		defer k.wg.Done()
		for {
			if err := k.ConsumerGroup.Consume(ctx, topics, consumer); err != nil {
				log.Errorf(err.Error())
			}
			if ctx.Err() != nil {
				return
			}
			consumer.ready = make(chan bool)
		}
	}()
	<-consumer.ready
	log.Infof("Kafka consumer group up and running...")
}

func (k *kafka) ProducerMessage(topic string, key string, value string) {
	k.AsyncProducer.Input() <- &sarama.ProducerMessage{Topic: topic, Key: sarama.StringEncoder(key), Value: sarama.StringEncoder(value)}
}

func (k *kafka) ProducerMsgFormatJson(i interface{}, topic, key string) (err error) {
	value, err := json.Marshal(i)
	if err != nil {
		return
	}
	k.ProducerMessage(topic, key, string(value))
	return
}

func (k *kafka) Close() {
	k.CancelFunc()
	if err = k.PartitionConsumer.Close(); err != nil {
		log.Errorf(err.Error())
	}
	if err = k.AsyncProducer.Close(); err != nil {
		log.Errorf(err.Error())
	}
	if err = k.Consumer.Close(); err != nil {
		log.Errorf(err.Error())
	}
	if err = k.client.Close(); err != nil {
		log.Errorf(err.Error())
	}
	k.wg.Wait()
}

// GroupHandler represents a Sarama consumer group consumer
type GroupHandler struct {
	ready chan bool
}

// Setup is run at the beginning of a new session, before ConsumeClaim
func (consumer GroupHandler) Setup(sarama.ConsumerGroupSession) error {
	// Mark the consumer as ready
	close(consumer.ready)
	return nil
}

// Cleanup is run at the end of a session, once all ConsumeClaim goroutines have exited
func (consumer GroupHandler) Cleanup(sarama.ConsumerGroupSession) error {
	return nil
}

// ConsumeClaim must start a consumer loop of ConsumerGroupClaim's Messages().
func (consumer GroupHandler) ConsumeClaim(session sarama.ConsumerGroupSession, claim sarama.ConsumerGroupClaim) error {
	// NOTE:
	// Do not move the code below to a goroutine.
	// The `ConsumeClaim` itself is called within a goroutine.
	for message := range claim.Messages() {
		Kafka.Messages <- message
		session.MarkMessage(message, "")
	}
	return nil
}
