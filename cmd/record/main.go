package main

import (
	"flag"
	"os"
	"os/signal"
	"log"
	"time"
	"strings"

	mqtt "github.com/eclipse/paho.mqtt.golang"
)

type flagArray []string

func (i *flagArray) String() string {
	return ""
}

func (i *flagArray) Set(value string) error {
	*i = append(*i, strings.TrimSpace(value))
	return nil
}

var cameraTopics flagArray
var mqttBroker string
var debugLogEnabled bool
var outputPath string
var recordingKeepDays uint

func init() {
	flag.BoolVar(&debugLogEnabled, "debug", false, "Debug logging enabled.")
	flag.StringVar(&mqttBroker, "broker", "127.0.0.1:1883", "MQTT broker with port. [127.0.0.1:1883]")
	flag.Var(&cameraTopics, "topic", "Camera topics (mulitple allowed). [home/garage/camera]")
	flag.StringVar(&outputPath, "output", "", "Recording output path. [/srv/camera]")
	flag.UintVar(&recordingKeepDays, "saveDays", 30, "Save recordings for this many days.")
	
	flag.Parse()

	if len(cameraTopics) == 0 {
		log.Fatal("No camera topics specified.")
	}
	if outputPath == "" {
		log.Fatal("No output path specified.")
	}
}

var client mqtt.Client

func main() {
	errLog := log.New(os.Stderr, "", 0)
	mqtt.ERROR = errLog
	opts := mqtt.NewClientOptions().AddBroker("tcp://" + mqttBroker).SetClientID("sv3c_b01_record")
	opts.SetKeepAlive(time.Second * 5)
	opts.SetPingTimeout(time.Second * 1)
	opts.SetConnectTimeout(time.Second * 5)
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(true)

	client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	for _, topic := range cameraTopics {
		if token := client.Subscribe(topic + "/+", 1, handleCameraMessage); token.Wait() && token.Error() != nil {
			log.Fatal(token.Error())
		}
	}

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<- sigint

	for _, topic := range cameraTopics {
		if token := client.Unsubscribe(topic + "/+"); token.Wait() && token.Error() != nil {
			log.Fatal(token.Error())
		}
	}
}

var cameras map[string]*Camera = make(map[string]*Camera)

func handleCameraMessage(client mqtt.Client, msg mqtt.Message) {
	primaryTopic := msg.Topic()[:strings.LastIndex(msg.Topic(), "/")]
	
	if strings.HasSuffix(msg.Topic(), "/ip") {
		addCamera(primaryTopic, string(msg.Payload()))
	} else if strings.HasSuffix(msg.Topic(), "/motion") {
		if string(msg.Payload()) == "true" {
			cameras[primaryTopic].StartMotion()
		} else if string(msg.Payload()) == "false" {
			cameras[primaryTopic].StopMotion()
		}
	}
}

func addCamera(topic string, ip string) {
	if _, ok := cameras[topic]; !ok {
		logDebug("Found camera", topic, ip)
		cameras[topic] = NewCamera(ip, topic, time.Second * 5)
	}
}

func startPurgeTask() {
	ticker := time.NewTicker(time.Hour * 24)
	defer ticker.Stop()
	for _ = range ticker.C {
		purge()
	}
}

func logDebug(v ...interface{}) {
	if debugLogEnabled {
		log.Println(v)
	}
}
