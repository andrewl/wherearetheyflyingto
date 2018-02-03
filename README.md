[![Build Status](https://travis-ci.org/andrewl/wherearetheyflyingto.svg?branch=master)](https://travis-ci.org/andrewl/wherearetheyflyingto) 

[Latest Map](https://andrewl.github.io/wherearetheyflyingto/)

What does this do?
==================

Where Are They Flying To (WATFT) outputs information about the destination and
altitude of aircraft flying over your house. It can be used to build heatmaps
of destinations or even an Alexa skill.
If you've ever looked up at a plane and thought, "I wonder where they're flying
to?" then WATFT is for you.

How does it work?
=================

WATFT reads data from a TCP stream in SBS format (eg from a machine running
dump1090), processes it and writes it to a database file. WATFT can read
this database and produce a heatmap showing the destinations of the aircraft it has plotted.
@todo add flightaware

![Architecture Diagram](https://github.com/andrewl/wherearetheyflyingto/blob/master/watft_architecture.png?raw=true)

How to run it
=============

What you'll need
@todo - pc/pi, adsb, flightaware
If you want to run WATFT on linux on either an x64 (eg Intel/AMD) or arm (eg Raspberry Pi) then all you need to do is download the relevant binary file from the [Latest Release](https://github.com/andrewl/wherearetheyflyingto/releases/latest).

to build alexa skill


to build heatmap

Set the value of the environment variable WATFT_SERVER to the address and port of the machine that is dumping out the SBS data, eg
`export WATFT_SERVER=127.0.0.1:33005`

About the code
==============

