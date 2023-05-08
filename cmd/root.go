package cmd

import (
	"fmt"
	"github.com/launchboxio/agent/pkg/agent"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/clientcredentials"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
	metrics "k8s.io/metrics/pkg/client/clientset/versioned"
	"log"
	"os"
)

var rootCmd = &cobra.Command{
	Use:   "agent",
	Short: "LaunchboxHQ Agent",
	Run: func(cmd *cobra.Command, args []string) {
		//url, _ := cmd.Flags().GetString("url")
		clientId, _ := cmd.Flags().GetString("client-id")
		clientSecret, _ := cmd.Flags().GetString("client-secret")
		authUrl, _ := cmd.Flags().GetString("auth-url")
		apiUrl, _ := cmd.Flags().GetString("api-url")

		ctx := context.Background()
		conf := &clientcredentials.Config{
			ClientID:     clientId,
			ClientSecret: clientSecret,
			TokenURL:     fmt.Sprintf("%s/oauth/token", authUrl),
		}

		token, err := conf.Token(ctx)
		if err != nil {
			log.Fatal(err)
		}

		// use the current context in kubeconfig
		config, err := clientcmd.BuildConfigFromFlags("", "/Users/rwittman/.kube/config")
		if err != nil {
			panic(err.Error())
		}

		// create the clientset
		clientset, err := kubernetes.NewForConfig(config)
		metricsClient, err := metrics.NewForConfig(config)
		ag := &agent.Agent{
			Url: apiUrl,
			// TODO: We need refreshing of this token
			Token:         token.AccessToken,
			Client:        clientset,
			MetricsClient: metricsClient,
		}
		//
		err = ag.Run()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().String("auth-url", "https://auth.launchboxhq.io:443/connection/websocket", "URL for webhook events")
	rootCmd.Flags().String("api-url", "https://api.launchboxhq.io", "Api Endpoint for launchbox")
	rootCmd.Flags().String("cluster-id", "", "Cluster ID")
	_ = rootCmd.MarkFlagRequired("cluster-id")

	rootCmd.Flags().String("client-id", "", "Application ID for the cluster")
	rootCmd.Flags().String("client-secret", "", "Application secret for the cluster")
}

// go run main.go --url ws://localhost:8000/connection/websocket --token 'eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM3MjIiLCJleHAiOjE2ODM5NDgzNTMsImlhdCI6MTY4MzM0MzU1M30.21ahIMPjE-oKFTIZgxj4mx0Ovew41nXPD1pIm0x2SAo' --cluster-id 1
