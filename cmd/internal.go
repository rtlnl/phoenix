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
		addr := viper.GetString(addressFlag)
		dbHosts := viper.GetString(dbHostsFlag)
		dbPassword := viper.GetString(dbPasswordFlag)
		s3Bucket := viper.GetString(s3BucketFlag)
		s3Region := viper.GetString(s3RegionFlag)

		i, err := internal.NewInternalAPI()
		if err != nil {
			panic(err)
		}

		if err = i.Run(addr, dbHosts, dbPassword, s3Bucket, s3Region); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(internalCmd)

	f := internalCmd.PersistentFlags()

	f.String(addressFlag, ":8081", "server address")
	f.String(dbHostsFlag, ":7000,:7001,:7002,:7003,:7004,:7005", "database hosts separated by comma")
	f.String(dbPasswordFlag, "", "database password")
	f.String(s3BucketFlag, "test", "s3 bucket")
	f.String(s3RegionFlag, "eu-west-1", "s3 region")

	viper.BindEnv(addressFlag, "ADDRESS_HOST")
	viper.BindEnv(dbHostsFlag, "DB_HOSTS")
	viper.BindEnv(dbPasswordFlag, "DB_PASSOWRD")
	viper.BindEnv(s3BucketFlag, "S3_BUCKET")
	viper.BindEnv(s3RegionFlag, "S3_REGION")

	viper.BindPFlags(f)
}
