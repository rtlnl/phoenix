package cmd

import (
	"github.com/rtlnl/data-personalization-api/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	s3BucketFlag = "s3-bucket"
	s3RegionFlag = "s3-region"
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
		dbPort := viper.GetInt(dbPortInternalFlag)
		dbNamespace := viper.GetString(dbNamespaceInternalFlag)
		s3Bucket := viper.GetString(s3BucketFlag)
		s3Region := viper.GetString(s3RegionFlag)

		i, err := internal.NewInternalAPI()
		if err != nil {
			panic(err)
		}

		if err = i.Run(addr, dbHost, dbNamespace, s3Bucket, s3Region, dbPort); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(internalCmd)

	f := internalCmd.PersistentFlags()

	f.String(addressInternalFlag, ":8081", "server address")
	f.String(dbHostInternalFlag, "127.0.0.1", "database host")
	f.Int(dbPortInternalFlag, 3000, "database port")
	f.String(dbNamespaceInternalFlag, "personalization", "namespace of the database")
	f.String(s3BucketFlag, "test", "s3 bucket")
	f.String(s3RegionFlag, "eu-west-1", "s3 region")

	viper.BindEnv(addressInternalFlag, "ADDRESS_HOST")
	viper.BindEnv(dbHostInternalFlag, "DB_HOST")
	viper.BindEnv(dbPortInternalFlag, "DB_PORT")
	viper.BindEnv(dbNamespaceInternalFlag, "DB_NAMESPACE")
	viper.BindEnv(s3BucketFlag, "S3_BUCKET")
	viper.BindEnv(s3RegionFlag, "S3_REGION")

	viper.BindPFlags(f)
}
