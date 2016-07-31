[![Build Status](https://travis-ci.org/andrewl/wherearetheyflyingto.svg?branch=master)](https://travis-ci.org/andrewl/wherearetheyflyingto) 

What does this do?
==================

Where Are They Flying To (WATFT) produces a heatmap showing the destinations that the aircraft flying above your house are heading for. If you've ever looked up at a plane and thought, "I wonder where they're flying to?" then WATFT is for you.

How does it work?
=================

WATFT reads data from a TCP stream in SBS format (eg from a machine running dump1090), processes it and writes it to a database file. WATFT can read this database and produce a heatmap showing the destinations of the aircraft it has plotted.

How to use it
=============

If you want to run WATFT on linux on either an x64 (eg Intel/AMD) or arm (eg Raspberry Pi) then all you need to do is download the relevant binary file from the [Latest Release](https://github.com/andrewl/wherearetheyflyingto/releases/latest).

If you want to run WATFT on another system then you'll need to download or clone this repo, then you'll need to have [go](https://golang.org) installed then run the following to install the dependencies and build:
```
go get -t -v ./...
go build
```

@todo - running

Set the value of the environment variable WATFT_SERVER to the address and port of the machine that is dumping out the SBS data, eg
`export WATFT_SERVER=127.0.0.1:33005`

About the code
==============

@todo
