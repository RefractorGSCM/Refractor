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
	"encoding/json"
	"fmt"
	kratos "github.com/ory/kratos-client-go"
	"github.com/pkg/errors"
	"net/http"
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

func (r *authRepo) CreateUser(userTraits *domain.Traits) (*domain.AuthUser, error) {
	const op = opTag + "CreateUser"

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

func (r *authRepo) GetUserByID(id string) (*domain.AuthUser, error) {
	const op = opTag + "GetUserByID"

	url := fmt.Sprintf("%s/identities/%s", r.config.KratosAdmin, id)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.Wrap(fmt.Errorf("response status was %d and not %d", res.StatusCode, http.StatusOK), op)
	}

	session := &kratos.Session{}
	if err := json.NewDecoder(res.Body).Decode(session); err != nil {
		return nil, errors.Wrap(err, op)
	}

	traitBytes, err := json.Marshal(session.Identity.Traits)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	traits := &domain.Traits{}
	if err := json.Unmarshal(traitBytes, traits); err != nil {
		return nil, errors.Wrap(err, op)
	}

	user := &domain.AuthUser{
		Traits:  traits,
		Session: session,
	}

	return user, nil
}

func (r *authRepo) GetAllUsers() ([]*domain.AuthUser, error) {
	const op = opTag + "GetAllUsers"

	url := fmt.Sprintf("%s/identities", r.config.KratosAdmin)

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, errors.Wrap(err, op)
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.Wrap(fmt.Errorf("response status was %d and not %d", res.StatusCode, http.StatusOK), op)
	}

	var sessions []kratos.Session
	if err := json.NewDecoder(res.Body).Decode(&sessions); err != nil {
		return nil, errors.Wrap(err, op)
	}

	var users []*domain.AuthUser

	for _, session := range sessions {
		traitBytes, err := json.Marshal(session.Identity.Traits)
		if err != nil {
			return nil, errors.Wrap(err, op)
		}

		traits := &domain.Traits{}
		if err := json.Unmarshal(traitBytes, traits); err != nil {
			return nil, errors.Wrap(err, op)
		}

		users = append(users, &domain.AuthUser{
			Traits:  traits,
			Session: &session,
		})
	}

	return users, nil
}
