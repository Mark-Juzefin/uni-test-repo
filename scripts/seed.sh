#!/usr/bin/env bash
# Seeds the products service with sample data to test list pagination.
# Usage:
#   ./scripts/seed.sh [count]      # default count = 25
#   BASE=http://localhost:3000 ./scripts/seed.sh 50
set -euo pipefail

BASE="${BASE:-http://localhost:3000}"
COUNT="${1:-25}"

echo "Seeding $COUNT products to $BASE ..."
for i in $(seq 1 "$COUNT"); do
  curl -sS -X POST "$BASE/products" \
    -H "Content-Type: application/json" \
    -d "{\"name\":\"Product $i\",\"description\":\"seeded product #$i\",\"price\":$((i * 100))}" \
    -o /dev/null -w "%{http_code}  product $i\n" || echo "request $i failed"
done
echo "Done. Try: curl \"$BASE/products?limit=10&offset=0\""
