package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	addressPublicFlag      = "address-host-public"
	dbHostPublicFlag       = "db-host-public"
	dbPasswordPublicFlag   = "db-password-public"
	addressInternalFlag    = "address-host-internal"
	dbHostInternalFlag     = "db-host-internal"
	dbPasswordInternalFlag = "db-password-internal"
	workerBrokerFlag       = "worker-broker-url"
	workerPasswordFlag     = "worker-password"
	logDebugFlag           = "debug"
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "phoenix",
	Short: "Phoenix root command for initializing APIs",
	Long: `Phoenix is able to spin up two different type of services: Internal
	or Public APIs. The internal APIs have the objective of populating the personalized content
	given from the data-science team. The Public APIs will be the frontend part where clients
	can connect to`,
}

// Execute will start the application
func Execute() {
	cobra.OnInitialize(initConfig)
	if err := rootCmd.Execute(); err != nil {
		log.Fatal().Err(err).Msg("")
	}
}

// initConfig sets AutomaticEnv in viper to true.
func initConfig() {
	viper.AutomaticEnv() // read in environment variables that match
}
