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
    CREATE TYPE playerNameSearchResult AS (
        PlayerID varchar(80),
        Platform varchar(128),
        LastSeen timestamp,
        Name varchar(128)
    );
EXCEPTION
    WHEN duplicate_object THEN null;
END $$;

create or replace function search_player_names(term varchar, limt int, offst int)
    returns table (
        playerid varchar,
        platform varchar,
        lastseen timestamp,
        playername varchar
    ) language plpgsql as $$
declare looprow playerNameSearchResult;
    declare tmp playerNameSearchResult;
begin
    for looprow in select pn.playerid, pn.platform from playernames pn
                   where lower(name) like concat('%', lower(term), '%')
                   group by pn.playerid, pn.platform
                   limit limt offset offst
        loop
            select p.playerid, p.platform, p.lastseen, pn.name into tmp from playernames pn
            inner join players p on p.playerid = pn.playerid and p.platform = pn.platform
            where pn.playerid = looprow.playerid
              and pn.platform = looprow.platform
            order by daterecorded desc
            limit 1;

            playerid := tmp.playerid;
            platform := tmp.platform;
            lastseen := tmp.lastseen;
            playername := tmp.name;

            return next;
        end loop;
end; $$;