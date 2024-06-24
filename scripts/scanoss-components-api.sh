#!/bin/bash

##########################################
#
# This script is designed to run by Systemd SCANOSS Components API service.
# It rotates scanoss log file and starts Components API.
# Install it in /usr/local/bin
#
################################################################
DEFAULT_ENV="prod"
ENVIRONMENT="${1:-$DEFAULT_ENV}"
LOGFILE=/var/log/scanoss/components/scanoss-components-${ENVIRONMENT}.log
CONF_FILE=/usr/local/etc/scanoss/components/app-config-${ENVIRONMENT}.json
# Rotate log
if [ -f "$LOGFILE" ] ; then
  echo "rotating logfile..."
  TIMESTAMP=$(date '+%Y%m%d-%H%M%S')
  BACKUP_FILE=$LOGFILE.$TIMESTAMP
  cp "$LOGFILE" "$BACKUP_FILE"
  gzip -f "$BACKUP_FILE"
fi
echo > "$LOGFILE"

#start API
echo "Starting SCANOSS Components API"

exec /usr/local/bin/scanoss-components-api --json-config "$CONF_FILE" > "$LOGFILE" 2>&1
