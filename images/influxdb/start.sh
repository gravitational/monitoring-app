#!/bin/sh
exec /influxdb/influxd -config /etc/influxdb.toml &
until $(/influxdb/influx -username root -password root -execute "create user root with password 'root' with all privileges"); do
    sleep 2
done
until $(/influxdb/influx -username root -password root -execute "create user grafana with password 'grafana'"); do
    sleep 2
done
while true; do sleep 10000; done
