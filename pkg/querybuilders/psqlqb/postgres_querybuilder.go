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

package psqlqb

import (
	"Refractor/domain"
	"fmt"
)

type qb struct{}

func NewPostgresQueryBuilder() domain.QueryBuilder {
	return &qb{}
}

func (qb *qb) BuildExistsQuery(table string, args map[string]interface{}) (string, []interface{}) {
	var query string = fmt.Sprintf("SELECT EXISTS(SELECT 1 FROM %s WHERE ", table)
	var values []interface{}

	// Build query
	var i = 1
	for key, val := range args {
		query += fmt.Sprintf("%s = $%d AND ", key, i)
		values = append(values, val)
		i++
	}

	// Cut off trailing AND
	query = query[:len(query)-5] + ");"

	return query, values
}

func (qb *qb) BuildFindQuery(table string, args map[string]interface{}) (string, []interface{}) {
	var query string = fmt.Sprintf("SELECT * FROM %s WHERE ", table)
	var values []interface{}

	// Build query
	var i = 1
	for key, val := range args {
		query += fmt.Sprintf("%s = $%d AND ", key, i)
		values = append(values, val)
		i++
	}

	// Cut off trailing AND
	query = query[:len(query)-5] + ";"

	return query, values
}

func (qb *qb) BuildUpdateQuery(table string, id int64, idName string, args map[string]interface{}) (string, []interface{}) {
	var query = fmt.Sprintf("UPDATE %s SET ", table)
	var values []interface{}

	// Build query
	var i = 1
	for key, val := range args {
		query += fmt.Sprintf("%s = $%d, ", key, i)
		values = append(values, val)
		i++
	}

	values = append(values, id)

	// Cut off trailing comma and space and add where and returning clauses
	query = query[:len(query)-2] + fmt.Sprintf(" WHERE %s = $%d RETURNING *;", idName, i)

	return query, values
}
