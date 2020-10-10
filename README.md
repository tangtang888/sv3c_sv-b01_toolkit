# SV3C SV-B01 Toolkit

## `motion`

Subscribes to motion events from the camera and publishes results to MQTT.

## `record`

Subscribes to MQTT topics and records streams when motion is detected.

## Hardware

[SV3C SV-B01-1080P-POE Camera](https://www.amazon.com/gp/product/B01G1U4MVA) with 2018-09-07 firmware.

### Known URLs

- `rtsp://[ip]:554/stream0` (main stream)
- `rtsp://[ip]:554/stream1` (sub stream)
- `http://[ip]/cgi-bin/snapshot.cgi?stream=0` (full-res JPG, changing stream doesn't seem to make a difference)


## Building

```
go build ./cmd/motion -o build/motion
go build ./cmd/record -o build/record
```
(ensure `GOPATH` is set when using nix)

## Notes

### Proxy

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

