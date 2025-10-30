# Pinguin

Pinguin is a production‑quality notification service written in Go. It exposes a gRPC interface for sending **email** and **SMS** notifications. The service uses SQLite (via GORM) for persistent storage and runs a background worker to retry failed notifications using exponential backoff. Structured logging is provided using Go’s built‑in `slog` package.

> **Note:** This version of Pinguin is gRPC‑only; all interactions are via gRPC.

---

## Table of Contents

- [Features](#features)
- [Requirements](#requirements)
- [Installation](#installation)
- [Configuration](#configuration)
- [Running the Server](#running-the-server)
- [Using the gRPC API](#using-the-grpc-api)
  - [Command‑Line Client Test](#command-line-client-test)
  - [Using grpcurl](#using-grpcurl)
- [End-to-End Flow](#end-to-end-flow)
- [Logging and Debugging](#logging-and-debugging)
- [License](#license)

---

## Features

- **gRPC-Only API:**  
  All interactions (sending notifications, retrieving statuses) are done via a gRPC interface.

- **Email and SMS Notifications:**  
  - **Email:** Delivered via SMTP using the credentials you configure for your preferred mail provider.
  - **SMS:** Delivered using Twilio’s REST API.

- **Scheduled Delivery:**  
  Clients can provide an optional `scheduled_time` to defer dispatch until a specific timestamp. The background worker releases the notification when the scheduled time arrives.

- **Persistent Storage:**  
  Uses SQLite with GORM to store notifications and track their statuses.

- **Background Worker:**  
  Processes queued or failed notifications and retries them with exponential backoff.

- **Structured Logging:**  
  Uses Go’s `slog` package for structured logging with configurable levels.

- **Bearer Token Authentication:**  
  Secure access to the gRPC endpoints via a bearer token.

---

## Requirements

- **Go 1.21+** (tested with Go 1.24)
- An SMTP-compatible service account (any provider that supports standard SMTP)
- A Twilio account for SMS notifications (if needed)
- SQLite (or any GORM‑compatible database)

---

## Installation

Clone the repository and navigate to the project directory:

```bash
git clone https://github.com/temirov/pinguin.git
cd pinguin
```

Install dependencies:

```bash
go mod tidy
```

Build the Pinguin server:

```bash
go build -o pinguin ./cmd/server
```

---

## Configuration

Pinguin is configured via environment variables. Create a `.env` file or export the variables manually. Below is an explanation of each variable:

- **DATABASE_PATH:**  
  Path to the SQLite database file (e.g., `app.db`).

- **LOG_LEVEL:**  
  Logging level. Possible values: `DEBUG`, `INFO`, `WARN`, `ERROR`.

- **NOTIFICATION_AUTH_TOKEN:**  
  Bearer token used for authenticating gRPC requests. All clients must supply this token.

- **MAX_RETRIES:**  
  Maximum number of times the background worker will retry sending a failed notification.

- **RETRY_INTERVAL_SEC:**  
  Base interval (in seconds) between retry scans. The actual backoff is exponential.

- **SMTP_USERNAME:**  
  SMTP username provided by your email service. Some providers require the full email address.

- **SMTP_PASSWORD:**  
  SMTP password or application-specific password issued by your provider.

- **FROM_EMAIL:**  
  The email address from which notifications are sent. This must be a verified sender with your SMTP provider.

- **SMTP_HOST:**  
  The hostname of the SMTP server (e.g., `smtp.yourdomain.com`).

- **SMTP_PORT:**  
  The SMTP port. Use `587` for STARTTLS or `465` for implicit TLS; the service will initiate TLS automatically when you specify `465`.

- **TWILIO_ACCOUNT_SID:**  
  Your Twilio Account SID, used for sending SMS messages.

- **TWILIO_AUTH_TOKEN:**  
  Your Twilio Auth Token.

- **TWILIO_FROM_NUMBER:**  
  The phone number (in E.164 format) from which SMS messages are sent.

Example `.env` file:

```bash
DATABASE_PATH=app.db
LOG_LEVEL=DEBUG
NOTIFICATION_AUTH_TOKEN=my-secret-token

SMTP_USERNAME=apikey
SMTP_PASSWORD=super-secret-password
FROM_EMAIL=support@yourdomain.com
SMTP_HOST=smtp.yourdomain.com
SMTP_PORT=587

TWILIO_ACCOUNT_SID=ACxxxxxxxxxxxx
TWILIO_AUTH_TOKEN=yyyyyyyyyyyyyy
TWILIO_FROM_NUMBER=+12015550123
```

Load the environment variables:

```bash
export $(cat .env | xargs)
```

---

## Running the Server

Start the Pinguin gRPC server by running the built executable:

```bash
./pinguin
```

By default, the server listens on port `50051`. The server initializes the SQLite database, starts the background retry worker, and registers the gRPC NotificationService with bearer token authentication.

---

## Using the gRPC API

### Command‑Line Client Test

A lightweight client test application is available under `cmd/client_test/main.go`. This client wraps the gRPC calls and demonstrates sending a notification. To run the client test, use:

```bash
go run cmd/client_test/main.go --to your-email@yourdomain.com --subject "Test Email" --message "Hello, world!"
```

If successful, you will see output similar to:

```
Notification sent successfully. Notification ID: notif-1741932356116855000
```

### Using grpcurl

You can also use [grpcurl](https://github.com/fullstorydev/grpcurl) to interact directly with the gRPC API. For example, to send an email notification:

```bash
grpcurl -d '{
  "notification_type": "EMAIL",
  "recipient": "someone@example.com",
  "subject": "Test Email",
  "message": "Hello from Pinguin!",
  "scheduled_time": "2024-05-03T17:00:00Z"
}' -H "Authorization: Bearer my-secret-token" localhost:50051 pinguin.NotificationService/SendNotification
```

To retrieve the status of a notification (replace `<notification_id>` with the actual ID):

```bash
grpcurl -d '{
  "notification_id": "<notification_id>"
}' -H "Authorization: Bearer my-secret-token" localhost:50051 pinguin.NotificationService/GetNotificationStatus
```

---

## End-to-End Flow

1. **Submission:**  
   A client submits a notification (email or SMS) via gRPC using the `SendNotification` RPC. The notification is stored in the SQLite database with a status of `queued`. If `scheduled_time` is in the future, the notification remains queued until the target time.

2. **Immediate Dispatch:**  
   The server attempts to dispatch the notification immediately:
    - **Email:** Sent via SMTP using the configured credentials. When you supply port `465`, Pinguin initiates the connection over TLS before issuing SMTP commands; otherwise it uses STARTTLS on demand.
    - **SMS:** Sent using Twilio’s REST API.

3. **Background Worker:**  
   A background worker periodically polls the database for notifications that are still queued or have failed and reattempts sending them with exponential backoff.

4. **Status Retrieval:**  
   Clients can query the notification’s status using the `GetNotificationStatus` RPC until the status changes to `sent` or `failed`.

---

## Logging and Debugging

- **Structured Logging:**  
  Pinguin uses Go’s `slog` package for structured logging. Set the logging level via the `LOG_LEVEL` environment variable.

- **Debug Output:**  
  When `LOG_LEVEL` is set to `DEBUG`, detailed messages (including SMTP debug output and fallback warnings) are logged. Sensitive data (such as API keys) is masked in the logs.

---

## License

This project is proprietary software. All rights reserved by Marco Polo Research Lab.  
See the [LICENSE](./LICENSE) file for details.
