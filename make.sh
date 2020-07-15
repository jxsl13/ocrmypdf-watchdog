#!/bin/sh
git pull
sudo docker-compose build --force-rm --no-cache 
sudo docker-compose up -d
sudo docker ps
sleep 5
sudo docker logs ocrmypdf-watchdog