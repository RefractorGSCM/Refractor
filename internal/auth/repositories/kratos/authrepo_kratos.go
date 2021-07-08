package kratos

import (
	"Refractor/domain"
	"Refractor/params"
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

func NewAuthRepo(config *conf.Config) domain.AuthRepository {
	return &authRepo{
		config: config,
	}
}

func (r *authRepo) CreateUser(body *params.CreateUserParams) (*domain.User, error) {
	const op = opTag + "CreateUser"

	url := fmt.Sprintf("%s/identities", r.config.KratosAdmin)

	marshalled, err := json.Marshal(body)
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

	identity := &kratos.Identity{}
	if err := json.NewDecoder(res.Body).Decode(identity); err != nil {
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
	user := &domain.User{
		Traits:   traits,
		Identity: identity,
	}

	return user, nil
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
