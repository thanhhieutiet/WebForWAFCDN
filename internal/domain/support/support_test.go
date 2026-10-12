package support

import (
	"strings"
	"testing"
	"time"
)

func TestTicketAttachmentValidation(t *testing.T) {
	att := TicketAttachment{
		ID:          "att-1",
		FileName:    "error.log",
		FileSize:    1024,
		SHA256:      "dummyhash",
		ContentType: "text/plain",
	}
	if err := att.Validate(); err != nil {
		t.Fatalf("expected valid attachment: %v", err)
	}

	attBadExt := att
	attBadExt.FileName = "malware.exe"
	if err := attBadExt.Validate(); err == nil {
		t.Fatalf("expected error for forbidden extension")
	}

	attTooBig := att
	attTooBig.FileSize = 3 * 1024 * 1024
	if err := attTooBig.Validate(); err == nil {
		t.Fatalf("expected error for file > 2MB")
	}
}

func TestTicketFlow(t *testing.T) {
	ticket := &Ticket{
		ID:        "tick-1",
		UserID:    "u-1",
		Subject:   "Help with order",
		Status:    StatusOpen,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Add customer message
	err := ticket.AddMessage(TicketMessage{
		ID:         "msg-1",
		SenderID:   "u-1",
		SenderRole: "customer",
		Message:    "Hello, I need help",
	})
	if err != nil {
		t.Fatalf("failed to add message: %v", err)
	}

	// Add admin message -> status becomes ANSWERED
	err = ticket.AddMessage(TicketMessage{
		ID:         "msg-2",
		SenderID:   "admin-1",
		SenderRole: "admin",
		Message:    "We are looking into it",
	})
	if err != nil || ticket.Status != StatusAnswered {
		t.Fatalf("expected status ANSWERED, got %s (err: %v)", ticket.Status, err)
	}

	// Add customer reply -> status becomes OPEN again
	err = ticket.AddMessage(TicketMessage{
		ID:         "msg-3",
		SenderID:   "u-1",
		SenderRole: "customer",
		Message:    "Thank you",
	})
	if err != nil || ticket.Status != StatusOpen {
		t.Fatalf("expected status OPEN, got %s", ticket.Status)
	}

	// Add 3 attachments
	for i := 1; i <= 3; i++ {
		err := ticket.AddAttachment(TicketAttachment{
			ID:       "att",
			FileName: "shot.png",
			FileSize: 500,
		})
		if err != nil {
			t.Fatalf("unexpected error adding attachment %d: %v", i, err)
		}
	}

	// 4th attachment should fail
	err = ticket.AddAttachment(TicketAttachment{
		FileName: "shot4.png",
		FileSize: 500,
	})
	if err == nil {
		t.Fatalf("expected error for 4th attachment")
	}

	// Close ticket
	ticket.Close()
	if ticket.Status != StatusClosed {
		t.Fatalf("expected CLOSED status")
	}

	// Adding message to closed ticket should fail
	err = ticket.AddMessage(TicketMessage{
		SenderRole: "customer",
		Message:    "one more thing",
	})
	if err == nil || !strings.Contains(err.Error(), "closed") {
		t.Fatalf("expected error adding message to closed ticket")
	}
}
