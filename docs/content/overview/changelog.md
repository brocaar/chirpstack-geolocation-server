---
title: Changelog
menu:
    main:
        parent: overview
        weight: 3
toc: false
description: Lists the changes per LoRa Geo Server release, including steps how to upgrade.
---

# Changelog

## v3.2.0

### Features

#### LoRa Cloud

This release adds support for the [LoRa Cloud Geolocation](https://www.loracloud.com/)
geolocation service.

## v3.1.0

### Features

#### Collos multi-frame

The Collos multi-frame integration makes it possible to perform geolocation
using the meta-data of multiple uplink frames to increase accuracy.

#### Prometheus metrics

Metrics can now be exposed using a [Prometheus](https://prometheus.io/) metrics endpoint.

## v3.0.0

This release bumps the major version to v3, to stay in sync with the other
LoRa Server components.

### Improvements

* Update dependencies to their latest versions.

## v2.0.1

### Bugfixes

* Fix keypair loading error (`load key-pair error: tls: found a certificate rather than a key in the PEM for the private key`).

## v2.0.0

Initial release.

## Requirements

You need LoRa Server v2.2+ for geolocation.
