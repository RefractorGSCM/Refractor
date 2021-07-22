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
	"Refractor/pkg/conf"
	"fmt"
	"net/url"
)

type mailService struct {
}

func NewMailService(config *conf.Config) (domain.MailService, error) {
	uri, err := url.Parse(config.SmtpConnectionUri)
	if err != nil {
		return nil, err
	}

	fmt.Println(uri)

	return &mailService{}, nil
}

func (s *mailService) SendMail(to []string, msg string) error {
	return nil
}
