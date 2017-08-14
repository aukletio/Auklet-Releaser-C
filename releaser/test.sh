#!/bin/bash

# Defines APP_ID and API_KEY
source secrets.sh

./releaser -appid $APP_ID\
           -apikey $API_KEY\
           -debug test/debug\
           -deploy test/deploy
