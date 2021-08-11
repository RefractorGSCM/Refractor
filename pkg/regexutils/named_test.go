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

package regexutils

import (
	"github.com/stretchr/testify/assert"
	"regexp"
	"testing"
)

func TestMapNamedMatches(t *testing.T) {
	type args struct {
		pattern *regexp.Regexp
		data    string
	}
	tests := []struct {
		name string
		args args
		want map[string]string
	}{
		{
			name: "regexutils.namedmatches.1",
			args: args{
				pattern: regexp.MustCompile("^(?P<ID>[0-9]+)$"),
				data:    "1",
			},
			want: map[string]string{
				"ID": "1",
			},
		},
		{
			name: "regexutils.namedmatches.2",
			args: args{
				pattern: regexp.MustCompile("^(?P<ID>[0-9]+),(?P<Username>[0-9a-zA-Z]+)$"),
				data:    "1,test",
			},
			want: map[string]string{
				"ID":       "1",
				"Username": "test",
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fields := MapNamedMatches(tt.args.pattern, tt.args.data)

			assert.Equal(t, tt.want, fields)
		})
	}
}
