package email

import "context"

// Sender sends emails.
type Sender interface {
	SendEmail(ctx context.Context, m Email) error
	SendTemplatedEmail(ctx context.Context, m TemplatedEmail) error
}
