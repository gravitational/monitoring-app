#!/bin/sh

if [ -d ${OLD_DATA_DIR}/influxdb ]; then 
  cp -a ${OLD_DATA_DIR}/influxdb/* ${NEW_DATA_DIR}/
  ls -and $NEW_DATA_DIR | awk '{system("chown -R "$3":"$4" $NEW_DATA_DIR/*")}'
  echo "Data migrated."
fi
