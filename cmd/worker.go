package cmd

import (
	"github.com/rtlnl/phoenix/worker"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const (
	workerConsumerName = "worker-consumer"
	workerProducerName = "worker-producer"
	workerQueueName    = "worker-queue"
)

// workerCmd represents the internal command
var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Worker for queueing and running tasks",
	Long: `This command will start the server for creating tasks
	that will be executed one they arrive.`,
	Run: func(cmd *cobra.Command, args []string) {
		brokerWorker := viper.GetString(workerBrokerFlag)

		w, err := worker.New(brokerWorker, workerConsumerName, workerQueueName)
		if err != nil {
			panic(err)
		}

		if err := w.Consume(); err != nil {
			panic(err)
		}
		// TODO: fix this with sig handling
		select {}
	},
}

func init() {
	rootCmd.AddCommand(workerCmd)

	f := workerCmd.PersistentFlags()

	f.String(workerBrokerFlag, "127.0.0.1:6379", "broker url for the workers")

	viper.BindEnv(workerBrokerFlag, "WORKER_BROKER_URL")

	viper.BindPFlags(f)
}
