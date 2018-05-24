#!/bin/bash

#
# Assumes aws cli tool is setup and configured correctly this downloads the destinations from the dynamodb table
#
aws dynamodb query --attributes-to-get destination altitude --table-name watft --query-filter file://query-filter.json --key-conditions file://key-conditions.json
