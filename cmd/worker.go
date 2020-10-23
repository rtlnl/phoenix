package cmd

import (
	"os"
	"os/signal"
	"syscall"
	"time"

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
		passwordWorker := viper.GetString(workerPasswordFlag)

		// instantiate Redis client
		rc, err := db.NewRedisClient(brokerWorker, db.Password(passwordWorker))
		if err != nil {
			panic(err)
		}

		l, err := rc.Lock(worker.WorkerLockKey)
		defer rc.Unlock(worker.WorkerLockKey)

		if l == false || err != nil {
			log.Error().Err(err).Msg("REDIS failed to acquire lock")
			os.Exit(0)
		}

		w, err := worker.New(rc.Client, workerConsumerName, workerQueueName)
		if err != nil {
			panic(err)
		}

		if err := w.Consume(); err != nil {
			panic(err)
		}

		log.Info().Msg(" [*] Waiting for messages. To exit press CTRL+C")

		ticker := time.NewTicker(db.TTLRefreshInterval)

		sigterm := make(chan os.Signal, 1)
		signal.Notify(sigterm, syscall.SIGINT, syscall.SIGTERM)
	EXITLOOP:
		for {
			select {
			case <-ticker.C:
				if err := rc.ExtendTTL(worker.WorkerLockKey); err != nil {
					log.Error().Msg(err.Error())
					w.Close()
					break EXITLOOP
				}
			case <-sigterm:
				log.Info().Msg("terminating: via signal")
				w.Close()
				break EXITLOOP
			}
		}
		log.Info().Msg("queue close. Cleaning up...")
	},
}

func init() {
	rootCmd.AddCommand(workerCmd)

	f := workerCmd.PersistentFlags()

	f.String(workerBrokerFlag, "127.0.0.1:6379", "broker url for the workers")
	f.String(workerPasswordFlag, "qwerty", "broker password")

	viper.BindEnv(workerBrokerFlag, "WORKER_BROKER_URL")
	viper.BindEnv(workerPasswordFlag, "WORKER_PASSWORD")

	viper.BindPFlags(f)
}
