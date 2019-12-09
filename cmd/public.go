package cmd

import (
	"github.com/Shopify/sarama"
	"github.com/elastic/go-elasticsearch/v7"
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rtlnl/phoenix/pkg/db"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	md "github.com/rtlnl/phoenix/middleware"
	"github.com/rtlnl/phoenix/pkg/logs"
	"github.com/rtlnl/phoenix/public"
)

var (
	tucsonGRPCAddressFlag                = "tucson-address"
	recommendationLogsFlag               = "rec-logs-type"
	recommendationKafkaBrokersFlag       = "kafka-brokers"
	recommendationKafkaTopicFlag         = "kafka-topic"
	recommendationKafkaSASLMechanismFlag = "kafka-sasl-mechanism"
	recommendationESHostsFlag            = "es-hosts"
	recommendationESIndexFlag            = "es-index"
	recommendationUsernameFlag           = "log-username"
	recommendationPasswordFlag           = "log-password"
)

// publicCmd represents the public command
var publicCmd = &cobra.Command{
	Use:   "public",
	Short: "Starts the public APIs for personalized content",
	Long: `This command will start the server for the public
APIs for serving the personalized content.`,
	Run: func(cmd *cobra.Command, args []string) {
		// read parameters in input
		addr := viper.GetString(addressPublicFlag)
		dbHost := viper.GetString(dbHostPublicFlag)
		logType := viper.GetString(recommendationLogsFlag)
		tucsonAddress := viper.GetString(tucsonGRPCAddressFlag)
		logDebug := viper.GetBool(logDebugFlag)

		// log level debug
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		if logDebug {
			zerolog.SetGlobalLevel(zerolog.DebugLevel)
		}

		// instantiate Redis client
		redisClient, err := db.NewRedisClient(dbHost)
		if err != nil {
			panic(err)
		}

		// create recommendation logger
		recLogs, err := setRecommendationLogging(logType)
		if err != nil {
			panic(err)
		}

		// if log type is kafka, we need to close the producer when the server stops
		if _, ok := recLogs.(logs.KakfaLog); ok {
			defer recLogs.(logs.KakfaLog).Close()
		}

		// append all the middlewares here
		var middlewares []gin.HandlerFunc
		middlewares = append(middlewares, md.DB(redisClient))
		middlewares = append(middlewares, md.RecommendationLogs(recLogs))

		// only if we pass the tucson flag in the CLI we inject the client
		if tucsonAddress != "" {
			middlewares = append(middlewares, md.Tucson(tucsonAddress))
		}

		// create new Public api object
		p, err := public.NewPublicAPI(middlewares...)
		if err != nil {
			panic(err)
		}

		// start server
		if err := p.ListenAndServe(addr); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(publicCmd)

	f := publicCmd.Flags()

	// mandatory parameters
	f.StringP(addressPublicFlag, "a", ":8082", "server address")
	f.StringP(dbHostPublicFlag, "d", "127.0.0.1:6379", "database host")
	f.Bool(logDebugFlag, false, "sets log level to debug")

	// optional parameters
	f.StringP(recommendationLogsFlag, "l", "stdout", "[LOGS] where to store the recommendation logs. Accepted type: stdout,kafka,es")
	f.StringP(recommendationUsernameFlag, "u", "", "[LOGS] username used for either kafka or elastichsearch")
	f.StringP(recommendationPasswordFlag, "q", "", "[LOGS] password used for either kafka or elastichsearch")
	f.StringP(recommendationKafkaBrokersFlag, "b", "", "[LOGS] kafka brokers composed by host and port separated by comma. Example: broker1:9092,broker2:9093")
	f.StringP(recommendationKafkaTopicFlag, "t", "", "[LOGS] kafka topic where to send the recommendation logs")
	f.StringP(recommendationKafkaSASLMechanismFlag, "s", "", "[LOGS] kafka sasl mechanism. Accepted values 'PLAIN', 'OAUTHBEARER', 'SCRAM-SHA-256', 'SCRAM-SHA-512', 'GSSAPI'")
	f.StringP(recommendationESHostsFlag, "r", "", "[LOGS] elasticsearch addresses separated by comma. Example addr1:9200,addr2:9200")
	f.StringP(recommendationESIndexFlag, "i", "", "[LOGS] elasticsearch index on where to push the data")

	// tucson parameters
	f.String(tucsonGRPCAddressFlag, "", "tucson api gRPC server address")

	viper.BindEnv(addressPublicFlag, "ADDRESS_HOST")
	viper.BindEnv(dbHostPublicFlag, "DB_HOST")
	viper.BindEnv(logDebugFlag, "LOG_DEBUG")
	viper.BindEnv(tucsonGRPCAddressFlag, "TUCSON_ADDRESS")
	viper.BindEnv(recommendationLogsFlag, "REC_LOGS_TYPE")
	viper.BindEnv(recommendationKafkaBrokersFlag, "REC_LOGS_BROKERS")
	viper.BindEnv(recommendationKafkaTopicFlag, "REC_LOGS_TOPIC")
	viper.BindEnv(recommendationKafkaSASLMechanismFlag, "REC_LOGS_SASLMECHANISM")
	viper.BindEnv(recommendationUsernameFlag, "REC_LOGS_USERNAME")
	viper.BindEnv(recommendationPasswordFlag, "REC_LOGS_PASSWORD")

	viper.BindPFlags(f)
}

func setRecommendationLogging(logType string) (logs.RecommendationLog, error) {
	switch logType {
	case "kafka":
		var kafkaOptions []func(*sarama.Config)

		brokers := viper.GetString(recommendationKafkaBrokersFlag)
		topic := viper.GetString(recommendationKafkaTopicFlag)
		username := viper.GetString(recommendationUsernameFlag)
		password := viper.GetString(recommendationPasswordFlag)
		saslMechanism := viper.GetString(recommendationKafkaSASLMechanismFlag)

		if username != "" && password != "" {
			kafkaOptions = append(kafkaOptions, logs.KafkaCredentials(username, password))
		}

		if saslMechanism != "" {
			kafkaOptions = append(kafkaOptions, logs.KafkaSASLMechanism(saslMechanism))
		}

		return logs.NewKafkaLogs(brokers, topic, kafkaOptions...)
	case "es":
		var esConfig []func(*elasticsearch.Config)

		hosts := viper.GetString(recommendationESHostsFlag)
		index := viper.GetString(recommendationESIndexFlag)
		username := viper.GetString(recommendationUsernameFlag)
		password := viper.GetString(recommendationPasswordFlag)

		if username != "" && password != "" {
			esConfig = append(esConfig, logs.ESCredentials(username, password))
		}

		return logs.NewElasticSearchLogs(hosts, index, esConfig...)
	case "stdout":
		return logs.NewStdoutLog(), nil
	default:
		return logs.NewStdoutLog(), nil
	}
}
