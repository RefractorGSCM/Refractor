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

alter table ChatMessages add column if not exists MessageVectors tsvector;
create index if not exists idx_search_message_vectors on ChatMessages using gin(MessageVectors);
update ChatMessages set MessageVectors = to_tsvector(Message);

create or replace function update_chatmessage_vectors()
returns trigger as $$
begin
    new.MessageVectors = to_tsvector(new.Message);
    return new;
end;
$$ language 'plpgsql';

drop trigger if exists update_chatmessages_msgvecs on ChatMessages;
create trigger update_chatmessages_msgvecs before update of Message on ChatMessages
    for each row execute procedure update_chatmessage_vectors();