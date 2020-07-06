package main

import (
	"time"
	"log"
	"os"
	"os/exec"
	"path"
	"strings"
)

const RECORDING_TIME_FORMAT = "2006-01-02_15.04.05"

type Camera struct {
	Topic string
	Name string
	IP string
	RecordingStopTimer *time.Timer
	PostMotionRecordDuration time.Duration
	LastMotionEvent time.Time
	ffmpegCmd *exec.Cmd
}

func NewCamera(ip string, topic string, recordDuration time.Duration) *Camera {
	c := &Camera{
		Topic: topic,
		Name: strings.ReplaceAll(topic, "/", "-"),
		IP: ip,
		PostMotionRecordDuration: recordDuration,
		RecordingStopTimer: time.NewTimer(time.Hour),
	}

	go c.handleTimerExpire()

	return c
}

func (c *Camera) handleTimerExpire() {
	for {
		<- c.RecordingStopTimer.C
		if c.ffmpegCmd != nil {
			c.StopRecording()
		}
	}
}

func (c *Camera) StopRecording() {
	logDebug("Stopping recording for", c.Topic)
	c.ffmpegCmd.Process.Signal(os.Interrupt)
	c.ffmpegCmd = nil
}

func (c *Camera) StartMotion() {
	c.LastMotionEvent = time.Now()
	logDebug("Disabling timer for", c.Topic)
	c.RecordingStopTimer.Stop()

	if c.ffmpegCmd != nil {
		logDebug("Recording already in progress on", c.Topic)
		return
	}

	logDebug("Starting recording for", c.Topic)
	output := path.Join(outputPath, time.Now().Format(RECORDING_TIME_FORMAT) + "_" + c.Name + ".mp4")
	title := time.Now().Format("2006-01-02 15:04:05") + " - " + c.Name
	c.ffmpegCmd = recordCmd("rtsp://" + c.IP + ":554/stream0", output, title)
	c.ffmpegCmd.Dir = outputPath
	if err := c.ffmpegCmd.Start(); err != nil {
		log.Fatalf("[%s | %s] %+v", c.Topic, c.IP, err)
	}
}

func (c *Camera) StopMotion() {
	logDebug("Motion finished. Starting timer for", c.Topic)
	c.RecordingStopTimer.Stop()
	c.RecordingStopTimer.Reset(c.PostMotionRecordDuration)
}

func recordCmd(streamURL string, filename string, title string) *exec.Cmd {
	return exec.Command("ffmpeg", "-i", streamURL, "-vcodec", "copy", "-metadata", "title=" + title, filename)
}