# TLS certificates directory

Place your SSL/TLS certificate files here:
- `fullchain.pem`  — certificate + intermediate chain
- `privkey.pem`    — private key

## Development (self-signed)

```bash
openssl req -x509 -nodes -days 365 -newkey rsa:2048 \
  -keyout privkey.pem -out fullchain.pem \
  -subj "/CN=localhost"
```

## Production

Use Let's Encrypt (certbot) or your CA of choice.
These `.pem` files are excluded from version control via `.gitignore`.
