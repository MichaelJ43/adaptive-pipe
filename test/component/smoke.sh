#!/usr/bin/env sh
# Component smoke: expects docker compose up and curl.
set -e
BASE="${BASE_URL:-http://127.0.0.1:8080}"
echo "GET $BASE/healthz"
curl -sf "$BASE/healthz" | grep -q ok || curl -sf "$BASE/healthz"

echo "POST webhook (tenant demo)"
curl -sf -X POST "$BASE/api/v1/tenants/demo/webhooks/github" \
  -H "Content-Type: application/json" \
  -H "X-GitHub-Delivery: smoke-$(date +%s)" \
  -d '{"after":"deadbeef","repository":{"name":"widget","owner":{"login":"acme"}}}' \
  | grep -q run_id

echo "component smoke OK"
