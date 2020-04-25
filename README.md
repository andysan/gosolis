[![Go Report Card](https://goreportcard.com/badge/github.com/andysan/gosolis)](https://goreportcard.com/report/github.com/andysan/gosolis)
[![CI](https://github.com/andysan/gosolis/workflows/CI/badge.svg)](https://github.com/andysan/gosolis/actions?query=workflow%3ACI)

# What is GoSolis?

GoSolis is an access library and a set of tools for the Solis range of
inverters by Ginlong. It may work with other inverters from the same
company.

The main purpose of these tools is to query the status of the
inverter. This can either be done as a "one off" or using a daemon
that continuously polls the inverter and sends status messages to an
MQTT message queue.

**WARNING:** This library can be used to change the configuration of
the inverter. Configuration changes have not been tested on a real
inverter and could potentially be very dangerous.

# Inverter connection

The control interface on the inverter uses an RS485 connection at 9600
baud. RS485 to USB adaptors are readily available, but a custom cable
is needed to interface to the inverter.

The physical connector seems to be a custom 4-pin male connector with
the following pinout:

<pre>
    ---------
   /         \
  /   1   4   \
  |           |
  \   2   3   /
   \    _    /
    ---/ \---
</pre>

1. +5V
2. GND
3. DATA
4. DATA

The easiest way to get hold of a working connector is to re-purpose an
official WiFi interface. The PCB inside these interfaces is connected
to the custom connector using a standard JST XH 4-pin connector. The
PCB has a male connector and the custom connector is wired to a female
connector. The pinout is as follows:

1. GND
2. RS485: A
3. RS485: B
4. +5V

**WARNING:** The PCB in the WiFi interface uses has an incredibly
 insecure software stack. Do NOT connect it to a network you care
 about.

# Running in OpenWRT

There is a separate repository with OpenWRT source packages:
https://github.com/andysan/gosolis-feed
