package cmd

import (
	"errors"

	"github.com/rtlnl/phoenix/pkg/logs"
	"github.com/rtlnl/phoenix/public"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	tucsonGRPCAddressFlag          = "tucson-address"
	recommendationLogsFlag         = "recommendation-logs"
	recommendationKafkaBrokersFlag = "brokers"
	recommendationKafkaTopicFlag   = "topic"
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
		dbPort := viper.GetInt(dbPortPublicFlag)
		dbNamespace := viper.GetString(dbNamespacePublicFlag)
		logType := viper.GetString(recommendationLogsFlag)
		brokers := viper.GetString(recommendationKafkaBrokersFlag)
		topic := viper.GetString(recommendationKafkaTopicFlag)
		tucsonAddress := viper.GetString(tucsonGRPCAddressFlag)

		// create recommendation logger
		recLogs, err := setRecommendationLogging(logType, brokers, topic)
		if err != nil {
			panic(err)
		}

		// if log type is kafka, we need to close the producer when the server stops
		if _, ok := recLogs.(logs.KakfaLog); ok {
			defer recLogs.(logs.KakfaLog).Close()
		}

		// create new Public api object
		p, err := public.NewPublicAPI(dbHost, dbNamespace, dbPort, tucsonAddress, recLogs)
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
	f.StringP(dbHostPublicFlag, "d", "127.0.0.1", "database host")
	f.StringP(dbNamespacePublicFlag, "n", "personalization", "namespace of the database")
	f.IntP(dbPortPublicFlag, "p", 3000, "database port")

	// optional parameters
	f.StringP(recommendationLogsFlag, "l", "stdout", "[LOGS] where to store the recommendation logs. Accepted type: stdout,kafka. Kafka needs two extra flags: brokers and topic")
	f.StringP(recommendationKafkaBrokersFlag, "b", "", "[LOGS] kafka brokers composed by host and port separated by comma. Example: broker1:9092,broker2:9093")
	f.StringP(recommendationKafkaTopicFlag, "t", "", "[LOGS] kafka topic where to send the recommendation logs")

	// tucson parameters
	f.String(tucsonGRPCAddressFlag, "", "tucson api gRPC server address")

	viper.BindEnv(addressPublicFlag, "ADDRESS_HOST")
	viper.BindEnv(dbHostPublicFlag, "DB_HOST")
	viper.BindEnv(dbPortPublicFlag, "DB_PORT")
	viper.BindEnv(dbNamespacePublicFlag, "DB_NAMESPACE")
	viper.BindEnv(tucsonGRPCAddressFlag, "TUCSON_ADDRESS")
	viper.BindEnv(recommendationLogsFlag, "REC_LOGS_TYPE")
	viper.BindEnv(recommendationKafkaBrokersFlag, "REC_LOGS_BROKERS")
	viper.BindEnv(recommendationKafkaTopicFlag, "REC_LOGS_TOPIC")

	viper.BindPFlags(f)
}

func setRecommendationLogging(logType, brokers, topic string) (logs.RecommendationLog, error) {
	switch logType {
	case "stdout":
		return logs.NewStdoutLog(), nil
	case "kafka":
		return logs.NewKafkaLogs(brokers, topic)
	default:
		return nil, errors.New("recommendation logging type not valid. Use only stdout or kafka")
	}
}
