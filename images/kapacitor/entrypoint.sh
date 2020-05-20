#!/bin/sh
set -e

KAPACITOR_HOSTNAME=${KAPACITOR_HOSTNAME:-$HOSTNAME}
export KAPACITOR_HOSTNAME
mkdir -p /var/lib/kapacitor/logs

kapacitord
