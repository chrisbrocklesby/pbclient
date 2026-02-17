package pbclient

import (
	"fmt"
	"net/http"
	"net/url"
)

type authResponse struct {
	Token string `json:"token"`
}

func (c *Client) LoginFromConfig(cfg Config) error {
	if cfg.UserEmail != "" && cfg.UserPassword != "" {
		col := cfg.UserCollection
		if col == "" {
			col = "users"
		}
		return c.LoginUser(col, cfg.UserEmail, cfg.UserPassword)
	}
	if cfg.AdminEmail != "" && cfg.AdminPassword != "" {
		return c.LoginAdmin(cfg.AdminEmail, cfg.AdminPassword)
	}
	if cfg.SuperEmail != "" && cfg.SuperPassword != "" {
		return c.LoginSuperAdmin(cfg.SuperEmail, cfg.SuperPassword)
	}
	return nil
}

func (c *Client) LoginUser(collection, email, password string) error {
	var out authResponse
	err := c.doJSON(
		http.MethodPost,
		fmt.Sprintf("/api/collections/%s/auth-with-password", url.PathEscape(collection)),
		"",
		map[string]any{"identity": email, "password": password},
		&out,
	)
	if err != nil {
		return err
	}
	if out.Token == "" {
		return fmt.Errorf("pbclient: empty token from user login")
	}
	c.SetToken(out.Token)
	return nil
}

// Package-level convenience wrappers (use default client).
func LoginUser(collection, email, password string) error {
	return mustDefault().LoginUser(collection, email, password)
}

// RequestVerification triggers a PocketBase verification email for the auth record.
// If collection is empty, it defaults to "users".
func (c *Client) RequestVerification(collection, email string) error {
	if collection == "" {
		collection = "users"
	}
	return c.doJSON(
		http.MethodPost,
		fmt.Sprintf("/api/collections/%s/request-verification", url.PathEscape(collection)),
		"",
		map[string]any{"email": email},
		nil,
	)
}

// ConfirmVerification completes the verification flow with the provided token.
// If collection is empty, it defaults to "users".
func (c *Client) ConfirmVerification(collection, token string) error {
	if collection == "" {
		collection = "users"
	}
	return c.doJSON(
		http.MethodPost,
		fmt.Sprintf("/api/collections/%s/confirm-verification", url.PathEscape(collection)),
		"",
		map[string]any{"token": token},
		nil,
	)
}

// RequestVerification triggers verification email using the default client.
func RequestVerification(collection, email string) error {
	return mustDefault().RequestVerification(collection, email)
}

// ConfirmVerification completes verification using the default client.
func ConfirmVerification(collection, token string) error {
	return mustDefault().ConfirmVerification(collection, token)
}

// AuthRefresh validates and refreshes the current auth token for a collection record.
// If collection is empty, it defaults to "users".
func (c *Client) AuthRefresh(collection string) error {
	if collection == "" {
		collection = "users"
	}
	var out authResponse
	err := c.doJSON(
		http.MethodPost,
		fmt.Sprintf("/api/collections/%s/auth-refresh", url.PathEscape(collection)),
		"",
		nil,
		&out,
	)
	if err != nil {
		return err
	}
	if out.Token != "" {
		c.SetToken(out.Token)
	}
	return nil
}

// AuthRefresh refreshes the current token using the default client.
func AuthRefresh(collection string) error {
	return mustDefault().AuthRefresh(collection)
}

// AdminRefresh refreshes the admin/superuser token.
func (c *Client) AdminRefresh() error {
	var out authResponse
	err := c.doJSON(
		http.MethodPost,
		"/api/admins/auth-refresh",
		"",
		nil,
		&out,
	)
	if err != nil {
		return err
	}
	if out.Token != "" {
		c.SetToken(out.Token)
	}
	return nil
}

// AdminRefresh refreshes admin token using the default client.
func AdminRefresh() error { return mustDefault().AdminRefresh() }

// AuthMethods lists available auth methods (password, OAuth2 providers, MFA flags).
// If collection is empty, it defaults to "users".
func (c *Client) AuthMethods(collection string) (map[string]any, error) {
	if collection == "" {
		collection = "users"
	}
	var out map[string]any
	err := c.doJSON(
		http.MethodGet,
		fmt.Sprintf("/api/collections/%s/auth-methods", url.PathEscape(collection)),
		"",
		nil,
		&out,
	)
	return out, err
}

