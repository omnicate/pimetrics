#!/bin/bash

if [ ! -f /usr/bin/yq ]; then
  wget https://github.com/mikefarah/yq/releases/download/3.4.1/yq_linux_arm -O /usr/bin/yq
  chmod +x /usr/bin/yq
fi

function update {
  VERSION=$1
  PI_ARCH=$2

  systemctl stop pimetrics

  wget -q https://github.com/omnicate/pimetrics/releases/download/v$VERSION/pimetrics-v$VERSION.linux-$PI_ARCH.tar.gz

  tar -xvzf pimetrics-v$VERSION.linux-$PI_ARCH.tar.gz ./

  systemctl start pimetrics
}


PI_NAME=`hostname`

# Create temporary updater folder
mkdir updater

# Download config from github
wget -q https://raw.githubusercontent.com/omnicate/pimetrics/master/config.yaml -O ./updater/config.yaml

#Compare two configs
DIFF_RESULT=`diff ./config.yaml ./updater/config.yaml`
if [ ! $DIFF_RESULT ]; then
  # Something changed, update all configs and restart

  # Write new pimetrics config into updater folder
  yq read ./updater/config.yaml $PI_NAME.config > ./pimetrics-config.yaml

  OLD_VERSION=`yq read ./config.yaml $PI_NAME.sw_version`
  NEW_VERSION=`yq read ./updater/config.yaml $PI_NAME.sw_version`

  if [ ! $OLD_VERSION == $NEW_VERSION ]; then
    CONFIG_ARCH=`yq read ./updater/config.yaml $PI_NAME.target`
    update $NEW_VERSION $CONFIG_ARCH
  fi

  # Copy new config to old location
  cp ./updater/config.yaml ./config.yaml
fi

# Clean up
rm -rf updater
