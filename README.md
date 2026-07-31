# medx-synermaxx-sms-provider

Live Synermaxx SMS provider for **Provider Protocol v1**.

Independent of medx-base. Upstream client follows [`openapi/synermaxx-sendmsg.openapi.yaml`](openapi/synermaxx-sendmsg.openapi.yaml).

| Item | Value |
|------|--------|
| API | `http://localhost:8064` |
| Protocol | `/provider/v1/...` |
| Upstream | `POST https://svr36.synermaxx.asia/api/sendmsg.php` |
| Acceptance | gateway `status: ACK` |

## Run (Docker — preferred)

```bash
cp .env.example .env
# set SYNERMAXX_SMS_APPKEY
docker compose up -d --build
curl http://localhost:8064/healthz
docker ps | grep synermaxx-sms-provider
```

## Run (local)

```bash
cp .env.example .env
go run ./cmd/api
```

## Provider Protocol

```http
POST /provider/v1/invocations
{"operation":"send-sms","input":{"to":"+639985746717","message":"hello"}}
```

Maps to OpenAPI `sendMessage` (`mobilenum` PH local form).
