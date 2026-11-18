#!/usr/bin/env bash
# Prueba de humo de la API (requiere docker compose up). Uso: scripts/smoke.sh [base_url]
set -u
B=${1:-http://127.0.0.1:8080}
pass=0; fail=0
j() { python3 -c "import sys,json; d=json.load(sys.stdin); print(eval(sys.argv[1]))" "$1"; }
check() { if [ $# -ne 3 ]; then fail=$((fail+1)); echo "FAIL $1 (argumentos mal formados: $#)"; return; fi; if [ "$2" = "$3" ]; then pass=$((pass+1)); echo "ok   $1"; else fail=$((fail+1)); echo "FAIL $1 (esperado=$3 obtenido=$2)"; fi; }
TOKEN=""
# req METHOD PATH [BODY] -> cuerpo de la respuesta
req() {
  local m=$1 p=$2 body=${3:-}
  if [ -n "$body" ]; then curl -s -X "$m" -H "Authorization: Bearer $TOKEN" -d "$body" "$B$p"
  else curl -s -X "$m" -H "Authorization: Bearer $TOKEN" "$B$p"; fi
}
# status METHOD PATH [BODY] -> código HTTP
status() {
  local m=$1 p=$2 body=${3:-}
  if [ -n "$body" ]; then curl -s -o /dev/null -w "%{http_code}" -X "$m" -H "Authorization: Bearer $TOKEN" -d "$body" "$B$p"
  else curl -s -o /dev/null -w "%{http_code}" -X "$m" -H "Authorization: Bearer $TOKEN" "$B$p"; fi
}
mv() { status PUT "/api/cards/$1/move" "{\"column_id\":$2,\"position\":0}"; }

check "sin token -> 401" "$(status GET /api/boards)" 401
EMAIL="smoke$RANDOM@test.com"
TOKEN=$(curl -s -X POST "$B/api/auth/register" -d "{\"email\":\"$EMAIL\",\"name\":\"Smoke\",\"password\":\"secret1\"}" | j "d['token']")
UID_=$(req GET /api/me | j "d['id']")
BAD_LOGIN="{\"email\":\"$EMAIL\",\"password\":\"mala\"}"
GOT=$(curl -s -o /dev/null -w '%{http_code}' -X POST "$B/api/auth/login" -d "$BAD_LOGIN")
check "login incorrecto -> 401" "$GOT" 401

BID=$(req POST /api/boards '{"name":"Smoke board"}' | j "d['id']")
BOARD=$(req GET "/api/boards/$BID")
TODO=$(echo "$BOARD" | j "d['columns'][0]['id']")
PROG=$(echo "$BOARD" | j "d['columns'][1]['id']")
DONE=$(echo "$BOARD" | j "d['columns'][2]['id']")

check "limite WIP=1" "$(status PATCH "/api/columns/$PROG" '{"wip_limit":1}')" 200
C1=$(req POST /api/cards "{\"column_id\":$TODO,\"title\":\"A\",\"priority\":\"high\",\"assignee_id\":$UID_}" | j "d['id']")
C2=$(req POST /api/cards "{\"column_id\":$TODO,\"title\":\"B\"}" | j "d['id']")
check "mover A a En progreso" "$(mv $C1 $PROG)" 200
check "mover B a columna llena -> 409" "$(mv $C2 $PROG)" 409
FULL_BODY="{\"column_id\":$PROG,\"title\":\"X\"}"
GOT=$(status POST /api/cards "$FULL_BODY")
check "crear en columna llena -> 409" "$GOT" 409
check "reordenar dentro de la columna llena" "$(mv $C1 $PROG)" 200
sleep 2
check "mover A a Hecho" "$(mv $C1 $DONE)" 200
check "mover B a En progreso (ya libre)" "$(mv $C2 $PROG)" 200

M=$(req GET "/api/boards/$BID/metrics")
check "cycle time count=1" "$(echo "$M" | j "d['cycle_time']['count']")" 1
check "cycle time >= 2s" "$(echo "$M" | j "d['cycle_time']['avg_hours']*3600>=1.5")" True
check "totales por estado (done,active)" "$(echo "$M" | j "(d['totals_by_kind']['done'],d['totals_by_kind']['active'])")" "(1, 1)"
check "historial de A (3 eventos)" "$(req GET "/api/cards/$C1/history" | j "len(d)")" 3

LID=$(req POST "/api/boards/$BID/labels" '{"name":"bug","color":"#ff0000"}' | j "d['id']")
PATCH_BODY="{\"label_ids\":[$LID],\"priority\":\"urgent\"}"
GOT=$(req PATCH "/api/cards/$C1" "$PATCH_BODY" | j "(d['label_ids'],d['priority'])")
check "asignar etiqueta y prioridad" "$GOT" "([$LID], 'urgent')"
check "404 columna inexistente" "$(status PATCH /api/columns/999999 '{"title":"x"}')" 404
check "borrar tablero" "$(status DELETE "/api/boards/$BID")" 204
echo "--- $pass ok, $fail fallos"; [ $fail -eq 0 ]
