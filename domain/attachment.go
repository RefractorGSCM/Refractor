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

package domain

import "context"

type Attachment struct {
	AttachmentID int64  `json:"id"`
	InfractionID int64  `json:"infraction_id"`
	URL          string `json:"url"`
	Note         string `json:"note"`
}

type AttachmentRepo interface {
	Store(ctx context.Context, attachment *Attachment) error
	GetByInfraction(ctx context.Context, infractionID int64) ([]*Attachment, error)
	Delete(ctx context.Context, id int64) error
}

type AttachmentService interface {
	Store(c context.Context, attachment *Attachment) error
	GetByInfraction(c context.Context, infractionID int64) ([]*Attachment, error)
	Delete(c context.Context, id int64) error
}
