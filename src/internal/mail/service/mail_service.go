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
	"bytes"
	"fmt"
	"github.com/go-gomail/gomail"
	"html/template"
	"net/url"
	"strconv"
)

type mailService struct {
	uri    *url.URL
	dialer *gomail.Dialer
}

func NewMailService(config *conf.Config) (domain.MailService, error) {
	uri, err := url.Parse(config.SmtpConnectionUri)
	if err != nil {
		return nil, err
	}

	username := uri.User.Username()
	password, _ := uri.User.Password()

	var port int64
	if uri.Port() != "" {
		var err error
		port, err = strconv.ParseInt(uri.Port(), 10, 32)
		if err != nil {
			return nil, err
		}
	} else {
		port = 587 // default SMTP outgoing port
	}

	dialer := gomail.NewDialer(uri.Hostname(), int(port), username, password)

	return &mailService{
		uri:    uri,
		dialer: dialer,
	}, nil
}

func (s *mailService) SendMail(to []string, sub string, body string) error {
	m := gomail.NewMessage()
	m.SetHeader("From", fmt.Sprintf("refractor@%s", s.uri.Hostname()))
	m.SetHeader("To", to...)
	m.SetHeader("Subject", sub)
	m.SetBody("text/html", body)

	if err := s.dialer.DialAndSend(m); err != nil {
		return err
	}

	return nil
}

type welcomeEmailData struct {
	Inviter string
	Link    string
}

func (s *mailService) SendWelcomeEmail(to, inviterName, link string) error {
	data := welcomeEmailData{
		Inviter: inviterName,
		Link:    link,
	}

	body, err := s.parseTemplate("./internal/mail/templates/welcome.html", data)
	if err != nil {
		return err
	}

	if err := s.SendMail([]string{to}, "Welcome to Refractor", body.String()); err != nil {
		return err
	}

	return nil
}

func (s *mailService) parseTemplate(templateFile string, data interface{}) (*bytes.Buffer, error) {
	t, err := template.ParseFiles(templateFile)
	if err != nil {
		return nil, err
	}

	buf := new(bytes.Buffer)
	if err := t.Execute(buf, data); err != nil {
		return nil, err
	}

	return buf, nil
}
