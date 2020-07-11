package main

import (
	"time"
	"log"
)

const SUBSCRIPTION_DURATION = time.Minute * 5
const SUBSCRIPTION_RENEWAL = SUBSCRIPTION_DURATION - (time.Second * 5)

type Camera struct {
	Topic string
	IP string
	SubscriptionTimer *time.Timer
}

func NewCamera(ip string, topic string) *Camera {
	return &Camera{
		IP: ip,
		Topic: topic,
	}
}

func (c *Camera) Subscribe() {
	logDebug("Subscribing to", c.IP)

	cameraInit(c.Topic, c.IP)
	c.SubscriptionTimer = time.NewTimer(SUBSCRIPTION_RENEWAL)
	go c.handleSubscriptionRenewal()
}

func (c *Camera) handleSubscriptionRenewal() {
	for {
		<- c.SubscriptionTimer.C
		expiration := time.Now().Add(SUBSCRIPTION_DURATION)
		logDebug("Renewing subscription to", c.IP)
		err := sendSubscription(c.IP, expiration)
		if err != nil {
			log.Printf("[%s] %+v", c.IP, err)
		}
		c.SubscriptionTimer.Reset(SUBSCRIPTION_RENEWAL)
	}
}

func (c *Camera) Unsubscribe() {
	logDebug("Unsubscribing from", c.IP)

	if !c.SubscriptionTimer.Stop() {
		<- c.SubscriptionTimer.C
	}
	err := unsubscribe(c.IP)
	if err != nil {
		log.Printf("[%s] %+v", c.IP, err)
	}
	cameraRemove(c.Topic)
}

func (c *Camera) PostEvent(motion bool) {
	if motion {
		go motionStart(c.Topic)
	} else {
		go motionStop(c.Topic)
	}
}
