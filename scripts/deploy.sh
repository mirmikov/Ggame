#!/usr/bin/env bash
set -euo pipefail

domain="${1:-${SITE_ADDRESS:-}}"

write_env() {
  local site_address="$1"
  local postgres_password
  postgres_password="${POSTGRES_PASSWORD:-$(openssl rand -hex 24 2>/dev/null || date +%s%N)}"
  cat > .env <<EOF
SITE_ADDRESS=${site_address}
HTTP_PORT=80
HTTPS_PORT=443
POSTGRES_DB=${POSTGRES_DB:-ggame}
POSTGRES_USER=${POSTGRES_USER:-ggame}
POSTGRES_PASSWORD=${postgres_password}
EOF
}

if [[ -n "$domain" ]]; then
  write_env "$domain"
elif [[ ! -f .env ]]; then
  write_env "localhost"
fi

git fetch --prune
git pull --ff-only

docker compose up -d --build --force-recreate
docker compose ps

echo "Checking local health endpoint..."
curl -fsS http://127.0.0.1/api/health
echo
echo "Deploy complete."
