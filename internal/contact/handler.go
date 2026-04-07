package contact

import (
	"context"
	"net/http"

	"azadi-go/internal/auth"
	"azadi-go/internal/httputil"
	"azadi-go/internal/model"
)

type handlerService interface {
	GetCustomer(ctx context.Context, customerID string) (*model.Customer, error)
	UpdateContactDetails(ctx context.Context, customerID string, phone, mobilePhone, email, addressLine1, addressLine2, city, postcode, ipAddress, sessionID string) error
}

type Handler struct {
	service  handlerService
	sessions *auth.SessionStore
	flashes  *auth.FlashStore
	render   httputil.RenderFunc
}

func NewHandler(service handlerService, sessions *auth.SessionStore, flashes *auth.FlashStore,
	render httputil.RenderFunc) *Handler {
	return &Handler{service: service, sessions: sessions, flashes: flashes, render: render}
}

func (h *Handler) ContactPage(w http.ResponseWriter, r *http.Request) {
	customer := auth.CustomerFromContext(r.Context())
	c, err := h.service.GetCustomer(r.Context(), customer.CustomerID)
	if err != nil {
		httputil.ServerError(w, "failed to get customer", err)
		return
	}
	h.render(w, r, "my-contact-details.html", map[string]any{
		"CustomerData": c,
	})
}

func (h *Handler) UpdateContact(w http.ResponseWriter, r *http.Request) {
	customer := auth.CustomerFromContext(r.Context())
	sessionID := auth.SessionIDFromContext(r.Context())

	if !httputil.RequireForm(w, r) {
		return
	}

	// Get existing customer to merge partial updates
	existing, err := h.service.GetCustomer(r.Context(), customer.CustomerID)
	if err != nil {
		httputil.ServerError(w, "failed to get customer", err)
		return
	}

	// Merge: use form value if non-empty, else keep existing
	phone := nonEmpty(r.FormValue("homePhone"), existing.Phone)
	mobilePhone := nonEmpty(r.FormValue("mobilePhone"), existing.MobilePhone)
	email := nonEmpty(r.FormValue("email"), existing.Email)
	addressLine1 := nonEmpty(r.FormValue("houseName"), existing.AddressLine1)
	postcode := nonEmpty(r.FormValue("postcode"), existing.Postcode)

	err = h.service.UpdateContactDetails(r.Context(), customer.CustomerID,
		phone, mobilePhone, email, addressLine1, existing.AddressLine2, existing.City, postcode,
		r.RemoteAddr, sessionID)
	if err != nil {
		httputil.ServerError(w, "failed to update contact", err)
		return
	}

	h.flashes.Set(sessionID, "success", "Contact details updated successfully.")
	http.Redirect(w, r, "/my-contact-details", http.StatusFound)
}

func nonEmpty(value, fallback string) string {
	if value != "" {
		return value
	}
	return fallback
}
