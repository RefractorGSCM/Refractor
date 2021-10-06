/*
 * This file is part of Refractor.
 *
 * Refractor is free software: you can redistribute it and/or modify
 * it under the terms of the GNU General Public License as published by
 * the Free Software Foundation, either version 3 of the License, or
 * (at your option) any later version.
 *
 * This program is distributed in the hope that it will be useful,
 * but WITHOUT ANY WARRANTY; without even the implied warranty of
 * MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
 * GNU General Public License for more details.
 *
 * You should have received a copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package kratos

import (
	"Refractor/domain"
	"Refractor/pkg/conf"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/labstack/echo/v4"
	kratos "github.com/ory/kratos-client-go"
	"github.com/pkg/errors"
	"net/http"
	"strings"
)

const opTag = "AuthRepo.Kratos."

type authRepo struct {
	config *conf.Config
}

func NewAuthRepo(config *conf.Config) domain.AuthRepo {
	return &authRepo{
		config: config,
	}
}

type createIdentityPayload struct {
	Schema string `json:"schema_id"`
	Traits *domain.Traits
}

func (r *authRepo) CreateUser(ctx context.Context, userTraits *domain.Traits) (*domain.AuthUser, error) {
	const op = opTag + "CreateUser"

	// Check if a user already exists with any of the provided credentials. Normally, this kind of logic is redundant
	// and does not belong in a repository, but we place it here for convenience. This does not represent a pattern in
	// the code, this is a one-off workaround because Ory Kratos does not give us any human readable conflict reporting (why?????)
	// if the email or username is already in use. It simply returns a SQL error which is no good.
	allUsers, err := r.GetAllUsers(ctx)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	validationErrors := map[string]string{}
	for _, user := range allUsers {
		// If the username is taken, send back an error
		if strings.TrimSpace(user.Traits.Username) == userTraits.Username {
			validationErrors["username"] = "This username is taken"
		}

		// If the email is taken, send back an error
		if strings.TrimSpace(user.Traits.Email) == userTraits.Email {
			validationErrors["email"] = "This email is already in use"
		}
	}

	if len(validationErrors) > 0 {
		return nil, &domain.HTTPError{
			Cause:            nil,
			Message:          "Input errors exist",
			ValidationErrors: validationErrors,
			Status:           http.StatusBadRequest,
		}
	}

	// Create the user
	url := fmt.Sprintf("%s/identities", r.config.KratosAdmin)

	payload := createIdentityPayload{
		Schema: "default",
		Traits: userTraits,
	}

	marshalled, err := json.Marshal(payload)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(marshalled))
	if err != nil {
		return nil, errors.Wrap(err, op)
	}
	req = req.WithContext(ctx)
	req.Header.Set("Content-Type", "application/json; charset=UTF-8")

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	if res.StatusCode != http.StatusCreated {
		return nil, errors.Wrap(fmt.Errorf("response status was %d and not %d", res.StatusCode, http.StatusOK), op)
	}

	identity := kratos.Identity{}
	if err := json.NewDecoder(res.Body).Decode(&identity); err != nil {
		return nil, errors.Wrap(err, op)
	}

	traitBytes, err := json.Marshal(identity.Traits)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	traits := &domain.Traits{}
	if err := json.Unmarshal(traitBytes, traits); err != nil {
		return nil, errors.Wrap(err, op)
	}

	// Create user struct
	newUser := &domain.AuthUser{
		Traits: traits,
		Session: &kratos.Session{
			Identity: identity,
		},
	}

	return newUser, nil
}

func (r *authRepo) GetUserByID(ctx context.Context, id string) (*domain.AuthUser, error) {
	const op = opTag + "GetUserByID"

	url := fmt.Sprintf("%s/identities/%s", r.config.KratosAdmin, id)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}
	req = req.WithContext(ctx)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.Wrap(fmt.Errorf("response status was %d and not %d", res.StatusCode, http.StatusOK), op)
	}

	identity := kratos.Identity{}
	if err := json.NewDecoder(res.Body).Decode(&identity); err != nil {
		return nil, errors.Wrap(err, op)
	}

	traitBytes, err := json.Marshal(identity.Traits)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	traits := &domain.Traits{}
	if err := json.Unmarshal(traitBytes, traits); err != nil {
		return nil, errors.Wrap(err, op)
	}

	user := &domain.AuthUser{
		Traits: traits,
		Session: &kratos.Session{
			Identity: identity,
		},
	}

	return user, nil
}

func (r *authRepo) GetAllUsers(ctx context.Context) ([]*domain.AuthUser, error) {
	const op = opTag + "GetAllUsers"

	url := fmt.Sprintf("%s/identities", r.config.KratosAdmin)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}
	req = req.WithContext(ctx)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.Wrap(fmt.Errorf("response status was %d and not %d", res.StatusCode, http.StatusOK), op)
	}

	var identities []kratos.Identity
	if err := json.NewDecoder(res.Body).Decode(&identities); err != nil {
		return nil, errors.Wrap(err, op)
	}

	var users []*domain.AuthUser

	for _, identity := range identities {
		traitBytes, err := json.Marshal(identity.Traits)
		if err != nil {
			return nil, errors.Wrap(err, op)
		}

		traits := &domain.Traits{}
		if err := json.Unmarshal(traitBytes, traits); err != nil {
			return nil, errors.Wrap(err, op)
		}

		users = append(users, &domain.AuthUser{
			Traits: traits,
			Session: &kratos.Session{
				Identity: identity,
			},
		})
	}

	return users, nil
}

type recoveryData struct {
	ExpiresIn string `json:"expires_in"`
	UserID    string `json:"identity_id"`
}

type recoveryRes struct {
	Link string `json:"recovery_link"`
}

func (r *authRepo) GetRecoveryLink(ctx context.Context, userID string) (string, error) {
	const op = opTag + "GetRecoveryLink"

	url := fmt.Sprintf("%s/recovery/link", r.config.KratosAdmin)

	data, err := json.Marshal(recoveryData{"24h", userID})
	if err != nil {
		return "", errors.Wrap(err, op)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return "", errors.Wrap(err, op)
	}
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	req = req.WithContext(ctx)

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", errors.Wrap(err, op)
	}

	link := &recoveryRes{}
	if err := json.NewDecoder(res.Body).Decode(link); err != nil {
		return "", errors.Wrap(err, op)
	}

	return link.Link, nil
}
