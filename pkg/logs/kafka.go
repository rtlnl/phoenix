package logs

import (
	"strings"

	"github.com/Shopify/sarama"
)

// KakfaLog is the object for sending the logs to Kafa
type KakfaLog struct {
	Producer sarama.SyncProducer
	Topic    string
}

// NewKafkaLogs create a new object for interacting with Kafka
func NewKafkaLogs(brokers, topic string) (KakfaLog, error) {
	bs := strings.Split(brokers, ",")

	producer, err := sarama.NewSyncProducer(bs, nil)
	if err != nil {
		return KakfaLog{}, err
	}
	return KakfaLog{
		Producer: producer,
		Topic:    topic,
	}, nil
}

func (k KakfaLog) Write(rl RowLog) error {
	for _, itemScore := range rl.ItemScores {
		// create the log message
		msg, err := CreateLogMessage(rl.PublicationPoint, rl.Campaign, rl.SignalID, itemScore)
		if err != nil {
			return err
		}

		// send to kafka topic
		if _, _, err := k.Producer.SendMessage(&sarama.ProducerMessage{
			Topic: k.Topic,
			Value: sarama.StringEncoder(msg),
		}); err != nil {
			return err
		}
	}
	return nil
}

// Close closes the producer object
func (k KakfaLog) Close() error {
	return k.Producer.Close()
}
