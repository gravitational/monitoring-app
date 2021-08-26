#!/usr/bin/env bash
set -eux

echo "---> Reverting changeset $RIG_CHANGESET"
for release_name in monitoring watcher nethealth; do
    if helm3 --namespace monitoring status "$release_name" &>/dev/null; then
        helm3 --namespace monitoring uninstall "$release_name"
    fi
done
rig revert
rig cs delete --force -c cs/"$RIG_CHANGESET"
