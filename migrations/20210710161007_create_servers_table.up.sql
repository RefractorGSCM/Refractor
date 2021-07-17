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

CREATE TABLE IF NOT EXISTS Servers(
    ServerID SERIAL NOT NULL PRIMARY KEY,
    Game VARCHAR(32) NOT NULL,
    Name VARCHAR(20) NOT NULL,
    Address VARCHAR(15) NOT NULL,
    RCONPort VARCHAR(5) NOT NULL,
    RCONPassword VARCHAR(128) NOT NULL,
    CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ModifiedAt TIMESTAMP
);

CREATE TRIGGER update_servers_modat BEFORE UPDATE ON Servers
    FOR EACH ROW EXECUTE PROCEDURE update_modified_at_column();