#!/usr/bin/env bash
set -euo pipefail

image="${1:?image is required}"
expected_version="${2:?expected version is required}"
test "$(docker run --rm "${image}" -version)" = "${expected_version}"

container="$(docker run --detach --publish 127.0.0.1::8080 "${image}" -capacity 1 -refill 1h)"
cleanup() {
  docker logs "${container}" || true
  docker rm --force "${container}" >/dev/null
}
trap cleanup EXIT

address="$(docker port "${container}" 8080/tcp)"
url="http://${address}/"
ready=false
for ((attempt = 0; attempt < 30; attempt++)); do
  status="$(curl --silent --output /dev/null --write-out '%{http_code}' --max-time 1 "${url}" || true)"
  if [ "${status}" = 200 ]; then
    ready=true
    break
  fi
  sleep 1
done
if [ "${ready}" != true ]; then
  echo 'Server did not return HTTP 200 within the startup deadline' >&2
  exit 1
fi

headers="$(curl --silent --show-error --include --max-time 5 "${url}")"
[[ "${headers}" == *'429 Too Many Requests'* ]]
[[ "${headers}" == *'Retry-After:'* ]]

# docker stop sends SIGTERM; exit 137 would reveal a forced kill on timeout.
docker stop --time 15 "${container}" >/dev/null
test "$(docker inspect --format '{{.State.ExitCode}}' "${container}")" = 0
