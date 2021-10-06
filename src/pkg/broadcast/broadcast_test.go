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

package broadcast

import (
	"reflect"
	"regexp"
	"testing"
)

var (
	mordhauJoinPattern = regexp.MustCompile("^Login: (?P<date>[0-9\\.-]+): (?P<name>.+) \\((?P<playfabid>[0-9a-fA-F]+)\\) logged in$")
	mordhauQuitPattern = regexp.MustCompile("^Login: (?P<date>[0-9\\.-]+): (?P<name>.+) \\((?P<playfabid>[0-9a-fA-F]+)\\) logged out$")
)

func TestGetBroadcastType(t *testing.T) {
	type args struct {
		broadcast string
		patterns  map[string]*regexp.Regexp
	}
	tests := []struct {
		name string
		args args
		want *Broadcast
	}{
		{
			name: "broadcast.gettype.mordhau.1",
			args: args{
				broadcast: "Login: 2021.01.01-00.00.00: Test (52DAB212C79F5EC) logged in",
				patterns: map[string]*regexp.Regexp{
					TypeJoin: mordhauJoinPattern,
					TypeQuit: mordhauQuitPattern,
				},
			},
			want: &Broadcast{
				Type: "JOIN",
				Fields: map[string]string{
					"date":      "2021.01.01-00.00.00",
					"name":      "Test",
					"playfabid": "52DAB212C79F5EC",
				},
			},
		},
		{
			name: "broadcast.gettype.mordhau.2",
			args: args{
				broadcast: "Login: 3000.01.05-00.30.30: Us#rWith* WeIrDN@mE (537AB82CF9F82A5) logged out",
				patterns: map[string]*regexp.Regexp{
					TypeJoin: mordhauJoinPattern,
					TypeQuit: mordhauQuitPattern,
				},
			},
			want: &Broadcast{
				Type: "QUIT",
				Fields: map[string]string{
					"date":      "3000.01.05-00.30.30",
					"name":      "Us#rWith* WeIrDN@mE",
					"playfabid": "537AB82CF9F82A5",
				},
			},
		},
		{
			name: "broadcast.gettype.mordhau.3",
			args: args{
				broadcast: "invalid broadcast. no match!",
				patterns: map[string]*regexp.Regexp{
					TypeJoin: mordhauJoinPattern,
					TypeQuit: mordhauQuitPattern,
				},
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := GetBroadcastType(tt.args.broadcast, tt.args.patterns); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetBroadcastType() = %v, want %v", got, tt.want)
			}
		})
	}
}
