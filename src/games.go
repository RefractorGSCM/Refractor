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

package main

import (
	"Refractor/domain"
	"Refractor/games/minecraft"
	"Refractor/games/mordhau"
	"Refractor/platforms/mojang"
	"Refractor/platforms/playfab"
)

func registerGames(gs domain.GameService) {
	// Create platform instances
	_playfab := playfab.NewPlayfabPlatform()
	_mojang := mojang.NewMojangPlatform()

	gs.AddGame(mordhau.NewMordhauGame(_playfab))
	gs.AddGame(minecraft.NewMinecraftGame(_mojang))
	// ADD NEW GAME PACKAGES HERE
}
