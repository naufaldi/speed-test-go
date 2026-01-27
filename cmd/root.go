package cmd

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
	"github.com/user/speed-test-go/internal/output"
	"github.com/user/speed-test-go/internal/test"
)

var (
	jsonFlag       bool
	bytesFlag      bool
	verboseFlag    bool
	serverIDFlag   string
	numServersFlag int
	timeoutFlag    time.Duration
)

var rootCmd = &cobra.Command{
	Use:   "speed-test",
	Short: "Test your internet connection speed and ping",
	Long: `Test your internet connection speed and ping using speedtest.net from the CLI.
    
Supports multiple output formats and configuration options.`,
	RunE: runSpeedTest,
}

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.Flags().BoolVarP(&jsonFlag, "json", "j", false, "Output the result as JSON")
	rootCmd.Flags().BoolVarP(&bytesFlag, "bytes", "b", false, "Output the result in megabytes per second (MBps)")
	rootCmd.Flags().BoolVarP(&verboseFlag, "verbose", "v", false, "Output more detailed information")
	rootCmd.Flags().StringVarP(&serverIDFlag, "server", "s", "", "Specify a server ID to use")
	rootCmd.Flags().IntVarP(&numServersFlag, "servers", "n", 5, "Number of closest servers to test for selection")
	rootCmd.Flags().DurationVarP(&timeoutFlag, "timeout", "t", 30*time.Second, "Timeout for the speed test")
}

func runSpeedTest(cmd *cobra.Command, args []string) error {
	ctx, cancel := context.WithTimeout(context.Background(), timeoutFlag)
	defer cancel()

	formatter := output.NewFormatter(bytesFlag, jsonFlag, verboseFlag)

	runner := test.NewRunner()
	runner.SetServerID(serverIDFlag)
	runner.SetNumServersToTest(numServersFlag)

	result, err := runner.Run(ctx)
	if err != nil {
		fmt.Print(formatter.FormatError(err))
		return err
	}

	fmt.Print(formatter.Format(result))

	return nil
}
