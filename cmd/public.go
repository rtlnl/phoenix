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
		dbHost := viper.GetString(dbHostFlag)
		dbPort := viper.GetString(dbPortFlag)
		dbUser := viper.GetString(dbUserFlag)
		dbPassword := viper.GetString(dbPasswordFlag)
		dbName := viper.GetString(dbNameFlag)

		p, err := public.NewPublicAPI()
		if err != nil {
			panic(err)
		}

		if err = p.Run(addr, dbHost, dbPort, dbUser, dbPassword, dbName); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(publicCmd)

	f := publicCmd.Flags()

	f.String(addressFlag, ":8080", "server address")
	f.String(dbHostFlag, "127.0.0.1", "database host")
	f.String(dbPortFlag, "6379", "database port")
	f.String(dbUserFlag, "", "database username")
	f.String(dbPasswordFlag, "", "database password")
	f.String(dbNameFlag, "0", "database name")

	viper.BindEnv(addressFlag, "ADDRESS_HOST")
	viper.BindEnv(dbHostFlag, "DB_HOST")
	viper.BindEnv(dbPortFlag, "DB_PORT")
	viper.BindEnv(dbUserFlag, "DB_USER")
	viper.BindEnv(dbPasswordFlag, "DB_PASSOWRD")
	viper.BindEnv(dbNameFlag, "DB_NAME")

	viper.BindPFlags(f)

}
