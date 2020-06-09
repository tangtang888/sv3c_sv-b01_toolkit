package main

import (
	"time"
	"os"
	"os/exec"
	"path"
)

const SUBSCRIPTION_DURATION = time.Minute * 5
const SUBSCRIPTION_RENEWAL = SUBSCRIPTION_DURATION - (time.Second * 5)
const RECORDING_TIME_FORMAT = "2006-01-02_15.04.05"

type Camera struct {
	Name string
	IP string
	Subscribed bool
	SubscriptionExpiration time.Time
	SubscriptionTimer *time.Timer
	RecordingStopTimer *time.Timer
	PostMotionRecordDuration time.Duration
	LastMotionEvent time.Time
	ffmpegCmd *exec.Cmd
}

func NewCamera(ip string, name string, minDuration time.Duration) *Camera {
	return &Camera{
		IP: ip,
		Name: name,
		PostMotionRecordDuration: minDuration,
	}
}

func (c *Camera) Subscribe() {
	log_Debugf("[%s] Subscribing...", c.IP)

	expiration := time.Now().Add(SUBSCRIPTION_DURATION)
	err := sendSubscription(c.IP, expiration)
	if err != nil {
		log_Errorf("[%s] %+v", c.IP, err)
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
		log_Debugf("[%s] Renewing subscription", c.IP)
		err := sendSubscription(c.IP, expiration)
		if err != nil {
			log_Errorf("[%s] %+v", c.IP, err)
		}
		c.SubscriptionTimer.Reset(SUBSCRIPTION_RENEWAL)
	}
}

func (c *Camera) Unsubscribe() {
	log_Debugf("[%s] Unsubscribing", c.IP)

	if !c.SubscriptionTimer.Stop() {
		<- c.SubscriptionTimer.C
	}
	err := unsubscribe(c.IP)
	if err != nil {
		log_Errorf("[%s] %+v", c.IP, err)
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
	log_Debugf("[%s] Stopping recording", c.IP)
	c.ffmpegCmd.Process.Signal(os.Interrupt)
	c.ffmpegCmd = nil
}

func (c *Camera) PostEvent(motion bool) {
	if motion {
		c.LastMotionEvent = time.Now()
		log_Debugf("[%s] Disabling timer", c.IP)
		c.RecordingStopTimer.Stop()

		if c.ffmpegCmd != nil {
			log_Debugf("[%s] Recording already in progress", c.IP)
			return
		}
		
		output := path.Join(outputPath, c.Name + "_" + time.Now().Format(RECORDING_TIME_FORMAT) + ".mp4")
		log_Debugf("[%s] Starting recording at %s", c.IP, output)
		c.ffmpegCmd = recordCmd("rtsp://" + c.IP + ":554/stream0", output)
		c.ffmpegCmd.Dir = outputPath
		if err := c.ffmpegCmd.Start(); err != nil {
			log_Fatalf("[%s] %+v", c.IP, err)
		}
	} else {
		if c.ffmpegCmd == nil {
			return
		}
		log_Debugf("[%s] Motion finished. Starting timer.", c.IP)
		c.RecordingStopTimer.Stop()
		c.RecordingStopTimer.Reset(c.PostMotionRecordDuration)
	}
}

func recordCmd(streamURL string, filename string) *exec.Cmd {
	return exec.Command("ffmpeg", "-i", streamURL, "-vcodec", "copy", filename)
}