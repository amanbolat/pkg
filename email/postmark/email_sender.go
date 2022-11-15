package postmarkapp

import (
	"context"
	"errors"
	"fmt"
	"github.com/amanbolat/pkg/email"
	"github.com/keighl/postmark"
	"net/http"
	"time"
)

var (
	ErrFailedSendEmail           = errors.New("postmark email: failed to send email")
	ErrFailedSendTemaplatedEmail = errors.New("postmark email: failed to send templated email")
)

type emailSender struct {
	defaultFromSignature string
	client               *postmark.Client
}

func NewPostmarkAppEmailSender(serverToken string, requestTimeout time.Duration) email.Sender {
	pmClient := postmark.NewClient(serverToken, "")
	pmClient.HTTPClient = &http.Client{
		Timeout: requestTimeout,
	}
	return &emailSender{
		defaultFromSignature: "",
		client:               pmClient,
	}
}

func (e *emailSender) SendEmail(ctx context.Context, m email.Email) error {
	from := e.defaultFromSignature
	if m.From != "" {
		from = m.From
	}
	msg := postmark.Email{
		From:       from,
		To:         m.To,
		Cc:         m.Cc,
		Bcc:        m.Bcc,
		Subject:    m.Subject,
		HtmlBody:   m.HtmlBody,
		TextBody:   m.TextBody,
		TrackOpens: true,
	}
	_, err := e.client.SendEmail(msg)
	if err != nil {
		return fmt.Errorf("%v: %w", ErrFailedSendEmail, err)
	}

	return nil
}

func (e *emailSender) SendTemplatedEmail(ctx context.Context, m email.TemplatedEmail) error {
	from := e.defaultFromSignature
	if m.From != "" {
		from = m.From
	}

	msg := postmark.TemplatedEmail{
		TemplateId:    m.TemplateId,
		TemplateModel: m.TemplateData,
		From:          from,
		To:            m.To,
		Cc:            m.Cc,
		Bcc:           m.Bcc,
	}

	res, err := e.client.SendTemplatedEmail(msg)
	if err != nil {
		return fmt.Errorf("%v, %w", ErrFailedSendTemaplatedEmail, err)
	}
	if res.ErrorCode != 0 {
		return fmt.Errorf("%v: %d, %s", ErrFailedSendTemaplatedEmail, res.ErrorCode, res.Message)
	}

	return nil
}
