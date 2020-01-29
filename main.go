package main

import (
	"flag"
	"log"
	"strconv"
	"strings"
	"time"
	"os"
	"os/signal"
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
var callbackURL string
var ffmpegPath string
var outputPath string

func init() {
	flag.StringVar(&outputPath, "outputPath", "", "Output directory for recordings.")
	flag.StringVar(&localIP, "localIP", "0.0.0.0", "IP of this machine, where cameras will make event callbacks.")
	flag.StringVar(&ffmpegPath, "ffmpeg", "ffmpeg", "ffmpeg path")
	flag.UintVar(&port, "port", 8080, "Port to bind to.")
	flag.Var(&cameraIPs, "camera", "Camera IP and port to subscribe to (multiple allowed).")
	flag.Parse()

	if localIP == "0.0.0.0" {
		log.Fatal("Local IP not specified.")
	}
	if len(cameraIPs) == 0 {
		log.Fatal("No cameras specified.")
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

func main() {
	cameras = make([]*Camera, 0, len(cameraIPs))
	for _, ip := range cameraIPs {
		cam := NewCamera(ip, time.Second * 5)
		cam.Subscribe()
		cameras = append(cameras, cam)
	}

	go startServer(port)

	sigint := make(chan os.Signal, 1)
	signal.Notify(sigint, os.Interrupt)
	<- sigint

	for _, camera := range cameras {
		camera.Stop()
	}
	
	stopServer()
}
