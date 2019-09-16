package cmd

import (
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Here we place the common variables and then will be available for subcommands
var (
	addressPublicFlag       = "address-host-public"
	dbHostPublicFlag        = "db-host-public"
	dbPortPublicFlag        = "db-port-public"
	dbNamespacePublicFlag   = "db-namespace-public"
	addressInternalFlag     = "address-host-internal"
	dbHostInternalFlag      = "db-host-internal"
	dbPortInternalFlag      = "db-port-internal"
	dbNamespaceInternalFlag = "db-namespace-internal"
)

// rootCmd represents the base command when called without any subcommands
// Descriptors form -h/--help
var rootCmd = &cobra.Command{
	Use:   "personalization",
	Short: "Personalization root command for initializing APIs",
	Long: `Personalization is able to spin up two different type of services: Internal
	or Public APIs. The internal APIs have the objective of populating the personalized content
	given from the data-science team. The Public APIs will be the frontend part where clients
	can connect to`,
}

// Execute will start the application
// cobra: cli
// viper: arguments, environmental variables, read from files
// Funcion name must start in capital letter when you what to make it visible (otherwise will be private)
// Return always an error
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
