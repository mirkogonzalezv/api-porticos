#!/usr/bin/env bash
set -euo pipefail

if [[ -z "${JWT:-}" || -z "${VEHICULO_ID:-}" ]]; then
  echo "Usage: JWT=... VEHICULO_ID=... $0"
  exit 1
fi

DEBUG=${DEBUG:-0}

# Pórtico: C001 Costanera Entrada Prueba
LAT=3.50703
LNG=-76.297292

read T1 T2 T3 T4 <<< "$(python3 - <<'PY'
import datetime
base = datetime.datetime.now(datetime.timezone.utc)
t1 = base
t2 = base + datetime.timedelta(seconds=6)
t3 = base + datetime.timedelta(seconds=16)
t4 = base + datetime.timedelta(seconds=18)
print(t1.strftime("%Y-%m-%dT%H:%M:%SZ"),
      t2.strftime("%Y-%m-%dT%H:%M:%SZ"),
      t3.strftime("%Y-%m-%dT%H:%M:%SZ"),
      t4.strftime("%Y-%m-%dT%H:%M:%SZ"))
PY
)"

send_payload () {
  local payload="$1"
  if [[ "$DEBUG" == "1" ]]; then
    echo "PAYLOAD: $payload"
  fi

  curl --request POST \
    --url http://localhost:3200/api/v1/tracking/position \
    --header "Authorization: Bearer $JWT" \
    --header 'Content-Type: application/json' \
    --data-raw "$payload"

  echo ""
}

# 1) Dentro (entrada)
payload1=$(printf '{"vehiculoId":"%s","lat":%.6f,"lng":%.6f,"speed":%.1f,"heading":%d,"timestamp":"%s"}' \
  "$VEHICULO_ID" "$LAT" "$LNG" 6.0 5 "$T1")
send_payload "$payload1"

# 2) Dentro (confirma ENTERED/INSIDE)
payload2=$(printf '{"vehiculoId":"%s","lat":%.6f,"lng":%.6f,"speed":%.1f,"heading":%d,"timestamp":"%s"}' \
  "$VEHICULO_ID" 3.507028 -76.297290 5.0 3 "$T2")
send_payload "$payload2"

# 3) Fuera (1ra salida) ~18m norte
payload3=$(printf '{"vehiculoId":"%s","lat":%.6f,"lng":%.6f,"speed":%.1f,"heading":%d,"timestamp":"%s"}' \
  "$VEHICULO_ID" 3.507192 -76.297292 12.0 2 "$T3")
send_payload "$payload3"

# 4) Fuera (2da salida) ~30m norte => VALIDATED
payload4=$(printf '{"vehiculoId":"%s","lat":%.6f,"lng":%.6f,"speed":%.1f,"heading":%d,"timestamp":"%s"}' \
  "$VEHICULO_ID" 3.507300 -76.297292 12.0 2 "$T4")
send_payload "$payload4"
