package cmd

import (
	"fmt"
	action_cable "github.com/launchboxio/action-cable"
	"github.com/launchboxio/agent/pkg/client"
	"github.com/launchboxio/agent/pkg/events"
	"github.com/launchboxio/agent/pkg/pinger"
	"github.com/launchboxio/agent/pkg/watcher"
	"github.com/launchboxio/operator/api/v1alpha1"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/clientcredentials"
	"k8s.io/apimachinery/pkg/runtime/schema"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"net/http"
	"os"
	"path/filepath"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
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
			TokenURL:     tokenUrl,
		}
		sdk := client.New(apiUrl, conf)

		logger := zap.New()

		// Start our pinger. It just updates Launchbox that the agent is running
		go func() {
			ping := pinger.New(sdk, logger)
			_ = ping.Init()
			ping.Start(clusterId, time.Second*5)
		}()

		//Start our watcher process. Whenever a project CRDs status is
		//updated, it will post the data to Launchbox
		go func() {
			kubeClient, err := loadDynamicClient()
			if err != nil {
				logger.Error(err, "Failed to build watcher client")
				os.Exit(1)
			}
			watch := watcher.New(kubeClient, sdk, logger)
			if err := watch.Run(schema.GroupVersionResource{
				Resource: "projects",
				Group:    "core.launchboxhq.io",
				Version:  "v1alpha1",
			}); err != nil {
				logger.Error(err, "Failed to start watcher")
				os.Exit(1)
			}
		}()

		token, err := conf.Token(ctx)
		if err != nil {
			logger.Error(err, "Failed to get authentication token")
			os.Exit(1)
		}
		stream, _ := action_cable.New(streamUrl, http.Header{
			"Authorization": []string{"Bearer " + token.AccessToken},
		})
		stream.OnMessage = func(message []byte) {
			logger.Info(string(message))
		}
		stream.ErrorHandler = func(err error) {
			logger.Error(err, "Error on stream")
		}

		client, err := loadClient()
		if err != nil {
			logger.Error(err, "Failed to get build runtime client")
			os.Exit(1)
		}

		handler := events.New(logger, client)
		handler.RegisterSubscriptions(stream, identifier)

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

func loadDynamicClient() (*dynamic.DynamicClient, error) {
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		config, err := rest.InClusterConfig()
		if err != nil {
			return nil, err
		}
		return dynamic.NewForConfig(config)
	}

	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	} else {
		kubeconfig = os.Getenv("KUBECONFIG")
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		panic(err.Error())
	}

	return dynamic.NewForConfig(config)
}

func loadClient() (runtimeclient.Client, error) {
	kubeClient, err := runtimeclient.New(config.GetConfigOrDie(), runtimeclient.Options{})
	if err != nil {
		return nil, err
	}
	utilruntime.Must(v1alpha1.AddToScheme(kubeClient.Scheme()))
	return kubeClient, nil
}
