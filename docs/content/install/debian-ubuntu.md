---
title: Debian / Ubuntu
menu:
    main:
        parent: install
        weight: 2
description: Instructions how to install LoRa Geo Server on a Debian or Ubuntu based Linux installation.
---

# Debian / Ubuntu installation

These steps have been tested on:

* Ubuntu 16.04 LTS
* Ubuntu 18.04 LTS
* Debian 9 (Stretch)

## LoRa Server Debian repository

The LoRa Server project provides pre-compiled binaries packaged as Debian (.deb)
packages. In order to activate this repository, execute the following
commands:

{{<highlight bash>}}
sudo apt-key adv --keyserver keyserver.ubuntu.com --recv-keys 1CE2AFD36DBCCA00

sudo echo "deb https://artifacts.loraserver.io/packages/2.x/deb stable main" | sudo tee /etc/apt/sources.list.d/loraserver.list
sudo apt-get update
{{< /highlight >}}

## Install LoRa Geo Server

In order to install LoRa Geo Server, execute the following command:

{{<highlight bash>}}
sudo apt-get install lora-geo-server
{{< /highlight >}}

After installation, modify the configuration file which is located at
`/etc/lora-geo-server/lora-geo-server.toml`.

Settings you probably want to set / change:

* `geo_server.backend.collos.subscription_key`

## Starting LoRa Geo Server

How you need to (re)start and stop LoRa Geo Server depends on if your
distribution uses init.d or systemd.

### init.d

{{<highlight bash>}}
sudo /etc/init.d/lora-geo-server [start|stop|restart|status]
{{< /highlight >}}

### systemd

{{<highlight bash>}}
sudo systemctl [start|stop|restart|status] lora-geo-server
{{< /highlight >}}

## LoRa Geo Server log output

Now you've setup LoRa Geo Server, it is a good time to verify that it
is actually up-and-running. This can be done by looking at the LoRa Geo Server
log output.

Like the previous step, which command you need to use for viewing the
log output depends on if your distribution uses init.d or systemd.

### init.d

All logs are written to `/var/log/lora-geo-server/lora-geo-server.log`.
To view and follow this logfile:

{{<highlight bash>}}
tail -f /var/log/lora-geo-server/lora-geo-server.log
{{< /highlight >}}

### systemd

{{<highlight bash>}}
journalctl -u lora-geo-server -f -n 50
{{< /highlight >}}
