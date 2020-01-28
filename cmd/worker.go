package cmd

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/rtlnl/phoenix/worker"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var ()

// workerCmd represents the internal command
var workerCmd = &cobra.Command{
	Use:   "worker",
	Short: "Worker for queueing and running tasks",
	Long: `This command will start the server for creating tasks
	that will be executed one they arrive.`,
	Run: func(cmd *cobra.Command, args []string) {
		addr := viper.GetString(addressWorkerFlag)
		brokerWorker := viper.GetString(brokerWorkerFlag)
		resultBackendWorker := viper.GetString(resultBackendWorkerFlag)

		w, err := worker.NewWorker(brokerWorker, resultBackendWorker)
		if err != nil {
			panic(err)
		}

		if err := w.RegisterTasks(nil); err != nil {
			panic(err)
		}

		// start prometheus here
		http.Handle("/metrics", promhttp.HandlerFor(
			prometheus.DefaultGatherer,
			promhttp.HandlerOpts{},
		))
		go func() {
			http.ListenAndServe(addr, nil)
		}()

		if err := w.Launch(); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(workerCmd)

	f := workerCmd.PersistentFlags()

	f.String(addressWorkerFlag, ":9000", "prometheus address")
	f.String(brokerWorkerFlag, "redis://127.0.0.1:6379", "broker url for the workers")
	f.String(resultBackendWorkerFlag, "redis://127.0.0.1:6379", "result backend url for the workers")

	viper.BindEnv(addressWorkerFlag, "ADDRESS_HOST")
	viper.BindEnv(brokerWorkerFlag, "BROKER_URL")
	viper.BindEnv(resultBackendWorkerFlag, "RESULT_BACKEND_URL")

	viper.BindPFlags(f)
}
