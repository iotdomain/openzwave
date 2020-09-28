# OpenZwave publisher on the IoTDomain

This publisher brings zwave devices to the IoTDomain. It uses the openzwave library and the goopenzwave package to interface with the library.

## Dependencies

1. Openzwave-1.5
  
The openzwave library can be installed via a package manager or from source. The easiest is to use a package manager, although most are not up to date and are missing information on newer devices. Installing from source might not be for everyone though. 

On Ubuntu:
$ sudo apt install libopenzwave libopenzwave-dev 

2. goopenzwave from https://github.com/jimjibone/goopenzwave

This library provides the golang interface to openzwave. It is pulled in automatically by go.mod and requires libopenzwave-dev to be installed.

$ go dep ensure installs the OpenZWave adapter from https://github.com/jimjibone/goopenzwave

This assumes that the openzwave configuration is in /etc/openzwave or /usr/local/etc/openzwave. This path can be changed in the openzwave.yaml config file. 

## Configuration

See iotdomain's config/openzwave.yaml for the configuration options. This publisher runs out of the box with most zwave USB controllers.

## Todo

1. Update the value of pushbuttons AddNode, RemoveNode, Healnetwork while the process is running.
2. Get neighbours. This needs an update to goopenzwave
