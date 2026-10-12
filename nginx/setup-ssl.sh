#!/bin/sh
set -e

# Target paths for active certificates
mkdir -p /etc/ssl/active

# Default paths if not specified
SSL_CERT_PATH="${SSL_CERT_PATH:-/etc/letsencrypt/live/thanhhieutiet.shop/fullchain.pem}"
SSL_KEY_PATH="${SSL_KEY_PATH:-/etc/letsencrypt/live/thanhhieutiet.shop/privkey.pem}"

# Check if valid live certificates exist
if [ -f "$SSL_CERT_PATH" ] && [ -f "$SSL_KEY_PATH" ]; then
    echo "[LabShop Nginx] Found active SSL certificates at $SSL_CERT_PATH"
    ln -sf "$SSL_CERT_PATH" /etc/ssl/active/fullchain.pem
    ln -sf "$SSL_KEY_PATH" /etc/ssl/active/privkey.pem
else
    echo "[LabShop Nginx] Live SSL certificates not found at $SSL_CERT_PATH (or key missing)."
    echo "[LabShop Nginx] Falling back to generated self-signed certificate."
    ln -sf /etc/ssl/dummy/fullchain.pem /etc/ssl/active/fullchain.pem
    ln -sf /etc/ssl/dummy/privkey.pem /etc/ssl/active/privkey.pem
fi
