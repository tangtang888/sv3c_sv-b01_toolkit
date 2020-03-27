package main

import (
	"time"
	"net/http"
	"context"
	"strings"
	"errors"
	"bytes"
	"strconv"
)

var server *http.Server
var mux *http.ServeMux

var SubscribeError = errors.New("Could not subscribe to events.")
var UnsubscribeError = errors.New("Could not unsubscribe.")

func startServer(port uint) {
	mux = http.NewServeMux()
	mux.HandleFunc("/events", handleEvent)

	server = &http.Server{
		Addr: ":" + strconv.FormatUint(uint64(port), 10),
		Handler: mux,
		ReadTimeout: time.Second * 60,
	}

	if err := server.ListenAndServe(); err != http.ErrServerClosed {
		log_Fatalf("HTTP server error: %+v", err)
	}

	http.DefaultClient.Timeout = time.Second * 15
}

func stopServer() {
	if err := server.Shutdown(context.Background()); err != nil {
		log_Fatalf("HTTP shutdown error: %+v", err)
	}
}

var CAMERA_MOTION_TOPIC = []byte("tns1:VideoSoure/MotionAlarm")
var CAMERA_MOTION_START = []byte(`<tt:Data><tt:SimpleItem Name="State" Value="true"/></tt:Data>`)
var CAMERA_MOTION_END = []byte(`<tt:Data><tt:SimpleItem Name="State" Value="false"/></tt:Data>`)
func handleEvent(w http.ResponseWriter, r *http.Request) {
	defer r.Body.Close()

	ip := strings.Split(r.RemoteAddr, ":")[0]
	camera := findCamera(ip)
	if camera == nil {
		log_Infof("Received event from %v but it's not a registered camera.", ip)
		return
	}

	buf := new(bytes.Buffer)
	buf.ReadFrom(r.Body)
	if !bytes.Contains(buf.Bytes(), CAMERA_MOTION_TOPIC) {
		return
	}

	if bytes.Contains(buf.Bytes(), CAMERA_MOTION_START) {
		camera.PostEvent(true)
	} else if bytes.Contains(buf.Bytes(), CAMERA_MOTION_END) {
		camera.PostEvent(false)
	} else {
		log_Debugf("[%s] Unknown event message.", ip)
	}
}

func sendSubscription(cameraIP string, expiration time.Time) error {
	body := strings.NewReader(renderSubscribeXML("subscribe", callbackURL, expiration))
	res, err := http.Post("http://" + cameraIP + "/onvif/events", "application/soap+xml", body)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log_Errorf("[%s] Could not open subscription", cameraIP)
		return SubscribeError
	}
	return nil
}

func unsubscribe(cameraIP string) error {
	body := strings.NewReader(renderUnsubscribeXML("unsubscribe"))
	res, err := http.Post("http://" + cameraIP + "/onvif/events", "application/soap+xml", body)
	if err != nil {
		return err
	}

	defer res.Body.Close()
	if res.StatusCode != 200 {
		return UnsubscribeError
	}
	return nil
}

