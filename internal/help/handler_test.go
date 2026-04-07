package help

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func testRender(w http.ResponseWriter, _ *http.Request, name string, _ map[string]any) {
	w.Header().Set("X-Template", name)
	w.WriteHeader(http.StatusOK)
}

func TestHelpHandler_AllPages(t *testing.T) {
	h := NewHandler(testRender)

	tests := []struct {
		name     string
		handler  func(http.ResponseWriter, *http.Request)
		template string
	}{
		{"FAQs", h.FAQs, "help/faqs.html"},
		{"WaysToPay", h.WaysToPay, "help/ways-to-pay.html"},
		{"ContactUs", h.ContactUs, "help/contact-us.html"},
		{"Cookies", h.Cookies, "legal/cookies.html"},
		{"Privacy", h.Privacy, "legal/privacy.html"},
		{"Terms", h.Terms, "legal/terms.html"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			w := httptest.NewRecorder()
			tt.handler(w, req)
			if w.Code != http.StatusOK {
				t.Errorf("status = %d, want 200", w.Code)
			}
			if got := w.Header().Get("X-Template"); got != tt.template {
				t.Errorf("template = %q, want %q", got, tt.template)
			}
		})
	}
}
