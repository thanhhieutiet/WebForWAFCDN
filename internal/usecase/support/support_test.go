package support

import (
	"context"
	"errors"
	"testing"

	domainsupport "labshop/internal/domain/support"
)

type fakeSupportRepo struct {
	tickets map[string]*domainsupport.Ticket
}

func newFakeSupportRepo() *fakeSupportRepo {
	return &fakeSupportRepo{tickets: make(map[string]*domainsupport.Ticket)}
}

func (f *fakeSupportRepo) CreateTicket(ctx context.Context, t *domainsupport.Ticket) error {
	f.tickets[t.ID] = t
	return nil
}

func (f *fakeSupportRepo) GetTicketByID(ctx context.Context, id string) (*domainsupport.Ticket, error) {
	t, ok := f.tickets[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return t, nil
}

func (f *fakeSupportRepo) ListTickets(ctx context.Context, userID string, limit, offset int) ([]domainsupport.Ticket, int, error) {
	var list []domainsupport.Ticket
	for _, t := range f.tickets {
		if userID == "" || t.UserID == userID {
			list = append(list, *t)
		}
	}
	return list, len(list), nil
}

func (f *fakeSupportRepo) AddMessage(ctx context.Context, msg *domainsupport.TicketMessage) error {
	t, ok := f.tickets[msg.TicketID]
	if !ok {
		return errors.New("not found")
	}
	t.Messages = append(t.Messages, *msg)
	return nil
}

func (f *fakeSupportRepo) AddAttachment(ctx context.Context, att *domainsupport.TicketAttachment) error {
	t, ok := f.tickets[att.TicketID]
	if !ok {
		return errors.New("not found")
	}
	t.Attachments = append(t.Attachments, *att)
	return nil
}

func (f *fakeSupportRepo) UpdateTicketStatus(ctx context.Context, ticketID string, status domainsupport.TicketStatus) error {
	t, ok := f.tickets[ticketID]
	if !ok {
		return errors.New("not found")
	}
	t.Status = status
	return nil
}

type fakeIDGen struct {
	counter int
}

func (f *fakeIDGen) NewID() string {
	f.counter++
	return "id-" + string(rune('0'+f.counter))
}

func TestSupportUsecase(t *testing.T) {
	repo := newFakeSupportRepo()
	idGen := &fakeIDGen{}
	uc := New(repo, idGen)
	ctx := context.Background()

	// Create ticket
	ticket, err := uc.CreateTicket(ctx, "u-1", "user@example.com", CreateTicketInput{
		Subject:        "Payment issue",
		InitialMessage: "I was charged twice",
	})
	if err != nil {
		t.Fatalf("failed to create ticket: %v", err)
	}
	if ticket.Status != domainsupport.StatusOpen || len(ticket.Messages) != 1 {
		t.Fatalf("unexpected ticket: %+v", ticket)
	}

	// Add customer reply
	_, err = uc.AddMessage(ctx, "u-1", "customer", ticket.ID, "Any update?")
	if err != nil {
		t.Fatalf("failed to add customer message: %v", err)
	}

	// Add admin message
	_, err = uc.AddMessage(ctx, "admin-1", "admin", ticket.ID, "We refunded the extra charge.")
	if err != nil {
		t.Fatalf("failed to add admin message: %v", err)
	}
	if repo.tickets[ticket.ID].Status != domainsupport.StatusAnswered {
		t.Fatalf("expected status ANSWERED, got %s", repo.tickets[ticket.ID].Status)
	}

	// Add attachment
	att, err := uc.AddAttachment(ctx, "u-1", ticket.ID, domainsupport.TicketAttachment{
		FileName:    "receipt.pdf",
		FileSize:    2048,
		SHA256:      "sha-123",
		ContentType: "application/pdf",
	}, false)
	if err != nil || att.ID == "" {
		t.Fatalf("failed to add attachment: %v", err)
	}

	// List tickets
	list, total, err := uc.ListTickets(ctx, "u-1", 10, 0)
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatalf("failed to list tickets: %v", err)
	}

	// Create ticket with empty subject
	_, err = uc.CreateTicket(ctx, "u-1", "user@example.com", CreateTicketInput{
		Subject:        "",
		InitialMessage: "message",
	})
	if err == nil {
		t.Fatalf("expected error for empty subject")
	}

	// Create ticket with empty message
	_, err = uc.CreateTicket(ctx, "u-1", "user@example.com", CreateTicketInput{
		Subject:        "Subject",
		InitialMessage: "",
	})
	if err == nil {
		t.Fatalf("expected error for empty initial message")
	}

	// Get non-existent ticket
	_, err = uc.GetTicket(ctx, "u-1", "non-existent", false)
	if err == nil {
		t.Fatalf("expected error getting non-existent ticket")
	}

	// Get another user's ticket without admin
	_, err = uc.GetTicket(ctx, "u-other", ticket.ID, false)
	if err == nil {
		t.Fatalf("expected error getting other user's ticket")
	}

	// Add message to non-existent ticket
	_, err = uc.AddMessage(ctx, "u-1", "customer", "non-existent", "hello")
	if err == nil {
		t.Fatalf("expected error adding message to non-existent ticket")
	}

	// Add attachment to non-existent ticket
	_, err = uc.AddAttachment(ctx, "u-1", "non-existent", domainsupport.TicketAttachment{}, false)
	if err == nil {
		t.Fatalf("expected error adding attachment to non-existent ticket")
	}
}
