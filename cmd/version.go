package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number of speed-test",
	Long:  `All software has versions. This is speed-test's`,
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("speed-test v1.0.0")
		os.Exit(0)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
