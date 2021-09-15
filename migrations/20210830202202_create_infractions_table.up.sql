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

DO $$ BEGIN
    CREATE TYPE InfractionType AS ENUM ('WARNING', 'MUTE', 'KICK', 'BAN');
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

CREATE TABLE IF NOT EXISTS Infractions(
    InfractionID SERIAL NOT NULL PRIMARY KEY,
    PlayerID VARCHAR(80) NOT NULL,
    Platform VARCHAR(128) NOT NULL,
    UserID VARCHAR(36),
    ServerID SERIAL NOT NULL,
    Type InfractionType NOT NULL,
    Reason TEXT NOT NULL DEFAULT '',
    Duration INT,
    SystemAction BOOLEAN DEFAULT FALSE,
    CreatedAt TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    ModifiedAt TIMESTAMP,

    FOREIGN KEY (PlayerID, Platform) REFERENCES Players (PlayerID, Platform),
    FOREIGN KEY (ServerID) References Servers (ServerID)
);

DROP TRIGGER IF EXISTS update_infractions_modat ON Infractions;
CREATE TRIGGER update_infractions_modat BEFORE UPDATE ON Infractions
    FOR EACH ROW EXECUTE PROCEDURE update_modified_at_column();