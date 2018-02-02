!#/bin/bash

#
# Assumes aws cli tool is setup and configured correctly
#
aws dynamodb query --attributes-to-get destination --table-name watft --key-conditions file://key-conditions.json
