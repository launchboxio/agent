package agent

import (
	"encoding/json"
	"fmt"
	"github.com/centrifugal/centrifuge-go"
	"github.com/gorilla/websocket"
	"k8s.io/client-go/kubernetes"
	"k8s.io/metrics/pkg/client/clientset/versioned"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/signal"
	"time"
)

type Agent struct {
	Url           string
	Token         string
	Client        *kubernetes.Clientset
	MetricsClient *versioned.Clientset
	webhookClient *centrifuge.Client
}

//
//func (a *Agent) initializeClient() error {
//
//	//client := centrifuge.NewJsonClient(a.Url, centrifuge.Config{Token: a.Token})
//	//client.OnConnecting(func(e centrifuge.ConnectingEvent) {
//	//	log.Printf("Connecting - %d (%s)", e.Code, e.Reason)
//	//})
//	//client.OnConnected(func(e centrifuge.ConnectedEvent) {
//	//	log.Printf("Connected with ID %s", e.ClientID)
//	//})
//	//client.OnDisconnected(func(e centrifuge.DisconnectedEvent) {
//	//	log.Printf("Disconnected: %d (%s)", e.Code, e.Reason)
//	//})
//	//client.OnError(func(e centrifuge.ErrorEvent) {
//	//	log.Printf("Error: %s", e.Error.Error())
//	//})
//	//client.OnSubscribed(func(e centrifuge.ServerSubscribedEvent) {
//	//	log.Printf("Subscribed to server-side channel %s: (was recovering: %v, recovered: %v)", e.Channel, e.WasRecovering, e.Recovered)
//	//})
//	//client.OnSubscribing(func(e centrifuge.ServerSubscribingEvent) {
//	//	log.Printf("Subscribing to server-side channel %s", e.Channel)
//	//})
//	//client.OnUnsubscribed(func(e centrifuge.ServerUnsubscribedEvent) {
//	//	log.Printf("Unsubscribed from server-side channel %s", e.Channel)
//	//})
//	//
//	//client.OnPublication(func(e centrifuge.ServerPublicationEvent) {
//	//	log.Printf("Publication from server-side channel %s: %s (offset %d)", e.Channel, e.Data, e.Offset)
//	//})
//	//client.OnJoin(func(e centrifuge.ServerJoinEvent) {
//	//	log.Printf("Join to server-side channel %s: %s (%s)", e.Channel, e.User, e.Client)
//	//})
//	//client.OnLeave(func(e centrifuge.ServerLeaveEvent) {
//	//	log.Printf("Leave from server-side channel %s: %s (%s)", e.Channel, e.User, e.Client)
//	//})
//	//
//	//client.OnMessage(func(e centrifuge.MessageEvent) {
//	//	//err := handlers.ProcessEvent(e)
//	//	//if err != nil {
//	//	//	log.Println(err)
//	//	//	// Broadcast the error back to HQ
//	//	//}
//	//})
//	//
//	//err := client.Connect()
//	//if err != nil {
//	//	return err
//	//}
//	//a.webhookClient = client
//	//return nil
//}

func (a *Agent) Run() error {
	//var wg sync.WaitGroup

	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: "api.lvh.me:3000", Path: "/cable"}
	fmt.Println(u.String())
	c, _, err := websocket.DefaultDialer.Dial(u.String(), http.Header{
		"Authorization": []string{fmt.Sprintf("Bearer %s", a.Token)},
	})
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()
	done := make(chan struct{})

	subscriptionMessage := struct {
		Command    string `json:"command"`
		Identifier string `json:"identifier"`
	}{
		Command:    "subscribe",
		Identifier: "{\"channel\":\"ClusterChannel\"}",
	}
	sub, err := json.Marshal(subscriptionMessage)
	go func() {
		fmt.Println("Subscribing to channel")
		err := c.WriteMessage(websocket.TextMessage, sub)
		if err != nil {
			log.Println("write:", err)
			return
		}
		defer close(done)
		for {
			messageType, message, err := c.ReadMessage()
			if err != nil {
				log.Println("read:", err)
				return
			}
			log.Printf(string(messageType))
			log.Printf("recv: %s", message)
		}
	}()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-done:
			return nil
		case <-interrupt:
			log.Println("interrupt")

			// Cleanly close the connection by sending a close message and then
			// waiting (with timeout) for the server to close the connection.
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
				return err
			}
			select {
			case <-done:
			case <-time.After(time.Second):
			}
			return nil
		}
	}

	//wg.Add(2)
	//go func() {
	//	err := a.Handler(&wg)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//}()
	//
	//go func() {
	//	err := a.Metrics(&wg)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//}()
	//
	//wg.Wait()
	return nil
}

