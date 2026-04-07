package email

import (
	"fmt"
	"html"
	"strings"
	"time"

	"azadi-go/internal/model"
)

func wrapLayout(title, bodyContent string) string {
	var b strings.Builder
	b.WriteString(`<!DOCTYPE html>
<html>
<head>
<meta charset="UTF-8">
<meta name="viewport" content="width=device-width, initial-scale=1.0">
<link href="https://fonts.googleapis.com/css2?family=Poppins:wght@400;500;600&display=swap" rel="stylesheet">
<title>`)
	b.WriteString(html.EscapeString(title))
	b.WriteString(`</title>
</head>
<body style="margin:0;padding:0;font-family:'Poppins',Arial,sans-serif;background-color:#f4f4f4;">
<table width="100%" cellpadding="0" cellspacing="0" style="max-width:600px;margin:0 auto;">
<tr><td style="background-color:#0c121d;padding:24px 32px;text-align:center;">
<h1 style="color:#ffffff;margin:0;font-size:24px;">Azadi Finance</h1>
</td></tr>
<tr><td style="background-color:#ffffff;padding:32px;">
`)
	b.WriteString(bodyContent)
	b.WriteString(`
</td></tr>
<tr><td style="background-color:#f0f0f0;padding:16px 32px;text-align:center;font-size:12px;color:#666;">
<p>This is an automated message from Azadi Finance. Please do not reply to this email.</p>
</td></tr>
</table>
</body>
</html>`)
	return b.String()
}

func PaymentConfirmationHTML(amountPence int64) string {
	amount := model.FormatPence(amountPence)
	body := fmt.Sprintf(`<h2 style="color:#0c121d;margin-top:0;">Payment Confirmed</h2>
<p>Your payment of <strong>%s</strong> has been received.</p>
<p>This payment will be reflected in your account within 2-3 business days.</p>
<p style="margin-top:24px;padding:16px;background-color:#fff3cd;border-radius:4px;">
⚠️ If you did not make this payment, please contact us immediately.
</p>`, amount)
	return wrapLayout("Payment Confirmed", body)
}

func SettlementFigureHTML(amountPence int64, validUntil time.Time) string {
	amount := model.FormatPence(amountPence)
	body := fmt.Sprintf(`<h2 style="color:#0c121d;margin-top:0;">Settlement Figure</h2>
<p>Your settlement figure is <strong>%s</strong>.</p>
<p>This figure is valid until <strong>%s</strong>.</p>
<p>To settle your agreement, please contact our team or use your online account.</p>`,
		amount, validUntil.Format("02 January 2006"))
	return wrapLayout("Settlement Figure", body)
}

func BankDetailsUpdatedHTML() string {
	body := `<h2 style="color:#0c121d;margin-top:0;">Bank Details Updated</h2>
<p>Your bank details have been successfully updated.</p>
<p>All future payments will be taken from your new account.</p>
<p style="margin-top:24px;padding:16px;background-color:#fff3cd;border-radius:4px;">
⚠️ If you did not request this change, please contact us immediately.
</p>`
	return wrapLayout("Bank Details Updated", body)
}

func PaymentDateChangedHTML(newDate string) string {
	escaped := html.EscapeString(newDate)
	body := fmt.Sprintf(`<h2 style="color:#0c121d;margin-top:0;">Payment Date Changed</h2>
<p>Your payment date has been changed to the <strong>%s</strong> of each month.</p>
<p>All future direct debit payments will be collected on this date.</p>
<p style="margin-top:24px;padding:16px;background-color:#fff3cd;border-radius:4px;">
⚠️ If you did not request this change, please contact us immediately.
</p>`, escaped)
	return wrapLayout("Payment Date Changed", body)
}

func LoginAlertHTML(ipAddress string) string {
	escaped := html.EscapeString(ipAddress)
	body := fmt.Sprintf(`<h2 style="color:#0c121d;margin-top:0;">Login Alert</h2>
<p>A new login to your account has been detected.</p>
<table style="margin:16px 0;border:1px solid #ddd;border-radius:4px;width:100%%;">
<tr><td style="padding:8px 16px;font-weight:600;">IP Address</td><td style="padding:8px 16px;">%s</td></tr>
</table>
<p>If this was not you, we recommend changing your credentials immediately.</p>`, escaped)
	return wrapLayout("Login Alert", body)
}
