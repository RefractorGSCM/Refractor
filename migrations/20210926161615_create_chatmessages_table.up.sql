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

CREATE TABLE IF NOT EXISTS ChatMessages (
    MessageID SERIAL NOT NULL PRIMARY KEY,
    PlayerID VARCHAR(80) NOT NULL,
    Platform VARCHAR(128) NOT NULL,
    ServerID SERIAL NOT NULL,
    Message TEXT NOT NULL,
    Flagged BOOLEAN DEFAULT FALSE,

    FOREIGN KEY (PlayerID, Platform) REFERENCES Players (PlayerID, Platform),
    FOREIGN KEY (ServerID) REFERENCES Servers (ServerID)
);