package support

import (
	"context"
	"strings"
	"time"

	"labshop/internal/domain/errs"
	domainsupport "labshop/internal/domain/support"
	"labshop/internal/usecase/port"
)

type CreateTicketInput struct {
	Subject        string
	InitialMessage string
}

type Usecase struct {
	supportRepo port.SupportRepository
	idGen       port.IDGenerator
}

func New(supportRepo port.SupportRepository, idGen port.IDGenerator) *Usecase {
	return &Usecase{
		supportRepo: supportRepo,
		idGen:       idGen,
	}
}

func (u *Usecase) CreateTicket(ctx context.Context, userID, userEmail string, in CreateTicketInput) (*domainsupport.Ticket, error) {
	subject := strings.TrimSpace(in.Subject)
	if subject == "" {
		return nil, errs.NewValidation("subject is required", errs.FieldViolation{Field: "subject", Issue: "cannot be empty"})
	}
	initMsg := strings.TrimSpace(in.InitialMessage)
	if initMsg == "" {
		return nil, errs.NewValidation("initial_message is required", errs.FieldViolation{Field: "initial_message", Issue: "cannot be empty"})
	}

	now := time.Now()
	ticketID := u.idGen.NewID()
	firstMsg := domainsupport.TicketMessage{
		ID:         u.idGen.NewID(),
		TicketID:   ticketID,
		SenderID:   userID,
		SenderRole: "customer",
		Message:    initMsg,
		CreatedAt:  now,
	}

	ticket := &domainsupport.Ticket{
		ID:          ticketID,
		UserID:      userID,
		UserEmail:   userEmail,
		Subject:     subject,
		Status:      domainsupport.StatusOpen,
		Messages:    []domainsupport.TicketMessage{firstMsg},
		Attachments: make([]domainsupport.TicketAttachment, 0),
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	if err := u.supportRepo.CreateTicket(ctx, ticket); err != nil {
		return nil, err
	}

	return ticket, nil
}

func (u *Usecase) ListTickets(ctx context.Context, userID string, limit, offset int) ([]domainsupport.Ticket, int, error) {
	if limit <= 0 {
		limit = 20
	} else if limit > 100 {
		limit = 100
	}
	if offset < 0 {
		offset = 0
	}

	return u.supportRepo.ListTickets(ctx, userID, limit, offset)
}

func (u *Usecase) GetTicket(ctx context.Context, userID, ticketID string, isAdmin bool) (*domainsupport.Ticket, error) {
	ticket, err := u.supportRepo.GetTicketByID(ctx, ticketID)
	if err != nil || ticket == nil {
		return nil, errs.NewNotFound("ticket", ticketID)
	}

	if !isAdmin && ticket.UserID != userID {
		return nil, errs.NewNotFound("ticket", ticketID)
	}

	return ticket, nil
}

func (u *Usecase) AddMessage(ctx context.Context, userID, userRole, ticketID, message string) (*domainsupport.TicketMessage, error) {
	isAdmin := userRole == "admin"
	ticket, err := u.GetTicket(ctx, userID, ticketID, isAdmin)
	if err != nil {
		return nil, err
	}

	msg := domainsupport.TicketMessage{
		ID:         u.idGen.NewID(),
		TicketID:   ticketID,
		SenderID:   userID,
		SenderRole: userRole,
		Message:    strings.TrimSpace(message),
		CreatedAt:  time.Now(),
	}

	if err := ticket.AddMessage(msg); err != nil {
		return nil, err
	}

	if err := u.supportRepo.AddMessage(ctx, &msg); err != nil {
		return nil, err
	}

	if err := u.supportRepo.UpdateTicketStatus(ctx, ticket.ID, ticket.Status); err != nil {
		return nil, err
	}

	return &msg, nil
}

func (u *Usecase) AddAttachment(ctx context.Context, userID, ticketID string, att domainsupport.TicketAttachment, isAdmin bool) (*domainsupport.TicketAttachment, error) {
	ticket, err := u.GetTicket(ctx, userID, ticketID, isAdmin)
	if err != nil {
		return nil, err
	}

	att.ID = u.idGen.NewID()
	att.TicketID = ticketID
	att.UploadedAt = time.Now()

	if err := ticket.AddAttachment(att); err != nil {
		return nil, err
	}

	if err := u.supportRepo.AddAttachment(ctx, &att); err != nil {
		return nil, err
	}

	return &att, nil
}
