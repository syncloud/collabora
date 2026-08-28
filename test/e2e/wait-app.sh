#!/bin/bash -e

APP_DOMAIN=$1

for i in $(seq 1 120); do
  CODE=$(curl -sk -o /dev/null -w '%{http_code}' "https://${APP_DOMAIN}" || true)
  case "${CODE}" in
    200|301|302|303|307|308)
      echo "${APP_DOMAIN} is ready (${CODE})"
      exit 0
      ;;
  esac
  echo "waiting for ${APP_DOMAIN}, got ${CODE}"
  sleep 5
done

echo "${APP_DOMAIN} did not become ready"
exit 1
