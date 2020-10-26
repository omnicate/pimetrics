#!/bin/bash

function update {
  VERSION=$1
  PI_ARCH=$2

  echo Stopping pimetrics
  systemctl stop pimetrics

  echo Downloading new version
  wget https://github.com/omnicate/pimetrics/releases/download/v$VERSION/pimetrics-v$VERSION.linux-$PI_ARCH.tar.gz

  echo Extracting new version
  tar -xvzf pimetrics-v$VERSION.linux-$PI_ARCH.tar.gz ./
  rm pimetrics-v$VERSION.linux-$PI_ARCH.tar.gz

  echo Starting pimetrics
  systemctl start pimetrics
}


if [ ! -f /usr/bin/yq ]; then
  wget -q https://github.com/mikefarah/yq/releases/download/3.4.1/yq_linux_arm -O /usr/bin/yq
  chmod +x /usr/bin/yq
fi

PI_NAME=`hostname`

# Create temporary updater folder
mkdir updater

# Download config from github
echo Downloading config from github
wget -q https://raw.githubusercontent.com/omnicate/pimetrics/master/config.yaml -O ./updater/config.yaml

#Compare two configs
echo Comparing Configs
diff ./config.yaml ./updater/config.yaml > /dev/null 2>&1
DIFF_RESULT=$?
if [ $DIFF_RESULT ]; then
  # Something changed, update all configs and restart
  echo Config has changed
  # Write new pimetrics config into updater folder
  yq read ./updater/config.yaml $PI_NAME.config > ./pimetrics-config.yaml

  OLD_VERSION=`yq read ./config.yaml $PI_NAME.sw_version`
  NEW_VERSION=`yq read ./updater/config.yaml $PI_NAME.sw_version`

  if [ ! $OLD_VERSION == $NEW_VERSION ]; then
    echo Updating to $NEW_VERSION
    CONFIG_ARCH=`yq read ./updater/config.yaml $PI_NAME.target`
    update $NEW_VERSION $CONFIG_ARCH
    echo Finished Updating
  fi

  # Copy new config to old location
  cp ./updater/config.yaml ./config.yaml
fi

# Clean up
echo Cleanup
rm -rf updater
