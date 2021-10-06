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

/* The UserMeta is for recording user info as well as other metadata not currently supported by Ory Kratos (the identity
   management solution we are using). Storing this data separately allows us to keep track of contextual info which may
   be useful to Refractor admins. This table also contains the 'Deactivated' property which allows us to block user
   requests if an account needs to be deactivated without having to delete their account.
*/
CREATE TABLE IF NOT EXISTS UserMeta(
    UserID VARCHAR(36) NOT NULL PRIMARY KEY,
    InitialUsername VARCHAR(20) NOT NULL,
    Username VARCHAR(20) NOT NULL,
    Deactivated BOOLEAN DEFAULT FALSE
)