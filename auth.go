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
