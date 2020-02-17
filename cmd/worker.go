package cmd

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog/log"
	"github.com/rtlnl/phoenix/pkg/db"
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

		// instantiate Redis client
		redisClient, err := db.NewRedisClient(brokerWorker)
		if err != nil {
			panic(err)
		}

		l, err := redisClient.Lock(worker.WorkerLockKey)
		if l == false || err != nil {
			log.Error().Msg(err.Error())
			os.Exit(0)
		}

		w, err := worker.New(brokerWorker, workerConsumerName, workerQueueName)
		if err != nil {
			panic(err)
		}

		if err := w.Consume(); err != nil {
			panic(err)
		}

		log.Info().Msg(" [*] Waiting for messages. To exit press CTRL+C")

		sigterm := make(chan os.Signal, 1)
		signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
		select {
		case <-sigterm:
			log.Info().Msg("terminating: via signal")
			w.Close()

			l, err := redisClient.Unlock(worker.WorkerLockKey)
			if l == false || err != nil {
				panic(err)
			}
		}
		log.Info().Msg("queue close. Cleaning up...")
	},
}

func init() {
	rootCmd.AddCommand(workerCmd)

	f := workerCmd.PersistentFlags()

	f.String(workerBrokerFlag, "127.0.0.1:6379", "broker url for the workers")

	viper.BindEnv(workerBrokerFlag, "WORKER_BROKER_URL")

	viper.BindPFlags(f)
}