type MetricChannel struct {
	Subscription *centrifuge.Subscription
	isSubscribed bool
}

//
//func (a *Agent) Handler(wg *sync.WaitGroup) error {
//	defer wg.Done()
//
//	handler := handlers.New(a.Client)
//
//	sub, err := a.webhookClient.NewSubscription(a.Channel, centrifuge.SubscriptionConfig{Recoverable: true, JoinLeave: true})
//	if err != nil {
//		log.Fatalln(err)
//	}
//
//	sub.OnSubscribing(func(e centrifuge.SubscribingEvent) {
//		log.Printf("Subscribing on channel %s - %d (%s)", sub.Channel, e.Code, e.Reason)
//	})
//	sub.OnSubscribed(func(e centrifuge.SubscribedEvent) {
//		log.Printf("Subscribed on channel %s, (was recovering: %v, recovered: %v)", sub.Channel, e.WasRecovering, e.Recovered)
//	})
//	sub.OnUnsubscribed(func(e centrifuge.UnsubscribedEvent) {
//		log.Printf("Unsubscribed from channel %s - %d (%s)", sub.Channel, e.Code, e.Reason)
//	})
//
//	sub.OnError(func(e centrifuge.SubscriptionErrorEvent) {
//		log.Printf("Subscription error %s: %s", sub.Channel, e.Error)
//	})
//
//	sub.OnPublication(func(e centrifuge.PublicationEvent) {
//		err := handler.ProcessEvent(e)
//		fmt.Println(string(e.Data))
//		if err != nil {
//			log.Println(err)
//			// Broadcast the error back to HQ
//		}
//		fmt.Println(string(e.Data))
//	})
//	sub.OnJoin(func(e centrifuge.JoinEvent) {
//		log.Printf("Someone joined %s: user id %s, client id %s", sub.Channel, e.User, e.Client)
//	})
//	sub.OnLeave(func(e centrifuge.LeaveEvent) {
//		log.Printf("Someone left %s: user id %s, client id %s", sub.Channel, e.User, e.Client)
//	})
//
//	err = sub.Subscribe()
//	if err != nil {
//		log.Fatalln(err)
//	}
//
//	select {}
//}

type MetricMessage struct {
	Container string `json:"container,omitempty"`
	Cpu       string `json:"cpu,omitempty"`
	Memory    string `json:"memory,omitempty"`
}

//
//func (a *Agent) Metrics(wg *sync.WaitGroup) error {
//	defer wg.Done()
//
//	fmt.Println("Polling metrics")
//
//	subs := map[string]*MetricChannel{}
//
//	for {
//		podMetrics, err := a.MetricsClient.MetricsV1beta1().PodMetricses(metav1.NamespaceAll).List(context.TODO(), metav1.ListOptions{
//			LabelSelector: "app=vcluster",
//		})
//		if err != nil {
//			log.Println(err)
//		}
//
//		for _, metric := range podMetrics.Items {
//			projectId := metric.Labels["release"]
//			// Ensure we're subscribed to the projects channel
//			sub, ok := subs[projectId]
//			if !ok {
//				//fmt.Println("Initializing new Project subscription")
//				sub := &MetricChannel{}
//				subSub, err := a.webhookClient.NewSubscription(fmt.Sprintf("projects:%s", projectId), centrifuge.SubscriptionConfig{Recoverable: true, JoinLeave: true})
//				if err != nil {
//					log.Fatalln(err)
//				}
//
//				subSub.OnSubscribed(func(e centrifuge.SubscribedEvent) {
//					fmt.Println("Subscribed..")
//					sub.isSubscribed = true
//				})
//				subSub.OnUnsubscribed(func(e centrifuge.UnsubscribedEvent) {
//					fmt.Println("Unsubscribed..")
//					sub.isSubscribed = false
//				})
//				err = subSub.Subscribe()
//				if err != nil {
//					log.Println(err)
//					continue
//				}
//				sub.Subscription = subSub
//				sub.isSubscribed = false
//				subs[projectId] = sub
//			}
//
//			//fmt.Println("Project subscription already exists")
//			//fmt.Println(sub)
//			for _, container := range metric.Containers {
//				if container.Name != "vcluster" {
//					continue
//				}
//				req := &MetricMessage{
//					Cpu:    container.Usage.Cpu().String(),
//					Memory: container.Usage.Memory().String(),
//				}
//				bytes, err := json.Marshal(req)
//				if err != nil {
//					continue
//				}
//
//				if sub != nil && sub.isSubscribed {
//					//fmt.Println("Publishing metrics usage")
//					sub.Subscription.Publish(context.TODO(), bytes)
//				}
//			}
//		}
//		time.Sleep(time.Second * 5)
//	}
//}
