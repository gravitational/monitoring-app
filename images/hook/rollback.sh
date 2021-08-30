#!/usr/bin/env bash
set -eux
# TODO (s.antipov): write rollback for Helm to Helm release
for release_name in monitoring watcher nethealth; do
    if helm3 --namespace monitoring status "$release_name" &>/dev/null; then
        helm3 --namespace monitoring uninstall "$release_name"
    fi
done

echo "---> Reverting changeset $RIG_CHANGESET"
rig revert
rig cs delete --force -c cs/"$RIG_CHANGESET"
