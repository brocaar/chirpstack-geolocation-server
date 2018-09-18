---
title: Requirements
menu:
    main:
        parent: install
        weight: 1
description: Instructions how to setup the LoRa Geo Server requirements.
---

# Requirement

## Gateway

When using geolocation, you need a gateway which provides geolocation
capabilities. These gateways provide accurate timestamps, which is a
requirement when doing geolocation based on time-difference of arrival.

### Tested gateways

The following gateways have been tested:

* [Kerlink iBTS](https://www.kerlink.com/product/wirnet-ibts/)


## Decryption key

Gateways implementing the Semtech v2 reference design will encrypt the
fine-timestamp. Before being able to use this timestamp for geolocation,
you must therefore request a decryption key. Contact your gateway vendor
or Semtech for more information.

One you have this decryption key, you must set this in the gateway configuration
so that LoRa Server will decrypt this before sending to LoRa Geo Server.
Please refer to [LoRa App Server Gateway Management](/use/gateways/) for
more information.
