package main

import (
	"time"
	"log"
	"os"
	"os/exec"
	"path"
)

const SUBSCRIPTION_DURATION = time.Minute * 10
const SUBSCRIPTION_RENEWAL = SUBSCRIPTION_DURATION - (time.Second * 30)

type Camera struct {
	IP string
	Subscribed bool
	SubscriptionExpiration time.Time
	SubscriptionTimer *time.Timer
	RecordingStopTimer *time.Timer
	PostMotionRecordDuration time.Duration
	LastMotionEvent time.Time
	ffmpegCmd *exec.Cmd
}

// subscribe
// tick for renew subscription
// unsubscribe on shutdown

// listen for events and start recording, update last motion event
// tick for stop recording


func NewCamera(ip string, minDuration time.Duration) *Camera {
	return &Camera{
		IP: ip,
		PostMotionRecordDuration: minDuration,
	}
}

func (c *Camera) Subscribe() {
	if c.Subscribed {
		log.Fatalf("[ERR] Already subscribed to %s\n", c.IP)
	}
	log.Printf("[DEBUG] Subscribing to %s...\n", c.IP)

	expiration := time.Now().Add(SUBSCRIPTION_DURATION)
	err := sendSubscription(c.IP, expiration)
	if err != nil {
		log.Fatal(err) // TODO: Retry?
	}

	c.SubscriptionExpiration = expiration
	c.Subscribed = true
	c.SubscriptionTimer = time.NewTimer(SUBSCRIPTION_RENEWAL)
	go c.handleSubscriptionRenewal()

	c.RecordingStopTimer = time.NewTimer(time.Hour)
	c.RecordingStopTimer.Stop()
	go c.handleRecordingStop()
}

func (c *Camera) handleSubscriptionRenewal() {
	for {
		<- c.SubscriptionTimer.C
		expiration := time.Now().Add(SUBSCRIPTION_DURATION)
		err := renewSubscription(c.IP, expiration)
		if err != nil {
			log.Fatal(err) // TODO: not fatal
		}
		c.SubscriptionTimer.Reset(SUBSCRIPTION_RENEWAL)
	}
}

func (c *Camera) Unsubscribe() {
	if !c.Subscribed {
		log.Fatal("[ERR] Camera not yet subscribed.\n")
	}
	log.Printf("[DEBUG] Unsubscribing from %s...\n", c.IP)

	if !c.SubscriptionTimer.Stop() {
		<- c.SubscriptionTimer.C
	}
	err := unsubscribe(c.IP)
	if err != nil {
		log.Fatal(err) // TODO: not fatal
	}

	c.Subscribed = false
}

func (c *Camera) Stop() {
	c.Unsubscribe()
	if c.ffmpegCmd != nil {
		c.StopRecording()
	}
}

func (c *Camera) handleRecordingStop() {
	for {
		<- c.RecordingStopTimer.C
		if c.ffmpegCmd != nil {
			c.StopRecording()
		}
	}
}

func (c *Camera) StopRecording() {
	log.Print("[DEBUG] Stopping recording.")
	c.ffmpegCmd.Process.Signal(os.Interrupt)
	c.ffmpegCmd = nil
}

func (c *Camera) PostEvent(motion bool) {
	if motion {
		c.LastMotionEvent = time.Now()
		log.Print("[DEBUG] Disabling timer.")
		c.RecordingStopTimer.Stop()

		if c.ffmpegCmd != nil {
			log.Print("[DEBUG] Recording already in progress...")
			return
		}
		
		log.Print("[DEBUG] Starting recording.")
		c.ffmpegCmd = recordCmd("rtsp://" + c.IP + ":554/stream0")
		c.ffmpegCmd.Dir = outputPath
		if err := c.ffmpegCmd.Start(); err != nil {
			log.Fatal(err)
		}
	} else {
		if c.ffmpegCmd == nil {
			return
		}
		log.Print("[DEBUG] Motion finished. Starting timer.")
		c.RecordingStopTimer.Stop()
		c.RecordingStopTimer.Reset(c.PostMotionRecordDuration)
	}
}

const RECORDING_TIME_FORMAT = "2006-01-02T15:04:05"

func recordCmd(streamURL string) *exec.Cmd {
	return exec.Command(ffmpegPath, "-i", streamURL, "-vcodec", "copy", path.Join(outputPath, time.Now().Format(RECORDING_TIME_FORMAT) + ".mp4"))
}