// AuthMethods lists auth methods using the default client.
func AuthMethods(collection string) (map[string]any, error) {
	return mustDefault().AuthMethods(collection)
}

// AuthWithOAuth2 exchanges an OAuth2 code for a PocketBase token.
// Pass extra to include optional fields like "codeChallenge", "createData", etc.
// If collection is empty, it defaults to "users".
func (c *Client) AuthWithOAuth2(collection, provider, code, codeVerifier, redirectURL string, extra map[string]any) error {
	if collection == "" {
		collection = "users"
	}
	body := map[string]any{
		"provider":     provider,
		"code":         code,
		"codeVerifier": codeVerifier,
		"redirectUrl":  redirectURL,
	}
	for k, v := range extra {
		body[k] = v
	}
	var out authResponse
	err := c.doJSON(
		http.MethodPost,
		fmt.Sprintf("/api/collections/%s/auth-with-oauth2", url.PathEscape(collection)),
		"",
		body,
		&out,
	)
	if err != nil {
		return err
	}
	if out.Token == "" {
		return fmt.Errorf("pbclient: empty token from oauth2 auth")
	}
	c.SetToken(out.Token)
	return nil
}

// AuthWithOAuth2 exchanges an OAuth2 code using the default client.
func AuthWithOAuth2(collection, provider, code, codeVerifier, redirectURL string, extra map[string]any) error {
	return mustDefault().AuthWithOAuth2(collection, provider, code, codeVerifier, redirectURL, extra)
}

// RequestOTP sends a one-time password to the user email.
// If collection is empty, it defaults to "users".
func (c *Client) RequestOTP(collection, email string) error {
	if collection == "" {
		collection = "users"
	}
	return c.doJSON(
		http.MethodPost,
		fmt.Sprintf("/api/collections/%s/request-otp", url.PathEscape(collection)),
		"",
		map[string]any{"email": email},
		nil,
	)
}

// RequestOTP sends OTP using the default client.
func RequestOTP(collection, email string) error { return mustDefault().RequestOTP(collection, email) }

// AuthWithOTP completes OTP authentication and sets the bearer token.
// If collection is empty, it defaults to "users".
func (c *Client) AuthWithOTP(collection, identity, otp, mfaToken string) error {
	if collection == "" {
		collection = "users"
	}
	body := map[string]any{
		"identity": identity,
		"otp":      otp,
	}
	if mfaToken != "" {
		body["mfaToken"] = mfaToken
	}
	var out authResponse
	err := c.doJSON(
		http.MethodPost,
		fmt.Sprintf("/api/collections/%s/auth-with-otp", url.PathEscape(collection)),
		"",
		body,
		&out,
	)
	if err != nil {
		return err
	}
	if out.Token == "" {
		return fmt.Errorf("pbclient: empty token from otp auth")
	}
	c.SetToken(out.Token)
	return nil
}

// AuthWithOTP completes OTP auth using the default client.
func AuthWithOTP(collection, identity, otp, mfaToken string) error {
	return mustDefault().AuthWithOTP(collection, identity, otp, mfaToken)
}

// RequestPasswordReset starts the password reset flow.
// If collection is empty, it defaults to "users".
func (c *Client) RequestPasswordReset(collection, email string) error {
	if collection == "" {
		collection = "users"
	}
	return c.doJSON(
		http.MethodPost,
		fmt.Sprintf("/api/collections/%s/request-password-reset", url.PathEscape(collection)),
		"",
		map[string]any{"email": email},
		nil,
	)
}

// RequestPasswordReset using the default client.
func RequestPasswordReset(collection, email string) error {
	return mustDefault().RequestPasswordReset(collection, email)
}

// ConfirmPasswordReset completes the password reset flow.
// If passwordConfirm is empty it reuses password.
// If collection is empty, it defaults to "users".
func (c *Client) ConfirmPasswordReset(collection, token, password, passwordConfirm string) error {
	if collection == "" {
		collection = "users"
	}
	if passwordConfirm == "" {
		passwordConfirm = password
	}
	return c.doJSON(
		http.MethodPost,
		fmt.Sprintf("/api/collections/%s/confirm-password-reset", url.PathEscape(collection)),
		"",
		map[string]any{
			"token":           token,
			"password":        password,
			"passwordConfirm": passwordConfirm,
		},
		nil,
	)
}

