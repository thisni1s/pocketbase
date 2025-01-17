package auth

import (
	"encoding/json"
	"errors"
	"strings"

	"golang.org/x/oauth2"
)

var _ Provider = (*Mailcow)(nil)

// NameMailcow is the unique name of this authentification provider
const NameMailcow string = "mailcow"

// Mailcow is a self-hosted mailserver suite
type Mailcow struct {
	*baseProvider
}

// NewMailcowProvider creates a new Mailcow provider instance with some defaults
func NewMailcowProvider() *Mailcow {
	return &Mailcow{&baseProvider{
		// Default scopes provided to any oauth2 providers
		scopes: []string{
			"profile",
		},
	}}
}

// FetchAuthUser returns a User instance based on the Mailcow user api
// Reference: https://github.com/mailcow/mailcow-dockerized/blob/master/data/web/oauth/profile.php
func (p *Mailcow) FetchAuthUser(token *oauth2.Token) (*AuthUser, error) {
	data, err := p.FetchRawUserData(token)
	if err != nil {
		return nil, err
	}
	rawUser := map[string]any{}
	if err := json.Unmarshal(data, &rawUser); err != nil {
		return nil, err
	}

	extracted := struct {
		Success     bool   `json:"success"`
		Username    string `json:"username"`
		Id          string `json:"id"`
		Identifier  string `json:"identifier"`
		Email       string `json:"email"`
		FullName    string `json:"full_name"`
		DisplayName string `json:"displayName"`
		Created     string `json:"created"`
		Modified    string `json:"modified"`
		Active      int    `json:"active"`
	}{}
	if err := json.Unmarshal(data, &extracted); err != nil {
		return nil, err
	}

	if extracted.Active != 1 {
		return nil, errors.New("User is marked as not active!")
	}

	user := &AuthUser{
		Id:           extracted.Id,
		Name:         extracted.FullName,
		Username:     extracted.Username,
		Email:        extracted.Email,
		RawUser:      rawUser,
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}

	// Mailcow usernames are usually just the email adresses, so we just take the part in front of the @
	if strings.Contains(user.Username, "@") {
		user.Username = strings.Split(user.Username, "@")[0]
	}

	return user, nil
}
