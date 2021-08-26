#!/bin/sh
set -eux
for release_name in monitoring watcher nethealth; do
    if /opt/bin/helm3 --namespace monitoring status "$release_name" >/dev/null 2>&1; then
        /opt/bin/helm3 --namespace monitoring uninstall "$release_name"
    fi
done
