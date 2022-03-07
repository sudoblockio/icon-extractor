package kafka

import (
	"time"

	"github.com/Shopify/sarama"
	"github.com/cenkalti/backoff/v4"
	"go.uber.org/zap"

	"github.com/sudoblockio/icon-go-etl/config"
)

type KafkaTopicProducer struct {
	BrokerURL       string
	TopicName       string
	TopicPartitions int
	TopicChan       chan *sarama.ProducerMessage
}

// map[TopicName] -> Producer
var KafkaTopicProducers = map[string]*KafkaTopicProducer{}

func StartProducers() {
	kafkaBroker := config.Config.KafkaBrokerURL
	blocksTopic := config.Config.KafkaBlocksTopic
	blocksPartitions := config.Config.KafkaBlocksPartitions
	deadMessageTopic := config.Config.KafkaDeadMessageTopic

	KafkaTopicProducers[blocksTopic] = &KafkaTopicProducer{
		kafkaBroker,
		blocksTopic,
		blocksPartitions,
		make(chan *sarama.ProducerMessage),
	}

	KafkaTopicProducers[deadMessageTopic] = &KafkaTopicProducer{
		kafkaBroker,
		deadMessageTopic,
		1,
		make(chan *sarama.ProducerMessage),
	}

	go KafkaTopicProducers[blocksTopic].produceTopic()
	go KafkaTopicProducers[deadMessageTopic].produceTopic()
}

func (k *KafkaTopicProducer) produceTopic() {
	config := sarama.NewConfig()

	//////////////////
	// Create topic //
	//////////////////
	admin, err := getAdmin(k, config)
	if err != nil {
		zap.S().Fatal("KAFKA ADMIN ERROR: ", err.Error())
	}
	defer func() { _ = admin.Close() }()

	// check if topic is already made
	topics, err := admin.ListTopics()
	if _, ok := topics[k.TopicName]; ok == false {

		// Create topic
		err = admin.CreateTopic(k.TopicName, &sarama.TopicDetail{
			NumPartitions:     int32(k.TopicPartitions),
			ReplicationFactor: 1,
		}, false)
		if err != nil {
			zap.S().Fatal("Error while creating topic: ", err.Error())
		}
	}

	/////////////////////
	// Create producer //
	/////////////////////
	config.Producer.Partitioner = sarama.NewRandomPartitioner
	config.Producer.RequiredAcks = sarama.WaitForAll
	config.Producer.Return.Successes = true
	producer, err := getProducer(k, config)
	if err != nil {
		zap.S().Fatal("KAFKA PRODUCER ERROR: Finally Connection cannot be established")
	}
	defer func() { _ = producer.Close() }()

	for {
		topicMessage := <-k.TopicChan

		partition, offset, err := producer.SendMessage(topicMessage)
		if err != nil {
			zap.S().Warn("Producer ", k.TopicName, ": Err sending message=", err.Error())
		}

		topicMessageKey, _ := topicMessage.Key.Encode()
		zap.S().Info("Topic=", k.TopicName, " Partition=", partition, " offset=", offset, " key=", string(topicMessageKey), " - Produced message to kafka")
	}
}

func getAdmin(k *KafkaTopicProducer, config *sarama.Config) (sarama.ClusterAdmin, error) {
	var admin sarama.ClusterAdmin
	operation := func() error {
		a, err := sarama.NewClusterAdmin([]string{k.BrokerURL}, config)
		if err != nil {
			zap.S().Warn("KAFKA ADMIN NEWCLUSTERADMIN WARN: ", err.Error())
		} else {
			admin = a
		}
		return err
	}
	neb := backoff.NewConstantBackOff(time.Second)
	err := backoff.Retry(operation, neb)
	return admin, err
}

func getProducer(k *KafkaTopicProducer, config *sarama.Config) (sarama.SyncProducer, error) {
	var producer sarama.SyncProducer
	operation := func() error {
		pro, err := sarama.NewSyncProducer([]string{k.BrokerURL}, config)
		if err != nil {
			zap.S().Warn("KAFKA PRODUCER NEWSYNCPRODUCER WARN: ", err.Error())
		} else {
			producer = pro
		}
		return err
	}
	neb := backoff.NewConstantBackOff(time.Second)
	err := backoff.Retry(operation, neb)
	return producer, err
}
