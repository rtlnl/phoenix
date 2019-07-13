package cmd

import (
	"github.com/rtlnl/data-personalization-api/public"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// publicCmd represents the public command
var publicCmd = &cobra.Command{
	Use:   "public",
	Short: "Starts the public APIs for personalized content",
	Long: `This command will start the server for the public
APIs for serving the personalized content.`,
	Run: func(cmd *cobra.Command, args []string) {

		addr := viper.GetString(addressFlag)
		dbHosts := viper.GetString(dbHostsFlag)
		dbPassword := viper.GetString(dbPasswordFlag)

		p, err := public.NewPublicAPI()
		if err != nil {
			panic(err)
		}

		if err = p.Run(addr, dbHosts, dbPassword); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(publicCmd)

	f := publicCmd.Flags()

	f.String(addressFlag, ":8081", "server address")
	f.String(dbHostsFlag, ":7000,:7001,:7002,:7003,:7004,:7005", "database hosts separated by comma")
	f.String(dbPasswordFlag, "", "database password")

	viper.BindEnv(addressFlag, "ADDRESS_HOST")
	viper.BindEnv(dbHostsFlag, "DB_HOSTS")
	viper.BindEnv(dbPasswordFlag, "DB_PASSOWRD")

	viper.BindPFlags(f)
}
