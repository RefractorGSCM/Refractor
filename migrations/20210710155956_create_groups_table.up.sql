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

CREATE TABLE IF NOT EXISTS Groups(
    GroupID SERIAL PRIMARY KEY,
    Name VARCHAR(20) NOT NULL,
    Color INT NOT NULL DEFAULT CAST(x'e0e0e0' AS INT),
    Position INT NOT NULL,
    Permissions VARCHAR(20) NOT NULL,
    CreatedAt TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    ModifiedAt TIMESTAMP
);

CREATE TRIGGER update_groups_modat BEFORE UPDATE ON Groups
    FOR EACH ROW EXECUTE PROCEDURE update_modified_at_column();