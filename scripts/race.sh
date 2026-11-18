#!/usr/bin/env bash
# Concurrencia: 8 tarjetas se mueven a la vez a una columna con WIP=2. Debe haber 2 x 200 y 6 x 409, sin 500.
set -u
B=${1:-http://127.0.0.1:8080}
g() { python3 -c "import sys,json;print(json.load(sys.stdin)$1)"; }
TOKEN=$(curl -s -X POST $B/api/auth/register -d "{\"email\":\"race$RANDOM$RANDOM@t.com\",\"name\":\"R\",\"password\":\"secret1\"}" | g "['token']")
H="Authorization: Bearer $TOKEN"
BID=$(curl -s -X POST -H "$H" $B/api/boards -d '{"name":"race"}' | g "['id']")
BD=$(curl -s -H "$H" $B/api/boards/$BID)
T=$(echo "$BD" | g "['columns'][0]['id']"); P=$(echo "$BD" | g "['columns'][1]['id']")
curl -s -X PATCH -H "$H" $B/api/columns/$P -d '{"wip_limit":2}' >/dev/null
IDS=""; for n in 1 2 3 4 5 6 7 8; do IDS="$IDS $(curl -s -X POST -H "$H" $B/api/cards -d "{\"column_id\":$T,\"title\":\"c$n\"}" | g "['id']")"; done
RES=$(echo $IDS | tr ' ' '\n' | xargs -P8 -I{} curl -s -o /dev/null -w "%{http_code}\n" -X PUT -H "$H" "$B/api/cards/{}/move" -d "{\"column_id\":$P,\"position\":0}" | sort | uniq -c | tr '\n' ' ')
COUNT=$(curl -s -H "$H" $B/api/boards/$BID/metrics | g "['columns'][1]['count']")
curl -s -X DELETE -H "$H" -o /dev/null $B/api/boards/$BID
echo "respuestas: $RES| tarjetas en la columna limitada: $COUNT"
[ "$COUNT" = "2" ] && ! echo "$RES" | grep -q 500
