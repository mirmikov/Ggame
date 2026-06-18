#!/usr/bin/env bash
set -euo pipefail

domain="${1:-${SITE_ADDRESS:-}}"

if [[ -n "$domain" ]]; then
  cat > .env <<EOF
SITE_ADDRESS=${domain}
HTTP_PORT=80
HTTPS_PORT=443
EOF
elif [[ ! -f .env ]]; then
  cp .env.example .env
fi

git fetch --prune
git pull --ff-only

docker compose up -d --build --force-recreate
docker compose ps

echo "Checking local health endpoint..."
curl -fsS http://127.0.0.1/api/health
echo
echo "Deploy complete."
