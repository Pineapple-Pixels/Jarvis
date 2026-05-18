package middleware

import (
	"net/http"

	"jarvis/web"
)

// WebhookAuth validates the X-Webhook-Secret header against the expected secret.
// If secret is empty and allowInsecure is false, all requests are rejected with 401.
// Set ALLOW_INSECURE=true to bypass authentication in local development.
func WebhookAuth(secret string, allowInsecure bool) web.Interceptor {
	return func(req web.InterceptedRequest) web.Response {
		if secret == "" {
			if allowInsecure {
				return req.Next()
			}
			return web.NewJSONResponse(http.StatusUnauthorized, map[string]string{
				"error": "unauthorized: WEBHOOK_SECRET is not set; set ALLOW_INSECURE=true to bypass in local dev",
			})
		}

		headers, ok := req.Header("X-Webhook-Secret")
		if !ok || len(headers) == 0 || headers[0] != secret {
			return web.NewJSONResponse(http.StatusUnauthorized, map[string]string{
				"error": "unauthorized",
			})
		}

		return req.Next()
	}
}
