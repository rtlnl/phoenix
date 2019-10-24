package cmd

import (
	"errors"

	"github.com/rtlnl/phoenix/pkg/logs"
	"github.com/rtlnl/phoenix/public"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
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
		dbPort := viper.GetInt(dbPortPublicFlag)
		dbNamespace := viper.GetString(dbNamespacePublicFlag)
		logType := viper.GetString(recommendationLogsFlag)
		tucsonAddress := viper.GetString(tucsonGRPCAddressFlag)

		// create recommendation logger
		recLogs, err := setRecommendationLogging(logType)
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
	viper.BindEnv(dbPortPublicFlag, "DB_PORT")
	viper.BindEnv(dbNamespacePublicFlag, "DB_NAMESPACE")
	viper.BindEnv(tucsonGRPCAddressFlag, "TUCSON_ADDRESS")
	viper.BindEnv(recommendationLogsFlag, "REC_LOGS_TYPE")
	viper.BindEnv(recommendationKafkaBrokersFlag, "REC_LOGS_BROKERS")
	viper.BindEnv(recommendationKafkaTopicFlag, "REC_LOGS_TOPIC")

	viper.BindPFlags(f)
}

func setRecommendationLogging(logType string) (logs.RecommendationLog, error) {
	switch logType {
	case "stdout":
		return logs.NewStdoutLog(), nil
	case "kafka":
		brokers := viper.GetString(recommendationKafkaBrokersFlag)
		topic := viper.GetString(recommendationKafkaTopicFlag)
		username := viper.GetString(recommendationUsernameFlag)
		password := viper.GetString(recommendationPasswordFlag)
		saslMechanism := viper.GetString(recommendationKafkaSASLMechanismFlag)
		kup := logs.KafkaCredentials(username, password)
		sm := logs.KafkaSASLMechanism(saslMechanism)

		return logs.NewKafkaLogs(brokers, topic, kup, sm)
	case "es":
		hosts := viper.GetString(recommendationESHostsFlag)
		index := viper.GetString(recommendationESIndexFlag)
		username := viper.GetString(recommendationUsernameFlag)
		password := viper.GetString(recommendationPasswordFlag)

		eup := logs.ESCredentials(username, password)

		return logs.NewElasticSearchLogs(hosts, index, eup)
	default:
		return nil, errors.New("recommendation logging type not valid. Use only stdout or kafka")
	}
}
