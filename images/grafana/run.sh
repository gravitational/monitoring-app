#!/bin/bash

HEADER_CONTENT_TYPE="Content-Type: application/json"
HEADER_ACCEPT="Accept: application/json"

INFLUXDB_HOST=${INFLUXDB_HOST:-"influxdb"}
INFLUXDB_DATABASE=${INFLUXDB_DATABASE:-k8s}
INFLUXDB_PASSWORD=${INFLUXDB_PASSWORD:-root}
INFLUXDB_USER=${INFLUXDB_USER:-root}

# Allow access to dashboards without having to log in
export GF_AUTH_ANONYMOUS_ENABLED=${GF_AUTH_ANONYMOUS_ENABLED:-true}
export GF_SERVER_HTTP_PORT=3000
export GF_SERVER_HTTP_ADDR=0.0.0.0

BACKEND_ACCESS_MODE=${BACKEND_ACCESS_MODE:-proxy}
INFLUXDB_SERVICE_URL=${INFLUXDB_SERVICE_URL}
if [ -n "$INFLUXDB_SERVICE_URL" ]; then
  echo "Influxdb service URL is provided."
else
  INFLUXDB_SERVICE_URL="http://${INFLUXDB_PORT_8086_TCP_ADDR}:${INFLUXDB_PORT_8086_TCP_PORT}"
fi

echo "Using the following URL for InfluxDB: ${INFLUXDB_SERVICE_URL}"
echo "Using the following backend access mode for InfluxDB: ${BACKEND_ACCESS_MODE}"

set -m
echo "Starting Grafana in the background"
exec /usr/sbin/grafana-server --homepath=/usr/share/grafana --config=/etc/grafana/grafana.ini cfg:default.paths.data=/var/lib/grafana cfg:default.paths.logs=/var/log/grafana &

echo "Waiting for Grafana to come up..."
until $(curl --fail --output /dev/null --silent --user "${GF_SECURITY_ADMIN_USER}":"${GF_SECURITY_ADMIN_PASSWORD}" http://localhost:${GF_SERVER_HTTP_PORT}/api/org); do
  printf "."
  sleep 2
done
echo "Grafana is up and running."
echo "Creating default influxdb datasource..."
curl -i -XPOST -H "${HEADER_ACCEPT}" -H "${HEADER_CONTENT_TYPE}" --user "${GF_SECURITY_ADMIN_USER}":"${GF_SECURITY_ADMIN_PASSWORD}" "http://localhost:${GF_SERVER_HTTP_PORT}/api/datasources" -d '
{
  "name": "influxdb-datasource",
  "type": "influxdb",
  "access": "'"${BACKEND_ACCESS_MODE}"'",
  "isDefault": true,
  "url": "'"${INFLUXDB_SERVICE_URL}"'",
  "password": "'"${INFLUXDB_PASSWORD}"'",
  "user": "'"${INFLUXDB_USER}"'",
  "database": "'"${INFLUXDB_DATABASE}"'"
}'

echo ""
echo "Bringing Grafana back to the foreground"
fg
