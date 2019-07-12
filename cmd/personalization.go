package cmd

import (
	"github.com/rtlnl/data-personalization-api/personalization"
	"github.com/spf13/cobra"
)

// personalizationCmd represents the personalization command
var personalizationCmd = &cobra.Command{
	Use:   "personalization",
	Short: "",
	Long:  ``,
	Run: func(cmd *cobra.Command, args []string) {
		c, err := personalization.NewPersonalization()
		if err != nil {
			panic(err)
		}

		if err = c.Run(); err != nil {
			panic(err)
		}
	},
}

func init() {
	rootCmd.AddCommand(personalizationCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// personalizationCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// personalizationCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}
