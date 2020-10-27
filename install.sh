#!/bin/bash

TARGET=$1

echo Installing helpers
scp deployment/systemd/pimetrics.service ubuntu@$TARGET:/home/ubuntu/pimetrics.service
scp deployment/autoupdate.sh ubuntu@$TARGET:/home/ubuntu/autoupdate.sh
ssh ubuntu@$TARGET sudo mv pimetrics.service /etc/systemd/system/pimetrics.service
ssh ubuntu@$TARGET sudo systemctl daemon-reload
ssh ubuntu@$TARGET chmod +x autoupdate.sh
ssh ubuntu@$TARGET ./autoupdate.sh

echo Copying initial config
scp config.yaml ubuntu@$TARGET:/home/ubuntu/config.yaml

echo Starting pimetrics
ssh ubuntu@$TARGET ./autoupdate.sh