package cmd

import (
	"github.com/rtlnl/data-personalization-api/public"
	"github.com/spf13/cobra"
)

// publicCmd represents the public command
var publicCmd = &cobra.Command{
	Use:   "public",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		p, err := public.NewPublic()
		if err != nil {
			panic(err)
		}

		if err = p.Run(); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(publicCmd)
}
