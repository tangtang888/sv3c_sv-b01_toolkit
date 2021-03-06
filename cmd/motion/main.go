package main

import (
	"flag"
	"strconv"
	"strings"
	"os"
	"os/signal"
	"net"
	"errors"
	"log"
	"time"
	"fmt"

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

var localIP string
var port uint
var cameraIPs flagArray
var cameraTopics flagArray
var callbackURL string
var mqttBroker string
var debugLogEnabled bool

func init() {
	flag.BoolVar(&debugLogEnabled, "debug", false, "Debug logging enabled.")
	flag.UintVar(&port, "port", 8080, "Port to bind to.")
	flag.StringVar(&mqttBroker, "broker", "127.0.0.1:1883", "MQTT broker with port. [127.0.0.1:1883]")
	flag.Var(&cameraIPs, "camera", "Camera IP for ONVIF over HTTP (multiple allowed). [192.168.1.100]")
	flag.Var(&cameraTopics, "topic", "Camera topic (multiple allowed). [home/garage/camera]")
	flag.Parse()
	
	if len(cameraIPs) == 0 {
		log.Fatal("No cameras specified.")
	}
	for _, ip := range cameraIPs {
		if net.ParseIP(ip) == nil {
			log.Fatal("Invalid camera address:", ip)
		}
	}
	if len(cameraTopics) != len(cameraIPs) {
		log.Fatal("Mismatched number of camera IPs and topics.")
	}

	var err error
	localIP, err = getExternalIP()
	if err != nil {
		log.Fatal(err)
	}
 
	callbackURL = "http://" + localIP + ":" + strconv.FormatUint(uint64(port), 10) + "/events"
}

var cameras []*Camera

func findCamera(ip string) *Camera {
	for _, camera := range cameras {
		if camera.IP == ip {
			return camera
		}
	}
	return nil
}

var client mqtt.Client

func main() {
	errLog := log.New(os.Stderr, "", 0)
	mqtt.ERROR = errLog
	opts := mqtt.NewClientOptions().AddBroker("tcp://" + mqttBroker).SetClientID("sv3c_b01_onvif_" + strconv.Itoa(os.Getpid()))
	opts.SetKeepAlive(time.Second * 5)
	opts.SetPingTimeout(time.Second * 1)
	opts.SetConnectTimeout(time.Second * 5)
	opts.SetAutoReconnect(true)
	opts.SetCleanSession(true)

	client = mqtt.NewClient(opts)
	if token := client.Connect(); token.Wait() && token.Error() != nil {
		log.Fatal(token.Error())
	}

	cameras = make([]*Camera, 0, len(cameraIPs))
	for i, ip := range cameraIPs {
		cam := NewCamera(ip, cameraTopics[i])
		cameras = append(cameras, cam)
		go cam.Subscribe()
	}

	go startServer(port)
	
	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<- sigint

	for _, camera := range cameras {
		camera.Unsubscribe()
	}

	client.Disconnect(1000)
	
	stopServer()
}

func logDebug(v ...interface{}) {
	if debugLogEnabled {
		log.Println(v)
	}
}

func getExternalIP() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		return "", err
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return "", err
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String(), nil
		}
	}
	return "", errors.New("Cannot determine local IP.")
}

func cameraInit(topic string, ip string) {
	t := client.Publish(topic + "/ip", 1, true, ip)
	t.Wait()
	if t.Error() != nil {
		log.Println(t.Error())
	}
}

func cameraRemove(topic string) {
	t1 := client.Publish(topic + "/motion", 1, true, nil)
	t2 := client.Publish(topic + "/lastMotion", 1, true, nil)
	t3 := client.Publish(topic + "/ip", 1, true, nil)
	t4 := client.Publish(topic + "/recording", 1, true, nil)
	t1.Wait()
	t2.Wait()
	t3.Wait()
	t4.Wait()
	if t1.Error() != nil {
		log.Println(t1.Error())
	}
	if t2.Error() != nil {
		log.Println(t2.Error())
	}
	if t3.Error() != nil {
		log.Println(t3.Error())
	}
	if t4.Error() != nil {
		log.Println(t4.Error())
	}
}

func motionStart(topic string) {
	t1 := client.Publish(topic + "/motion", 1, false, "true")
	t2 := client.Publish(topic + "/lastMotion", 1, true, fmt.Sprint(time.Now().Unix()))
	t1.Wait()
	t2.Wait()

	if t1.Error() != nil {
		log.Println(t1.Error())
	}
	if t2.Error() != nil {
		log.Println(t2.Error())
	}
}

func motionStop(topic string) {
	t := client.Publish(topic + "/motion", 1, false, "false")
	t.Wait()
	if t.Error() != nil {
		log.Println(t.Error())
	}
}