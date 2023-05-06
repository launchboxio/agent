package cmd

import (
	"fmt"
	"github.com/launchboxio/agent/pkg/agent"
	"github.com/spf13/cobra"
	"log"
	"os"
	"sync"
)

var rootCmd = &cobra.Command{
	Use:   "agent",
	Short: "LaunchboxHQ Agent",
	Run: func(cmd *cobra.Command, args []string) {
		url, _ := cmd.Flags().GetString("url")
		token, _ := cmd.Flags().GetString("token")
		clusterId, _ := cmd.Flags().GetString("cluster-id")

		ag := &agent.Agent{
			Url:     url,
			Token:   token,
			Channel: fmt.Sprintf("clusters:%s", clusterId),
		}

		var wg sync.WaitGroup

		wg.Add(2)
		go func() {
			err := ag.Run(&wg)
			if err != nil {
				log.Fatal(err)
			}
		}()

		go func() {
			err := ag.Metrics(&wg)
			if err != nil {
				log.Fatal(err)
			}
		}()

		wg.Wait()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().String("url", "https://api.launchboxhq.io:443/connection/websocket", "URL for webhook events")

	rootCmd.Flags().String("token", "", "Authentication token")

	rootCmd.Flags().String("cluster-id", "", "Cluster ID")
	_ = rootCmd.MarkFlagRequired("cluster-id")
}
