#!/bin/bash
CONF_DIR=/srv/data-api_interface/config

echo "setting up the data interface api"
sudo mkdir -p $CONF_DIR
sudo cp -rn ./config/dynamic_mapping.json $CONF_DIR/dynamic_mapping.json
docker compose up -d --build