# Notify

A simple REST service to send email notifications (and optionally SMS in the future) with support for:

- **SQLite + GORM** for persistent storage
- **Background worker** for retries on failed or queued messages
- **SendGrid SMTP** to send emails from your custom domain
- **Bearer token** authentication (optional)
- **Configurable** via environment variables
- **Graceful shutdown** and structured logging (`slog`)

## Table of Contents

- [Requirements](#requirements)
- [Installation](#installation)
- [Environment Variables](#environment-variables)
- [Running](#running)
- [Usage Examples](#usage-examples)
- [End-to-End Flow](#end-to-end-flow)
- [License](#license)

---

## Requirements

- **Go 1.21+** (for `log/slog` usage)
- A **SendGrid** account (or another SMTP-compatible service) if you want to send real emails.

---

## Installation

Clone or download this repo:

```bash
git clone https://github.com/temirov/notify.git
cd notify
```

Make sure Go is available:

```bash
go version
```

*(Should be Go 1.21+.)*

---

## Environment Variables

| Variable                  | Default                   | Description                                                                                                      |
|---------------------------|---------------------------|------------------------------------------------------------------------------------------------------------------|
| `SERVER_PORT`             | `8080`                    | Port where the HTTP server listens.                                                                              |
| `DATABASE_PATH`           | `app.db`                  | SQLite file location.                                                                                            |
| `LOG_LEVEL`               | `INFO`                    | Possible values: `DEBUG`, `INFO`, `WARN`, `ERROR`.                                                               |
| `NOTIFICATION_AUTH_TOKEN` | *(empty)*                 | If set, service requires `Authorization: Bearer <token>` header for all requests. If empty, no auth is required. |
| `MAX_RETRIES`             | `3`                       | How many times the background worker will attempt to resend a failed notification.                               |
| `RETRY_INTERVAL_SEC`      | `15`                      | Interval (in seconds) between retry scans.                                                                       |
| `SENDGRID_USERNAME`       | `apikey`                  | SMTP username for SendGrid (often literally "apikey").                                                           |
| `SENDGRID_PASSWORD`       | *(empty)*                 | Your SendGrid API key (used as the SMTP password).                                                               |
| `FROM_EMAIL`              | `support@rsvp.mprlab.com` | The default “from” address for sending emails. Must match your verified domain on SendGrid.                      |

### Example `.env` File

You can create a local `.env` file for convenience:

```
SERVER_PORT=8080
DATABASE_PATH=app.db
LOG_LEVEL=DEBUG
NOTIFICATION_AUTH_TOKEN=my-secret-token

SENDGRID_USERNAME=apikey
SENDGRID_PASSWORD=YOUR_SENDGRID_API_KEY
FROM_EMAIL=support@rsvp.mprlab.com
```

Then export them:

```bash
export $(cat .env | xargs)
```

---

## Running

1. **Install Dependencies**:

   ```bash
   go mod tidy
   ```

2. **Build**:

   ```bash
   go build -o notify ./cmd/notify
   ```

3. **Run**:

   ```bash
   ./notify
   ```

   The server listens on port `:8080` (or whatever you set via `SERVER_PORT`).

Logs will appear in your console. If you set `LOG_LEVEL=DEBUG`, you’ll see more detailed logs.

---

## Usage Examples

Below are some basic `curl` commands to interact with the service. **If** you set the `NOTIFICATION_AUTH_TOKEN`, make
sure to include the header `Authorization: Bearer <token>`.

### 1. Creating a Notification

**Endpoint**: `POST /notifications`

**Example**:

```bash
curl -X POST http://localhost:8080/notifications \
     -H "Content-Type: application/json" \
     -H "Authorization: Bearer my-secret-token" \
     -d '{
       "notification_type": "email",
       "recipient": "someone@example.com",
       "subject": "Test Subject",
       "message": "Hello from the notification service"
     }'
```

**Response** (JSON):

```json
{
  "notification_id": "9bb604f8-ea1f-4ea0-96c9-1f56e720909e",
  "notification_type": "email",
  "recipient": "someone@example.com",
  "subject": "Test Subject",
  "message": "Hello from the notification service",
  "status": "queued",
  "provider_message_id": "",
  "created_at": "2025-03-13T10:00:00Z",
  "updated_at": "2025-03-13T10:00:00Z",
  "retry_count": 0
}
```

### 2. Retrieving a Notification

**Endpoint**: `GET /notifications/{notification_id}`

**Example**:

```bash
curl -X GET http://localhost:8080/notifications/9bb604f8-ea1f-4ea0-96c9-1f56e720909e \
     -H "Authorization: Bearer my-secret-token"
```

**Response** (JSON):

```json
{
  "notification_id": "9bb604f8-ea1f-4ea0-96c9-1f56e720909e",
  "notification_type": "email",
  "recipient": "someone@example.com",
  "subject": "Test Subject",
  "message": "Hello from the notification service",
  "status": "sent",
  "provider_message_id": "sendgrid-provider-id",
  "created_at": "2025-03-13T10:00:00Z",
  "updated_at": "2025-03-13T10:00:10Z",
  "retry_count": 1
}
```

- The `status` might be `"queued"` initially, then become `"sent"` or `"failed"` after the background worker processes
  it.
- The background worker runs every `RETRY_INTERVAL_SEC` seconds and attempts to send messages with `status=queued` or
  `status=failed` (retrying up to `MAX_RETRIES` times).

---

## End-to-End Flow

1. **You** POST a new notification (an email) to the service (the record goes into SQLite with `status=queued`).
2. **The background worker** picks up queued notifications and calls SendGrid via SMTP.
3. If sending is successful:
    - The notification’s `status` becomes `sent`.
    - The `provider_message_id` might be something like `"sendgrid-provider-id"`.
4. If sending fails, the service sets `status=failed` and increments `retry_count`. On the next cycle, it tries again
   until `retry_count` reaches `MAX_RETRIES`.
5. **You** can poll `GET /notifications/{id}` to see if it’s been sent or failed.

---

## License

This project is licensed under the [MIT License](./LICENSE). Feel free to adapt for your own needs.