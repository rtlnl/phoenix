package cmd

import (
	"github.com/rtlnl/data-personalization-api/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// internalCmd represents the internal command
var internalCmd = &cobra.Command{
	Use:   "internal",
	Short: "Internal APIs for populating the personalized content",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		addr := viper.GetString(addressFlag)
		dbHost := viper.GetString(dbHostFlag)
		dbPort := viper.GetString(dbPortFlag)
		dbUser := viper.GetString(dbUserFlag)
		dbPassword := viper.GetString(dbPasswordFlag)
		dbName := viper.GetString(dbNameFlag)

		i, err := internal.NewInternalAPI()
		if err != nil {
			panic(err)
		}

		if err = i.Run(addr, dbHost, dbPort, dbUser, dbPassword, dbName); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(internalCmd)

	f := internalCmd.PersistentFlags()

	f.String(addressFlag, ":8081", "server address")
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