// ConfirmPasswordReset using the default client.
func ConfirmPasswordReset(collection, token, password, passwordConfirm string) error {
	return mustDefault().ConfirmPasswordReset(collection, token, password, passwordConfirm)
}

// RequestEmailChange starts the email change flow for the authenticated record.
// If collection is empty, it defaults to "users".
func (c *Client) RequestEmailChange(collection, newEmail string) error {
	if collection == "" {
		collection = "users"
	}
	return c.doJSON(
		http.MethodPost,
		fmt.Sprintf("/api/collections/%s/request-email-change", url.PathEscape(collection)),
		"",
		map[string]any{"newEmail": newEmail},
		nil,
	)
}

// RequestEmailChange using the default client.
func RequestEmailChange(collection, newEmail string) error {
	return mustDefault().RequestEmailChange(collection, newEmail)
}

// ConfirmEmailChange finalizes the email change with the token sent by email.
// If collection is empty, it defaults to "users".
func (c *Client) ConfirmEmailChange(collection, token string) error {
	if collection == "" {
		collection = "users"
	}
	return c.doJSON(
		http.MethodPost,
		fmt.Sprintf("/api/collections/%s/confirm-email-change", url.PathEscape(collection)),
		"",
		map[string]any{"token": token},
		nil,
	)
}

// ConfirmEmailChange using the default client.
func ConfirmEmailChange(collection, token string) error {
	return mustDefault().ConfirmEmailChange(collection, token)
}

// Impersonate returns a short-lived token that impersonates another record (superuser only).
// If collection is empty, it defaults to "users".
func (c *Client) Impersonate(collection, recordID string) error {
	if collection == "" {
		collection = "users"
	}
	var out authResponse
	err := c.doJSON(
		http.MethodPost,
		fmt.Sprintf("/api/collections/%s/impersonate/%s", url.PathEscape(collection), url.PathEscape(recordID)),
		"",
		nil,
		&out,
	)
	if err != nil {
		return err
	}
	if out.Token == "" {
		return fmt.Errorf("pbclient: empty token from impersonate")
	}
	c.SetToken(out.Token)
	return nil
}

// Impersonate using the default client.
func Impersonate(collection, recordID string) error {
	return mustDefault().Impersonate(collection, recordID)
}

// FileToken returns a short-lived file token for accessing protected files.
func (c *Client) FileToken() (string, error) {
	var out authResponse
	if err := c.doJSON(http.MethodPost, "/api/files/token", "", nil, &out); err != nil {
		return "", err
	}
	if out.Token == "" {
		return "", fmt.Errorf("pbclient: empty token from file token endpoint")
	}
	return out.Token, nil
}

// FileToken returns a file token using the default client.
func FileToken() (string, error) { return mustDefault().FileToken() }

// LoginAdmin authenticates the default client as an admin/superuser.
func LoginAdmin(email, password string) error { return mustDefault().LoginAdmin(email, password) }

// LoginSuperAdmin authenticates the default client as a superuser.
func LoginSuperAdmin(email, password string) error {
	return mustDefault().LoginSuperAdmin(email, password)
}

// PocketBase v0.23+: admin/superuser lives in system auth collection `_superusers`.
func (c *Client) LoginSuperAdmin(email, password string) error {
	var out authResponse
	err := c.doJSON(
		http.MethodPost,
		"/api/collections/_superusers/auth-with-password",
		"",
		map[string]any{"identity": email, "password": password},
		&out,
	)
	if err != nil {
		return err
	}
	if out.Token == "" {
		return fmt.Errorf("pbclient: empty token from superuser login")
	}
	c.SetToken(out.Token)
	return nil
}

// LoginAdmin authenticates with the superusers endpoint first and falls back
// to the legacy admins endpoint.
func (c *Client) LoginAdmin(email, password string) error {
	if err := c.LoginSuperAdmin(email, password); err == nil {
		return nil
	}
	var out authResponse
	err := c.doJSON(
		http.MethodPost,
		"/api/admins/auth-with-password",
		"",
		map[string]any{"identity": email, "password": password},
		&out,
	)
	if err != nil {
		return err
	}
	if out.Token == "" {
		return fmt.Errorf("pbclient: empty token from admin login")
	}
	c.SetToken(out.Token)
	return nil
}
