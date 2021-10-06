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
 * You should have received A copy of the GNU General Public License
 * along with this program.  If not, see <https://www.gnu.org/licenses/>.
 */

package env

import (
	"fmt"
	"os"
)

type Env struct {
	missingVars []string
}

func RequireEnv(varName string) *Env {
	_, exists := os.LookupEnv(varName)

	env := &Env{
		missingVars: []string{},
	}

	if !exists {
		env.missingVars = append(env.missingVars, varName)
	}

	return env
}

func (e *Env) RequireEnv(varName string) *Env {
	_, exists := os.LookupEnv(varName)

	if !exists {
		e.missingVars = append(e.missingVars, varName)
	}

	return e
}

func (e *Env) GetError() error {
	if len(e.missingVars) > 0 {
		builtError := "The following required environment variables are missing:\n"

		for _, name := range e.missingVars {
			builtError += "  - " + name + "\n"
		}

		builtError += "Please set them then restart the application."

		return fmt.Errorf(builtError)
	}

	return nil
}
