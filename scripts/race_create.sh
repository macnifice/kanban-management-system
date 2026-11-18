#!/usr/bin/env bash
# Concurrencia: 8 creaciones simultáneas en una columna con WIP=3. Debe haber 3 x 201 y 5 x 409, sin 500.
set -u
B=${1:-http://127.0.0.1:8080}
g() { python3 -c "import sys,json;print(json.load(sys.stdin)$1)"; }
TOKEN=$(curl -s -X POST $B/api/auth/register -d "{\"email\":\"rc$RANDOM$RANDOM@t.com\",\"name\":\"R\",\"password\":\"secret1\"}" | g "['token']")
H="Authorization: Bearer $TOKEN"
BID=$(curl -s -X POST -H "$H" $B/api/boards -d '{"name":"race-create"}' | g "['id']")
COL=$(curl -s -H "$H" $B/api/boards/$BID | g "['columns'][0]['id']")
curl -s -X PATCH -H "$H" $B/api/columns/$COL -d '{"wip_limit":3}' >/dev/null
RES=$(seq 1 8 | xargs -P8 -I{} curl -s -o /dev/null -w "%{http_code}\n" -X POST -H "$H" $B/api/cards -d "{\"column_id\":$COL,\"title\":\"c{}\"}" | sort | uniq -c | tr '\n' ' ')
COUNT=$(curl -s -H "$H" $B/api/boards/$BID/metrics | g "['columns'][0]['count']")
curl -s -X DELETE -H "$H" -o /dev/null $B/api/boards/$BID
echo "respuestas: $RES| tarjetas en la columna limitada: $COUNT"
[ "$COUNT" = "3" ] && ! echo "$RES" | grep -q 500
