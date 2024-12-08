#!/bin/bash

echo "Builds app"
go build -o birthday-scheduler.bin

cd ./deploy

echo "build the zip package"
./deploy.bin -target mailrelay -outdir ~/app/go/birthday-scheduler/zips/
cd ~/app/go/birthday-scheduler/

echo "update the service"
./update-service.sh

echo "Ready to fly"