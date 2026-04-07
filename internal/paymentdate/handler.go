package paymentdate

import (
	"context"
	"net/http"
	"strconv"

	"azadi-go/internal/auth"
	"azadi-go/internal/domain"
	"azadi-go/internal/httputil"
	"azadi-go/internal/model"
)

type paymentDateService interface {
	ChangePaymentDate(ctx context.Context, customerID string, agreementID int64, newDay int, ipAddress, sessionID string) error
}

type Handler struct {
	service          paymentDateService
	agreementService domain.AgreementLister
	sessions         *auth.SessionStore
	flashes          *auth.FlashStore
	render           httputil.RenderFunc
}

func NewHandler(service paymentDateService, agreementSvc domain.AgreementLister, sessions *auth.SessionStore,
	flashes *auth.FlashStore, render httputil.RenderFunc) *Handler {
	return &Handler{
		service: service, agreementService: agreementSvc,
		sessions: sessions, flashes: flashes, render: render,
	}
}

func (h *Handler) ChangeDatePage(w http.ResponseWriter, r *http.Request) {
	customer := auth.CustomerFromContext(r.Context())
	agreements, err := h.agreementService.GetAgreementsForCustomer(r.Context(), customer.CustomerID)
	if err != nil {
		httputil.ServerError(w, "failed to get agreements", err)
		return
	}

	data := map[string]any{"Agreements": agreements}
	if len(agreements) > 0 {
		first := agreements[0]
		day := GetCurrentPaymentDay(first)
		data["CurrentPaymentDay"] = day
		data["CurrentPaymentDate"] = model.DayWithSuffix(day)
		data["AlreadyChanged"] = first.PaymentDateChanged
	}
	h.render(w, r, "finance/change-payment-date.html", data)
}

func (h *Handler) ChangeDate(w http.ResponseWriter, r *http.Request) {
	customer := auth.CustomerFromContext(r.Context())
	sessionID := auth.SessionIDFromContext(r.Context())

	if !httputil.RequireForm(w, r) {
		return
	}

	newDay, err := strconv.Atoi(r.FormValue("newPaymentDate"))
	if err != nil {
		httputil.BadRequest(w, "Bad Request")
		return
	}
	agreementID, err := strconv.ParseInt(r.FormValue("agreementId"), 10, 64)
	if err != nil {
		httputil.BadRequest(w, "Bad Request")
		return
	}

	err = h.service.ChangePaymentDate(r.Context(), customer.CustomerID, agreementID, newDay, r.RemoteAddr, sessionID)
	if err != nil {
		h.flashes.Set(sessionID, "error", err.Error())
	} else {
		h.flashes.Set(sessionID, "success", "true")
	}
	http.Redirect(w, r, "/finance/change-payment-date", http.StatusFound)
}
