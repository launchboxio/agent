package cmd

import (
	"flag"
	"fmt"
	action_cable "github.com/launchboxio/action-cable"
	"github.com/launchboxio/agent/pkg/client"
	"github.com/launchboxio/agent/pkg/pinger"
	"github.com/launchboxio/agent/pkg/watcher"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/clientcredentials"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"strconv"
	"time"
)

var rootCmd = &cobra.Command{
	Use:   "agent",
	Short: "LaunchboxHQ Agent",
	Run: func(cmd *cobra.Command, args []string) {
		//url, _ := cmd.Flags().GetString("url")
		clientId := os.Getenv("CLIENT_ID")
		clientSecret := os.Getenv("CLIENT_SECRET")
		tokenUrl, _ := cmd.Flags().GetString("token-url")
		apiUrl, _ := cmd.Flags().GetString("api-url")
		streamUrl, _ := cmd.Flags().GetString("stream-url")

		clusterId, _ := cmd.Flags().GetInt("cluster-id")
		cid := strconv.Itoa(clusterId)
		channel, _ := cmd.Flags().GetString("channel")

		identifier := map[string]string{
			"channel":    channel,
			"cluster_id": cid,
		}

		ctx := context.Background()
		conf := &clientcredentials.Config{
			ClientID:     clientId,
			ClientSecret: clientSecret,
			TokenURL:     fmt.Sprintf(tokenUrl),
		}
		sdk := client.New(apiUrl, conf)

		logger := zap.New()

		kubeClient, err := loadClient()

		// Start our pinger. It just updates Launchbox that the agent is running
		go func() {
			ping := pinger.New(sdk, logger)
			_ = ping.Init()
			ping.Start(clusterId, time.Second*5)
		}()

		//Start our watcher process. Whenever a project CRDs status is
		//updated, it will post the data to Launchbox
		go func() {
			watch := watcher.New(kubeClient, sdk)
			if err := watch.Run(schema.GroupVersionResource{
				Resource: "project",
				Group:    "core.launchboxhq.io",
				Version:  "v1alpha1",
			}); err != nil {
				log.Fatal(err)
			}
		}()

		//handler := events.New(logger)
		token, err := conf.Token(ctx)
		if err != nil {
			log.Fatal(err)
		}
		stream, _ := action_cable.New(streamUrl, http.Header{
			"Authorization": []string{"Bearer " + token.AccessToken},
		})

		sub := &action_cable.Subscription{
			Identifier: identifier,
		}

		stream.Subscribe(sub)

		if err := stream.Connect(ctx); err != nil {
			log.Fatal(err)
		}
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().String("token-url", "https://launchboxhq.io/oauth/token", "Authentication URL for getting access tokens")
	rootCmd.Flags().String("api-url", "https://launchboxhq.io/api/v1/", "Api Endpoint for launchbox")
	rootCmd.Flags().String("stream-url", "https://launchboxhq.io/cable", "Launchbox websocket endpoint")
	rootCmd.Flags().Int("cluster-id", 0, "Cluster ID")
	rootCmd.Flags().String("channel", "ClusterChannel", "Stream channel to subscribe to")
}

func loadClient() (*dynamic.DynamicClient, error) {
	if os.Getenv("KUBERNETES_HOST") != "" {
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		return dynamic.NewForConfig(config)
	}

	var kubeconfig *string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = flag.String("kubeconfig", filepath.Join(home, ".kube", "config"), "(optional) absolute path to the kubeconfig file")
	} else {
		kubeconfig = flag.String("kubeconfig", "", "absolute path to the kubeconfig file")
	}
	flag.Parse()
	config, err := clientcmd.BuildConfigFromFlags("", *kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	return dynamic.NewForConfig(config)
}
