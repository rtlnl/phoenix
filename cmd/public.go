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

		addr := viper.GetString(addressPublicFlag)
		dbHost := viper.GetString(dbHostPublicFlag)
		dbPort := viper.GetInt(dbPortPublicFlag)
		dbNamespace := viper.GetString(dbNamespacePublicFlag)

		p, err := public.NewPublicAPI()
		if err != nil {
			panic(err)
		}

		if err = p.Run(addr, dbHost, dbNamespace, dbPort); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(publicCmd)

	f := publicCmd.Flags()

	f.String(addressPublicFlag, ":8082", "server address")
	f.String(dbHostPublicFlag, "127.0.0.1", "database host")
	f.String(dbNamespacePublicFlag, "personalization", "namespace of the database")
	f.Int(dbPortPublicFlag, 3000, "database port")

	viper.BindEnv(addressPublicFlag, "ADDRESS_HOST")
	viper.BindEnv(dbHostPublicFlag, "DB_HOST")
	viper.BindEnv(dbPortPublicFlag, "DB_PORT")
	viper.BindEnv(dbNamespacePublicFlag, "DB_NAMESPACE")

	viper.BindPFlags(f)
}
