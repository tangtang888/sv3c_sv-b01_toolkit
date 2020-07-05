# onvif_mqtt

Subscribes to motion events from an IP camera and publishes results to MQTT.

## Hardware

[SV3C SV-B01-1080P-POE Camera](https://www.amazon.com/gp/product/B01G1U4MVA) with 2018-09-07 firmware.

## Proxy

RTSP streams can be proxied with this minimal nginx config.

```
events {}

stream {
	server {
		listen 8554;
		proxy_pass camera_one;
	}
	upstream camera_one {
		server 192.168.1.10:554;
	}
}
```