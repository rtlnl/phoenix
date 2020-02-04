package cmd

import (
	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/rtlnl/phoenix/pkg/db"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/rtlnl/phoenix/internal"
	md "github.com/rtlnl/phoenix/middleware"
)

var (
	s3RegionFlag     = "s3-region"
	s3EndpointFlag   = "s3-endpoint"
	s3DisableSSLFlag = "s3-disable-ssl"
)

// internalCmd represents the internal command
var internalCmd = &cobra.Command{
	Use:   "internal",
	Short: "Internal APIs for populating the personalized content",
	Long: `This command will start the server for the internal
	APIs for populating the personalized content into the database.`,
	Run: func(cmd *cobra.Command, args []string) {
		addr := viper.GetString(addressInternalFlag)
		dbHost := viper.GetString(dbHostInternalFlag)
		s3Region := viper.GetString(s3RegionFlag)
		s3Endpoint := viper.GetString(s3EndpointFlag)
		s3DisableSSL := viper.GetBool(s3DisableSSLFlag)
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

		// append all the middlewares here
		var middlewares []gin.HandlerFunc
		middlewares = append(middlewares, md.DB(redisClient))
		middlewares = append(middlewares, md.AWSSession(s3Region, s3Endpoint, s3DisableSSL))
		middlewares = append(middlewares, md.Cors())
		middlewares = append(middlewares, md.NewWorker(dbHost, workerProducerName, workerQueueName))

		i, err := internal.NewInternalAPI(middlewares...)
		if err != nil {
			panic(err)
		}

		if err = i.ListenAndServe(addr); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(internalCmd)

	f := internalCmd.PersistentFlags()

	f.String(addressInternalFlag, ":8081", "server address")
	f.String(dbHostInternalFlag, "127.0.0.1:6379", "database host")
	f.String(s3RegionFlag, "eu-west-1", "s3 region")
	f.String(s3EndpointFlag, "localhost:4572", "s3 endpoint")
	f.Bool(s3DisableSSLFlag, true, "disable SSL verification for s3")
	f.Bool(logDebugFlag, false, "sets log level to debug")

	viper.BindEnv(addressInternalFlag, "ADDRESS_HOST")
	viper.BindEnv(dbHostInternalFlag, "DB_HOST")
	viper.BindEnv(s3RegionFlag, "S3_REGION")
	viper.BindEnv(s3EndpointFlag, "S3_ENDPOINT")
	viper.BindEnv(s3DisableSSLFlag, "S3_DISABLE_SSL")
	viper.BindEnv(logDebugFlag, "LOG_DEBUG")

	viper.BindPFlags(f)
}
