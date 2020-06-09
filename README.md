# Hardware

[SV3C SV-B01-1080P-POE Camera](https://www.amazon.com/gp/product/B01G1U4MVA) with 2018-09-07 firmware.

# Environment

- `ffmpeg` installed

# Usage

`onvif_record -outputPath /srv/camera -localIP 192.168.1.2 -camera 192.168.1.10/cam1 -camera 192.168.1.11/cam2`

# Misc.

Compiling for the Synology DS218j, use `GOOS=linux GOARCH=arm go build`.