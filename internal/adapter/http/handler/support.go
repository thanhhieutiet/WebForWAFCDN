package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"

	"labshop/internal/adapter/http/middleware"
	"labshop/internal/adapter/http/presenter"
	"labshop/internal/domain/errs"
	"labshop/internal/domain/support"
	supportuc "labshop/internal/usecase/support"
)

type SupportHandler struct {
	supportUC *supportuc.Usecase
}

func NewSupportHandler(supportUC *supportuc.Usecase) *SupportHandler {
	return &SupportHandler{supportUC: supportUC}
}

type CreateTicketRequest struct {
	Subject        string `json:"subject"`
	InitialMessage string `json:"initial_message"`
}

type AddTicketMessageRequest struct {
	Message string `json:"message"`
}

type TicketAttachmentDTO struct {
	ID          string    `json:"id"`
	FileName    string    `json:"file_name"`
	FileSize    int64     `json:"file_size"`
	SHA256      string    `json:"sha256"`
	ContentType string    `json:"content_type"`
	UploadedAt  time.Time `json:"uploaded_at"`
}

type TicketMessageDTO struct {
	ID         string    `json:"id"`
	SenderID   string    `json:"sender_id"`
	SenderRole string    `json:"sender_role"`
	Message    string    `json:"message"`
	CreatedAt  time.Time `json:"created_at"`
}

type TicketDTO struct {
	ID          string                `json:"id"`
	UserID      string                `json:"user_id"`
	UserEmail   string                `json:"user_email"`
	Subject     string                `json:"subject"`
	Status      string                `json:"status"`
	Messages    []TicketMessageDTO    `json:"messages"`
	Attachments []TicketAttachmentDTO `json:"attachments"`
	CreatedAt   time.Time             `json:"created_at"`
	UpdatedAt   time.Time             `json:"updated_at"`
}

func toTicketDTO(t *support.Ticket) TicketDTO {
	msgs := make([]TicketMessageDTO, 0, len(t.Messages))
	for _, m := range t.Messages {
		msgs = append(msgs, TicketMessageDTO{
			ID:         m.ID,
			SenderID:   m.SenderID,
			SenderRole: m.SenderRole,
			Message:    m.Message,
			CreatedAt:  m.CreatedAt,
		})
	}

	atts := make([]TicketAttachmentDTO, 0, len(t.Attachments))
	for _, a := range t.Attachments {
		atts = append(atts, TicketAttachmentDTO{
			ID:          a.ID,
			FileName:    a.FileName,
			FileSize:    a.FileSize,
			SHA256:      a.SHA256,
			ContentType: a.ContentType,
			UploadedAt:  a.UploadedAt,
		})
	}

	return TicketDTO{
		ID:          t.ID,
		UserID:      t.UserID,
		UserEmail:   t.UserEmail,
		Subject:     t.Subject,
		Status:      string(t.Status),
		Messages:    msgs,
		Attachments: atts,
		CreatedAt:   t.CreatedAt,
		UpdatedAt:   t.UpdatedAt,
	}
}

type ListTicketsResponse struct {
	Items    []TicketDTO `json:"items"`
	Total    int         `json:"total"`
	Page     int         `json:"page"`
	PageSize int         `json:"page_size"`
}

func (h *SupportHandler) CreateTicket(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	var req CreateTicketRequest
	if err := decodeJSON(w, r, &req); err != nil {
		presenter.Error(w, r, err)
		return
	}

	ticket, err := h.supportUC.CreateTicket(r.Context(), claims.UserID, claims.Email, supportuc.CreateTicketInput{
		Subject:        req.Subject,
		InitialMessage: req.InitialMessage,
	})
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	presenter.JSON(w, http.StatusCreated, toTicketDTO(ticket))
}

func (h *SupportHandler) ListTickets(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	q := r.URL.Query()

	page, _ := strconv.Atoi(q.Get("page"))
	if page <= 0 {
		page = 1
	}

	pageSize, _ := strconv.Atoi(q.Get("page_size"))
	if pageSize <= 0 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize
	tickets, total, err := h.supportUC.ListTickets(r.Context(), claims.UserID, pageSize, offset)
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	dtos := make([]TicketDTO, 0, len(tickets))
	for _, t := range tickets {
		dtos = append(dtos, toTicketDTO(&t))
	}

	presenter.JSON(w, http.StatusOK, ListTicketsResponse{
		Items:    dtos,
		Total:    total,
		Page:     page,
		PageSize: pageSize,
	})
}

func (h *SupportHandler) GetTicket(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	ticketID := chi.URLParam(r, "id")

	ticket, err := h.supportUC.GetTicket(r.Context(), claims.UserID, ticketID, claims.Role == "admin")
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	presenter.JSON(w, http.StatusOK, toTicketDTO(ticket))
}

func (h *SupportHandler) AddMessage(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	ticketID := chi.URLParam(r, "id")

	var req AddTicketMessageRequest
	if err := decodeJSON(w, r, &req); err != nil {
		presenter.Error(w, r, err)
		return
	}

	msg, err := h.supportUC.AddMessage(r.Context(), claims.UserID, string(claims.Role), ticketID, req.Message)
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	presenter.JSON(w, http.StatusCreated, TicketMessageDTO{
		ID:         msg.ID,
		SenderID:   msg.SenderID,
		SenderRole: msg.SenderRole,
		Message:    msg.Message,
		CreatedAt:  msg.CreatedAt,
	})
}

func (h *SupportHandler) AddAttachment(w http.ResponseWriter, r *http.Request) {
	claims := middleware.GetClaims(r.Context())
	ticketID := chi.URLParam(r, "id")

	// Limit request body to 5MB to handle multipart overhead for a 2MB file
	r.Body = http.MaxBytesReader(w, r.Body, 5<<20)
	if err := r.ParseMultipartForm(5 << 20); err != nil {
		presenter.Error(w, r, errs.NewValidation("multipart form too large or invalid"))
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		presenter.Error(w, r, errs.NewValidation("file is required in form-data", errs.FieldViolation{Field: "file", Issue: "file part missing"}))
		return
	}
	defer file.Close()

	if header.Size > support.MaxAttachmentSizeBytes {
		presenter.Error(w, r, errs.NewValidation("file size exceeds 2MB limit", errs.FieldViolation{Field: "file", Issue: "max 2MB"}))
		return
	}

	// Compute SHA256 and verify contents
	hasher := sha256.New()
	written, err := io.Copy(hasher, file)
	if err != nil {
		presenter.Error(w, r, errs.NewValidation("failed to process uploaded file"))
		return
	}

	sha256Hex := hex.EncodeToString(hasher.Sum(nil))
	sanitizedName := filepath.Base(strings.TrimSpace(header.Filename))

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	att := support.TicketAttachment{
		FileName:    sanitizedName,
		FileSize:    written,
		SHA256:      sha256Hex,
		ContentType: contentType,
	}

	savedAtt, err := h.supportUC.AddAttachment(r.Context(), claims.UserID, ticketID, att, claims.Role == "admin")
	if err != nil {
		presenter.Error(w, r, err)
		return
	}

	presenter.JSON(w, http.StatusCreated, TicketAttachmentDTO{
		ID:          savedAtt.ID,
		FileName:    savedAtt.FileName,
		FileSize:    savedAtt.FileSize,
		SHA256:      savedAtt.SHA256,
		ContentType: savedAtt.ContentType,
		UploadedAt:  savedAtt.UploadedAt,
	})
}
