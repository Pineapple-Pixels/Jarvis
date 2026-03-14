# Deploy & Operacional — Tutorial paso a paso

Guia para llevar el microservicio de local a produccion en Oracle Cloud.

---

## 1. Pre-requisitos

Antes de arrancar necesitas tener:

- [ ] Cuenta en Oracle Cloud (cloud.oracle.com/free) con VM creada (Ubuntu 22.04, VM.Standard.A1.Flex)
- [ ] Puertos abiertos en Security List: 22, 80, 443
- [ ] Clave SSH para conectarte a la VM
- [ ] Dominio apuntando a la IP de Oracle (ej: `sebas-asistente.duckdns.org`)
- [ ] API keys: Claude o OpenAI, Google Service Account (`credentials.json`), WhatsApp Business
- [ ] `.env` completo con todas las variables (copiar de `.env.example`)

---

## 2. Preparar la VM

### 2.1 Conectarse

```bash
ssh -i oracle_key ubuntu@TU_IP_ORACLE
```

### 2.2 Instalar Docker

```bash
sudo apt update && sudo apt upgrade -y
sudo apt install docker.io docker-compose-plugin -y
sudo systemctl enable docker
sudo usermod -aG docker ubuntu
# Desloguear y volver a loguear para que tome el grupo
exit
ssh -i oracle_key ubuntu@TU_IP_ORACLE
```

### 2.3 Verificar Docker

```bash
docker --version
docker compose version
```

---

## 3. SSL con Let's Encrypt

### 3.1 Instalar Certbot

```bash
sudo apt install certbot -y
```

### 3.2 Generar certificado

```bash
# Asegurate de que el puerto 80 este libre
sudo certbot certonly --standalone -d TU_DOMINIO.duckdns.org
```

Los certificados quedan en:
- `/etc/letsencrypt/live/TU_DOMINIO.duckdns.org/fullchain.pem`
- `/etc/letsencrypt/live/TU_DOMINIO.duckdns.org/privkey.pem`

### 3.3 Auto-renovacion

```bash
sudo crontab -e
# Agregar:
0 3 * * * certbot renew --quiet && docker compose restart nginx
```

---

## 4. Subir el proyecto

### 4.1 Clonar el repo

```bash
cd ~
git clone TU_REPO_URL asistente
cd asistente
```

### 4.2 Configurar environment

```bash
cp .env.example .env
nano .env
# Completar todas las variables: API keys, DB path, WhatsApp, etc.
```

### 4.3 Subir credentials.json

Desde tu maquina local:

```bash
scp -i oracle_key credentials.json ubuntu@TU_IP_ORACLE:~/asistente/credentials.json
```

---

## 5. Configurar Nginx

### 5.1 Crear nginx.conf

```bash
cp nginx.conf.example nginx.conf
nano nginx.conf
```

Asegurate de que tenga:

```nginx
server {
    listen 443 ssl;
    server_name TU_DOMINIO.duckdns.org;

    ssl_certificate /etc/letsencrypt/live/TU_DOMINIO.duckdns.org/fullchain.pem;
    ssl_certificate_key /etc/letsencrypt/live/TU_DOMINIO.duckdns.org/privkey.pem;

    # Microservicio Go
    location /api/ {
        proxy_pass http://asistente:8080;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /health {
        proxy_pass http://asistente:8080;
    }

    # n8n
    location /webhook/ {
        proxy_pass http://n8n:5678;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}

server {
    listen 80;
    server_name TU_DOMINIO.duckdns.org;
    return 301 https://$host$request_uri;
}
```

---

## 6. Deploy con Docker Compose

### 6.1 Build y levantar

```bash
# Solo el microservicio
make docker

# Full stack (asistente + n8n + nginx + postgres)
make docker-all
```

### 6.2 Verificar que levanto

```bash
docker compose ps
# Todos los servicios deben estar "Up"

curl http://localhost:8080/health
# {"status":"healthy"}

curl https://TU_DOMINIO.duckdns.org/health
# {"status":"healthy"}
```

### 6.3 Ver logs

```bash
# Todos
docker compose logs -f

# Solo el microservicio
docker compose logs -f asistente

# Solo n8n
docker compose logs -f n8n
```

---

## 7. Configurar WhatsApp webhook de entrada

### 7.1 Facebook Developer Console

1. Ir a developers.facebook.com
2. Tu app → WhatsApp → Configuration
3. Webhook URL: `https://TU_DOMINIO.duckdns.org/webhook/whatsapp`
4. Verify Token: el mismo valor que `WEBHOOK_SECRET` en tu `.env`
5. Suscribirse a: `messages`

### 7.2 Opcion A: n8n como receptor

Crear un workflow en n8n:

```
[Webhook Trigger] → [Code: extraer mensaje] → [HTTP Request: POST al microservicio] → [HTTP Request: responder por WhatsApp]
```

El Code node:

```javascript
const body = $input.first().json.body;
const entry = body.entry?.[0];
const change = entry?.changes?.[0];
const message = change?.value?.messages?.[0];

if (!message) return [];

return [{
  json: {
    from: message.from,
    text: message.text?.body || '',
    timestamp: message.timestamp
  }
}];
```

El HTTP Request al microservicio:

```
POST http://asistente:8080/api/chat
Headers: X-Webhook-Secret: {{$env.WEBHOOK_SECRET}}
Body: {
  "message": "{{$json.text}}",
  "sender": "{{$json.from}}",
  "session_id": "wa-{{$json.from}}"
}
```

### 7.3 Opcion B: endpoint directo en el microservicio (pendiente de implementar)

