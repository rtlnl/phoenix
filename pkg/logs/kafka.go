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

// KafkaSASLMechanism functional option for the kafka configuration
// Values accepted: "OAUTHBEARER", "PLAIN", "SCRAM-SHA-256", "SCRAM-SHA-512", "GSSAPI"
func KafkaSASLMechanism(m string) func(*sarama.Config) {
	return func(cfg *sarama.Config) {
		cfg.Net.SASL.Enable = true
		cfg.Net.SASL.Mechanism = sarama.SASLMechanism(m)
	}
}

// KafkaCredentials functional option for the kafka configuration
// For functional options see:
// https://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
func KafkaCredentials(username, password string) func(*sarama.Config) {
	return func(cfg *sarama.Config) {
		cfg.Net.SASL.User = username
		cfg.Net.SASL.Password = password
	}
}

// NewKafkaLogs create a new object for interacting with Kafka
func NewKafkaLogs(brokers, topic string, options ...func(*sarama.Config)) (KakfaLog, error) {
	bs := strings.Split(brokers, ",")

	cfg := sarama.NewConfig()
	cfg.Producer.Return.Errors = true
	cfg.Producer.Return.Successes = true

	for _, opt := range options {
		opt(cfg)
	}

	producer, err := sarama.NewSyncProducer(bs, cfg)
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
