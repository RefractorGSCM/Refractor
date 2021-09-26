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

package rules

import (
	"Refractor/domain"
	"Refractor/params/validators"
	validation "github.com/go-ozzo/ozzo-validation"
	"github.com/go-ozzo/ozzo-validation/is"
	"math"
	"regexp"
)

var PlatformRules = RuleGroup{
	validation.Length(1, 128),
	validation.By(validators.ValueInStrArray(domain.AllPlatforms)),
}

var PlayerIDRules = RuleGroup{
	validation.Length(1, 80),
}

var InfractionReasonRules = RuleGroup{
	validation.Length(1, 1024),
}

var InfractionDurationRules = RuleGroup{
	validation.Min(0),
	validation.Max(math.MaxInt32),
}

var attachmentUrlPattern = regexp.MustCompile(".(jpeg|jpg|gif|png)$")
var AttachmentURLRules = RuleGroup{
	is.RequestURL,
	validation.Match(attachmentUrlPattern),
}

var AttachmentNoteRules = RuleGroup{
	validation.Length(1, 512),
}

var SearchOffsetRules = RuleGroup{
	validation.Min(0),
	validation.Max(math.MaxInt32),
}

var SearchLimitRules = RuleGroup{
	validation.Min(0),
	validation.Max(100),
}

var UserIDRules = RuleGroup{
	is.UUIDv4,
}
