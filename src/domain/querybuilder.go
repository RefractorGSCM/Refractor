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

package domain

type QueryBuilder interface {
	BuildExistsQuery(table string, args map[string]interface{}) (string, []interface{})
	BuildFindQuery(table string, args map[string]interface{}) (string, []interface{})
	BuildUpdateQuery(table string, id interface{}, idName string, args map[string]interface{}, returnFields []string) (string, []interface{})
	BuildUpdateQueryComposite(table string, ids []interface{}, idNames []string, args map[string]interface{}) (string, []interface{})
}
