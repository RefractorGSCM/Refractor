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

package service

import (
	"Refractor/domain"
	"context"
	"time"
)

type authService struct {
	repo        domain.AuthRepo
	mailService domain.MailService
	timeout     time.Duration
}

func NewAuthService(repo domain.AuthRepo, mailService domain.MailService, to time.Duration) domain.AuthService {
	return &authService{
		repo:        repo,
		mailService: mailService,
		timeout:     to,
	}
}

func (s *authService) CreateUser(c context.Context, userTraits *domain.Traits, inviter string) (*domain.AuthUser, error) {
	ctx, cancel := context.WithTimeout(c, s.timeout)
	defer cancel()

	// Create the user
	user, err := s.repo.CreateUser(ctx, userTraits)
	if err != nil {
		return nil, err
	}

	// Generate a new recovery link for the user so they can set their password
	recoveryLink, err := s.repo.GetRecoveryLink(ctx, user.Identity.Id)
	if err != nil {
		return nil, err
	}

	// Send a welcome email containing the recovery link
	if err := s.mailService.SendWelcomeEmail(user.Traits.Email, inviter, recoveryLink); err != nil {
		return nil, err
	}

	return user, nil
}
