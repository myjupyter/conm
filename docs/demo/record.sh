#!/bin/sh
set -eu
cd "$(dirname "$0")"

work=$(mktemp -d)
cleanup() {
	docker compose down -v >/dev/null 2>&1 || true
	security delete-generic-password -s conm-demo -a shared >/dev/null 2>&1 || true
	security delete-generic-password -s conm.00000000-0000-4000-8000-00000000c0de -a orders >/dev/null 2>&1 || true
	rm -rf "$work"
}
trap cleanup EXIT

security delete-generic-password -s conm-demo -a shared >/dev/null 2>&1 || true
security delete-generic-password -s conm.00000000-0000-4000-8000-00000000c0de -a orders >/dev/null 2>&1 || true
docker compose up -d
until docker compose exec -T postgres pg_isready -q -U demo -d orders; do sleep 1; done
until nc -z 127.0.0.1 2222; do sleep 1; done
sleep 5

mkdir "$work/bin"
go build -C ../.. -o "$work/bin/conm" .

export XDG_CONFIG_HOME="$work/config" XDG_DATA_HOME="$work/data"
export PATH="$PWD/stubs:$work/bin:$PATH"
vhs "${1:-demo.tape}"
