package cmd

import (
	"fmt"
	crossplanepkgv1 "github.com/crossplane/crossplane/apis/pkg/v1"
	action_cable "github.com/launchboxio/action-cable"
	"github.com/launchboxio/agent/pkg/evaluator"
	"github.com/launchboxio/agent/pkg/events"
	"github.com/launchboxio/agent/pkg/pinger"
	"github.com/launchboxio/agent/pkg/server"
	"github.com/launchboxio/agent/pkg/watcher"
	launchbox "github.com/launchboxio/launchbox-go-sdk/config"
	"github.com/launchboxio/launchbox-go-sdk/service/cluster"
	"github.com/launchboxio/operator/api/v1alpha1"
	"github.com/spf13/cobra"
	"golang.org/x/net/context"
	"golang.org/x/oauth2/clientcredentials"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/util/homedir"
	"log"
	"net/http"
	"os"
	"path/filepath"
	runtimeclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/config"

	//"sigs.k8s.io/controller-runtime/pkg/client/config"
	"sigs.k8s.io/controller-runtime/pkg/log/zap"
	"strconv"
	"time"
)

var version string

var rootCmd = &cobra.Command{
	Use:   "agent",
	Short: "LaunchboxHQ Agent",
	Run: func(cmd *cobra.Command, args []string) {
		//url, _ := cmd.Flags().GetString("url")
		streamUrl, _ := cmd.Flags().GetString("stream-url")

		clusterId, _ := cmd.Flags().GetInt("cluster-id")
		cid := strconv.Itoa(clusterId)
		channel, _ := cmd.Flags().GetString("channel")

		identifier := map[string]string{
			"channel":    channel,
			"cluster_id": cid,
		}

		ctx := context.Background()

		sdk, err := launchbox.Default()
		if err != nil {
			log.Fatal(err)
		}

		logger := zap.New()

		// Start our pinger. It just updates Launchbox that the agent is running
		go func() {
			kubeclient, err := loadKubeClient()
			if err != nil {
				log.Fatal(err)
			}
			eval := evaluator.New(kubeclient)
			evaluation, err := eval.Evaluate()
			if err != nil {
				log.Fatal(err)
			}
			ping := pinger.New(cluster.New(sdk), logger)
			_ = ping.Init(clusterId, evaluation, version)
			ping.Start(time.Second * 5)
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
			if err := watch.Run(); err != nil {
				logger.Error(err, "Failed to start watcher")
				os.Exit(1)
			}
		}()

		go func() {
			bindAddress, _ := cmd.Flags().GetString("bind-address")
			if err := server.New(bindAddress).Run(); err != nil {
				log.Fatal(err)
			}
		}()

		conf := &clientcredentials.Config{
			ClientID:     os.Getenv("LAUNCHBOX_CLIENT_ID"),
			ClientSecret: os.Getenv("LAUNCHBOX_CLIENT_SECRET"),
			TokenURL:     os.Getenv("LAUNCHBOX_TOKEN_URL"),
		}

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

		handler := events.New(logger, client, sdk)
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
	rootCmd.Flags().String("stream-url", "wss://launchboxhq.io/cable", "Launchbox websocket endpoint")
	rootCmd.Flags().Int("cluster-id", 0, "Cluster ID")
	rootCmd.Flags().String("channel", "ClusterChannel", "Stream channel to subscribe to")
	rootCmd.Flags().String("bind-address", ":8080", "Bind address for the http server")
}

func loadDynamicClient() (*dynamic.DynamicClient, error) {
	conf, err := loadConfig()
	if err != nil {
		return nil, err
	}
	return dynamic.NewForConfig(conf)
}

func loadClient() (runtimeclient.Client, error) {
	kubeClient, err := runtimeclient.New(config.GetConfigOrDie(), runtimeclient.Options{})
	if err != nil {
		return nil, err
	}
	utilruntime.Must(v1alpha1.AddToScheme(kubeClient.Scheme()))
	utilruntime.Must(crossplanepkgv1.AddToScheme(kubeClient.Scheme()))

	return kubeClient, nil
}

func loadKubeClient() (*kubernetes.Clientset, error) {
	conf, err := loadConfig()
	if err != nil {
		return nil, err
	}
	return kubernetes.NewForConfig(conf)
}

func loadConfig() (*rest.Config, error) {
	if os.Getenv("KUBERNETES_SERVICE_HOST") != "" {
		return rest.InClusterConfig()
	}

	var kubeconfig string
	if home := homedir.HomeDir(); home != "" {
		kubeconfig = filepath.Join(home, ".kube", "config")
	} else {
		kubeconfig = os.Getenv("KUBECONFIG")
	}

	return clientcmd.BuildConfigFromFlags("", kubeconfig)
}
