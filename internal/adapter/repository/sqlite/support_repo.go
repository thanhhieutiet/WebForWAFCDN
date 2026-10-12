package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"labshop/internal/domain/support"
	"labshop/internal/usecase/port"
)

type SupportRepository struct {
	db *DB
}

var _ port.SupportRepository = (*SupportRepository)(nil)

func NewSupportRepository(db *DB) *SupportRepository {
	return &SupportRepository{db: db}
}

func (r *SupportRepository) CreateTicket(ctx context.Context, t *support.Ticket) error {
	queryTicket := `
		INSERT INTO tickets (id, user_id, user_email, subject, status, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.exec(ctx).ExecContext(ctx, queryTicket,
		t.ID, t.UserID, t.UserEmail, t.Subject, string(t.Status),
		t.CreatedAt.UTC().Format(time.RFC3339),
		t.UpdatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to insert ticket: %w", err)
	}

	for _, msg := range t.Messages {
		if err := r.AddMessage(ctx, &msg); err != nil {
			return err
		}
	}

	return nil
}

func (r *SupportRepository) loadMessages(ctx context.Context, ticketID string) ([]support.TicketMessage, error) {
	rows, err := r.db.exec(ctx).QueryContext(ctx,
		"SELECT id, ticket_id, sender_id, sender_role, message, created_at FROM ticket_messages WHERE ticket_id = ? ORDER BY created_at ASC",
		ticketID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var msgs []support.TicketMessage
	for rows.Next() {
		var (
			m         support.TicketMessage
			createdAt string
		)
		if err := rows.Scan(&m.ID, &m.TicketID, &m.SenderID, &m.SenderRole, &m.Message, &createdAt); err != nil {
			return nil, err
		}
		m.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		msgs = append(msgs, m)
	}
	return msgs, nil
}

func (r *SupportRepository) loadAttachments(ctx context.Context, ticketID string) ([]support.TicketAttachment, error) {
	rows, err := r.db.exec(ctx).QueryContext(ctx,
		"SELECT id, ticket_id, file_name, file_size, sha256, content_type, uploaded_at FROM ticket_attachments WHERE ticket_id = ?",
		ticketID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var atts []support.TicketAttachment
	for rows.Next() {
		var (
			a          support.TicketAttachment
			uploadedAt string
		)
		if err := rows.Scan(&a.ID, &a.TicketID, &a.FileName, &a.FileSize, &a.SHA256, &a.ContentType, &uploadedAt); err != nil {
			return nil, err
		}
		a.UploadedAt, _ = time.Parse(time.RFC3339, uploadedAt)
		atts = append(atts, a)
	}
	return atts, nil
}

func (r *SupportRepository) GetTicketByID(ctx context.Context, id string) (*support.Ticket, error) {
	query := `
		SELECT id, user_id, user_email, subject, status, created_at, updated_at
		FROM tickets WHERE id = ?
	`
	var (
		t         support.Ticket
		statusStr string
		createdAt string
		updatedAt string
	)
	err := r.db.exec(ctx).QueryRowContext(ctx, query, id).Scan(
		&t.ID, &t.UserID, &t.UserEmail, &t.Subject, &statusStr, &createdAt, &updatedAt,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("ticket not found")
	}
	if err != nil {
		return nil, err
	}

	t.Status = support.TicketStatus(statusStr)
	t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
	t.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

	msgs, err := r.loadMessages(ctx, t.ID)
	if err == nil {
		t.Messages = msgs
	}

	atts, err := r.loadAttachments(ctx, t.ID)
	if err == nil {
		t.Attachments = atts
	}

	return &t, nil
}

func (r *SupportRepository) ListTickets(ctx context.Context, userID string, limit, offset int) ([]support.Ticket, int, error) {
	var (
		whereSQL string
		args     []any
	)
	if userID != "" {
		whereSQL = "WHERE user_id = ?"
		args = append(args, userID)
	}

	var total int
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tickets %s", whereSQL)
	if err := r.db.exec(ctx).QueryRowContext(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, user_email, subject, status, created_at, updated_at
		FROM tickets
		%s
		ORDER BY updated_at DESC
		LIMIT ? OFFSET ?
	`, whereSQL)

	queryArgs := append(args, limit, offset)
	rows, err := r.db.exec(ctx).QueryContext(ctx, query, queryArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var tickets []support.Ticket
	for rows.Next() {
		var (
			t         support.Ticket
			statusStr string
			createdAt string
			updatedAt string
		)
		if err := rows.Scan(&t.ID, &t.UserID, &t.UserEmail, &t.Subject, &statusStr, &createdAt, &updatedAt); err != nil {
			return nil, 0, err
		}
		t.Status = support.TicketStatus(statusStr)
		t.CreatedAt, _ = time.Parse(time.RFC3339, createdAt)
		t.UpdatedAt, _ = time.Parse(time.RFC3339, updatedAt)

		msgs, _ := r.loadMessages(ctx, t.ID)
		t.Messages = msgs
		atts, _ := r.loadAttachments(ctx, t.ID)
		t.Attachments = atts

		tickets = append(tickets, t)
	}

	return tickets, total, nil
}

func (r *SupportRepository) AddMessage(ctx context.Context, msg *support.TicketMessage) error {
	query := `
		INSERT INTO ticket_messages (id, ticket_id, sender_id, sender_role, message, created_at)
		VALUES (?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.exec(ctx).ExecContext(ctx, query,
		msg.ID, msg.TicketID, msg.SenderID, msg.SenderRole, msg.Message,
		msg.CreatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to insert ticket message: %w", err)
	}
	return nil
}

func (r *SupportRepository) AddAttachment(ctx context.Context, att *support.TicketAttachment) error {
	query := `
		INSERT INTO ticket_attachments (id, ticket_id, file_name, file_size, sha256, content_type, uploaded_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`
	_, err := r.db.exec(ctx).ExecContext(ctx, query,
		att.ID, att.TicketID, att.FileName, att.FileSize, att.SHA256, att.ContentType,
		att.UploadedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return fmt.Errorf("failed to insert ticket attachment: %w", err)
	}
	return nil
}

func (r *SupportRepository) UpdateTicketStatus(ctx context.Context, ticketID string, status support.TicketStatus) error {
	query := `UPDATE tickets SET status = ?, updated_at = ? WHERE id = ?`
	_, err := r.db.exec(ctx).ExecContext(ctx, query, string(status), time.Now().UTC().Format(time.RFC3339), ticketID)
	if err != nil {
		return fmt.Errorf("failed to update ticket status: %w", err)
	}
	return nil
}
