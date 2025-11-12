// Package attachments converts CLI-friendly attachment specifiers into gRPC
// EmailAttachment messages, inferring MIME types and validating payloads so
// clients can hand off files to the notification service safely.
package attachments
