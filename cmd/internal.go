package cmd

import (
	"github.com/rtlnl/data-personalization-api/internal"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	s3BucketFlag     = "s3-bucket"
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
		dbPort := viper.GetInt(dbPortInternalFlag)
		dbNamespace := viper.GetString(dbNamespaceInternalFlag)
		s3Bucket := viper.GetString(s3BucketFlag)
		s3Region := viper.GetString(s3RegionFlag)
		s3Endpoint := viper.GetString(s3EndpointFlag)
		s3DisableSSL := viper.GetBool(s3DisableSSLFlag)

		i, err := internal.NewInternalAPI()
		if err != nil {
			panic(err)
		}

		if err = i.Run(addr, dbHost, dbNamespace, s3Bucket, s3Region, s3Endpoint, s3DisableSSL, dbPort); err != nil {
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
	f.String(s3EndpointFlag, "localhost:4572", "s3 endpoint")
	f.Bool(s3DisableSSLFlag, true, "disable SSL verification for s3")

	viper.BindEnv(addressInternalFlag, "ADDRESS_HOST")
	viper.BindEnv(dbHostInternalFlag, "DB_HOST")
	viper.BindEnv(dbPortInternalFlag, "DB_PORT")
	viper.BindEnv(dbNamespaceInternalFlag, "DB_NAMESPACE")
	viper.BindEnv(s3BucketFlag, "S3_BUCKET")
	viper.BindEnv(s3RegionFlag, "S3_REGION")
	viper.BindEnv(s3EndpointFlag, "S3_ENDPOINT")
	viper.BindEnv(s3DisableSSLFlag, "S3_DISABLE_SSL")

	viper.BindPFlags(f)
}
