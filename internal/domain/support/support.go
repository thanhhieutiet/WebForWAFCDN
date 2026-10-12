package support

import (
	"path/filepath"
	"strings"
	"time"

	"labshop/internal/domain/errs"
)

type TicketStatus string

const (
	StatusOpen     TicketStatus = "OPEN"
	StatusAnswered TicketStatus = "ANSWERED"
	StatusClosed   TicketStatus = "CLOSED"
)

const (
	MaxAttachmentsPerTicket = 3
	MaxAttachmentSizeBytes  = 2 * 1024 * 1024 // 2MB
)

var AllowedExtensions = map[string]bool{
	".png":  true,
	".jpg":  true,
	".jpeg": true,
	".pdf":  true,
	".txt":  true,
	".log":  true,
}

// TicketAttachment stores metadata for a support file attachment.
type TicketAttachment struct {
	ID          string
	TicketID    string
	FileName    string
	FileSize    int64
	SHA256      string
	ContentType string
	UploadedAt  time.Time
}

// Validate checks file size and allowed extension.
func (a *TicketAttachment) Validate() error {
	var violations []errs.FieldViolation

	if a.FileSize <= 0 || a.FileSize > MaxAttachmentSizeBytes {
		violations = append(violations, errs.FieldViolation{Field: "file_size", Issue: "file size must be greater than 0 and at most 2MB"})
	}

	ext := strings.ToLower(filepath.Ext(a.FileName))
	if !AllowedExtensions[ext] {
		violations = append(violations, errs.FieldViolation{Field: "file_name", Issue: "file extension not allowed (allowed: png, jpg, jpeg, pdf, txt, log)"})
	}

	if len(violations) > 0 {
		return errs.NewValidation("attachment validation failed", violations...)
	}
	return nil
}

// TicketMessage represents a dialogue turn in a ticket.
type TicketMessage struct {
	ID         string
	TicketID   string
	SenderID   string
	SenderRole string // "customer" or "admin"
	Message    string
	CreatedAt  time.Time
}

// Ticket represents a customer support request.
type Ticket struct {
	ID          string
	UserID      string
	UserEmail   string
	Subject     string
	Status      TicketStatus
	Messages    []TicketMessage
	Attachments []TicketAttachment
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CanAddMessage returns true if the ticket is not closed.
func (t *Ticket) CanAddMessage() bool {
	return t.Status != StatusClosed
}

// CanAddAttachment returns true if ticket is not closed and attachment limit is not reached.
func (t *Ticket) CanAddAttachment() bool {
	return t.Status != StatusClosed && len(t.Attachments) < MaxAttachmentsPerTicket
}

// AddMessage validates and appends a new message.
func (t *Ticket) AddMessage(msg TicketMessage) error {
	if !t.CanAddMessage() {
		return errs.NewBusinessRule("ticket_closed", "cannot add message to a closed ticket")
	}
	if strings.TrimSpace(msg.Message) == "" {
		return errs.NewValidation("message is empty", errs.FieldViolation{Field: "message", Issue: "message content is required"})
	}

	t.Messages = append(t.Messages, msg)
	if msg.SenderRole == "admin" {
		t.Status = StatusAnswered
	} else if t.Status == StatusAnswered {
		t.Status = StatusOpen
	}
	t.UpdatedAt = time.Now()
	return nil
}

// AddAttachment validates and records an attachment metadata.
func (t *Ticket) AddAttachment(att TicketAttachment) error {
	if !t.CanAddAttachment() {
		return errs.NewBusinessRule("max_attachments", "ticket already has maximum attachments or is closed")
	}
	if err := att.Validate(); err != nil {
		return err
	}
	t.Attachments = append(t.Attachments, att)
	t.UpdatedAt = time.Now()
	return nil
}

// Close marks the ticket as CLOSED.
func (t *Ticket) Close() {
	t.Status = StatusClosed
	t.UpdatedAt = time.Now()
}