Esto requiere crear un controller nuevo `WhatsAppWebhookController` que:
1. Reciba el POST de Meta
2. Verifique la firma
3. Extraiga el mensaje
4. Llame al chat usecase
5. Responda via WhatsApp client

Si elegis esta opcion, el endpoint seria `POST /api/whatsapp/webhook` y `GET /api/whatsapp/webhook` (para la verificacion de Meta).

---

## 8. Rate Limiting

### 8.1 Opcion simple: Nginx

Agregar al `nginx.conf` dentro del bloque `http`:

```nginx
limit_req_zone $binary_remote_addr zone=api:10m rate=10r/s;

server {
    location /api/ {
        limit_req zone=api burst=20 nodelay;
        proxy_pass http://asistente:8080;
    }
}
```

Esto limita a 10 requests por segundo por IP con burst de 20.

### 8.2 Opcion Go: middleware

Si necesitas mas control, crear un middleware en `internal/middleware/ratelimit.go` usando `golang.org/x/time/rate`:

```go
func RateLimit(rps int) web.Interceptor {
    limiter := rate.NewLimiter(rate.Limit(rps), rps*2)
    return func(req web.InterceptedRequest) web.Response {
        if !limiter.Allow() {
            return web.NewJSONResponse(429, map[string]string{"error": "rate limit exceeded"})
        }
        return req.Next()
    }
}
```

Registrarlo en `cmd/routes.go` dentro de `middlewareMapper()`.

---

## 9. Spotify OAuth Refresh

El access token de Spotify expira cada hora. Para produccion necesitas refresh automatico.

### 9.1 Obtener refresh token

1. Ir a developer.spotify.com → tu app → Credentials
2. Usar el Authorization Code flow para obtener un `refresh_token`
3. Guardar el `refresh_token` en `.env` como `SPOTIFY_REFRESH_TOKEN`

### 9.2 Implementar refresh en el client

Agregar a `clients/spotify.go`:

```go
type SpotifyClient struct {
    accessToken  string
    refreshToken string
    clientID     string
    clientSecret string
    httpClient   *http.Client
    mu           sync.Mutex
}

func (c *SpotifyClient) refreshAccessToken() error {
    // POST https://accounts.spotify.com/api/token
    // grant_type=refresh_token&refresh_token=XXX
    // Basic auth con clientID:clientSecret
    // Parsear response y actualizar c.accessToken
}
```

Llamar `refreshAccessToken()` cuando un request devuelva 401.

---

## 10. Backup de SQLite

### 10.1 Script de backup

Crear `scripts/backup.sh`:

```bash
#!/bin/bash
BACKUP_DIR=~/backups/asistente
DATE=$(date +%Y%m%d_%H%M)
DB_PATH=./data/asistente.db

mkdir -p $BACKUP_DIR

# SQLite hot backup (safe con WAL mode)
sqlite3 $DB_PATH ".backup '$BACKUP_DIR/asistente_$DATE.db'"

# Mantener solo los ultimos 7 dias
find $BACKUP_DIR -name "*.db" -mtime +7 -delete

echo "Backup completado: asistente_$DATE.db"
```

### 10.2 Programar con cron

```bash
chmod +x scripts/backup.sh
crontab -e
# Agregar:
0 4 * * * cd ~/asistente && ./scripts/backup.sh >> ~/backups/backup.log 2>&1
```

Esto hace backup todos los dias a las 4am y mantiene los ultimos 7.

### 10.3 Backup offsite (opcional)

Para copiar a otro lugar:

```bash
# Al final de backup.sh agregar:
rclone copy $BACKUP_DIR/asistente_$DATE.db gdrive:backups/asistente/
```

Requiere configurar `rclone` con Google Drive u otro cloud storage.

---

## 11. Monitoreo basico

### 11.1 Health check con cron

```bash
crontab -e
# Agregar:
*/5 * * * * curl -sf http://localhost:8080/health > /dev/null || docker compose restart asistente
```

Si el health check falla, reinicia el container automaticamente.

### 11.2 Logs persistentes

```bash
# docker-compose.yml — agregar a cada servicio:
logging:
  driver: json-file
  options:
    max-size: "10m"
    max-file: "3"
```

### 11.3 Disk space alert

```bash
# Agregar al crontab:
0 8 * * * df -h / | awk 'NR==2{if(int($5)>80) print "DISK ALERT: "$5" used"}' | mail -s "Disk Alert" tu@email.com
```

---

## 12. Checklist de deploy

### Pre-deploy

- [ ] `.env` completo con todas las variables
- [ ] `credentials.json` presente
- [ ] `nginx.conf` configurado con tu dominio
- [ ] Certificado SSL generado
- [ ] Puertos 80 y 443 abiertos en Oracle Security List

### Deploy

- [ ] `git pull` en la VM
- [ ] `make docker-all`
- [ ] `docker compose ps` — todos Up
- [ ] `curl https://TU_DOMINIO/health` — responde OK
- [ ] Probar `POST /api/finance/expense` con un gasto de prueba
- [ ] Verificar que aparece en Google Sheets

### Post-deploy

- [ ] Configurar WhatsApp webhook en Facebook Developer Console
- [ ] Probar enviar mensaje por WhatsApp y ver que responde
- [ ] Configurar cron de backup
- [ ] Configurar cron de health check
- [ ] Configurar auto-renovacion SSL
- [ ] Verificar que los cron jobs del microservicio funcionan (esperar al horario o testear manualmente)

---

*Ultima actualizacion: 11 de marzo de 2026*
