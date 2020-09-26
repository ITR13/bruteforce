package main

import (
	"github.com/ITR13/bruteforce"
	"github.com/ITR13/bruteforce/databases/ram"
)

const (
	Empty   Tile = 0
	Player1 Tile = 1
	Player2 Tile = 2
	Filled  Tile = 3
)

type Tile uint8

type GameState struct {
	tiles [9]Tile
	count uint8
}

var Lines = [][][3]int{
	[][3]int{
		[3]int{0, 1, 2},
		[3]int{0, 3, 6},
		[3]int{0, 4, 8},
	},
	[][3]int{
		[3]int{0, 1, 2},
		[3]int{1, 4, 7},
	},
	[][3]int{
		[3]int{0, 1, 2},
		[3]int{2, 5, 8},
		[3]int{2, 4, 6},
	},
	[][3]int{
		[3]int{3, 4, 5},
		[3]int{0, 3, 6},
	},
	[][3]int{
		[3]int{3, 4, 5},
		[3]int{1, 4, 7},
		[3]int{0, 4, 8},
		[3]int{2, 4, 6},
	},
	[][3]int{
		[3]int{3, 4, 5},
		[3]int{2, 5, 8},
	},
	[][3]int{
		[3]int{6, 7, 8},
		[3]int{0, 3, 6},
		[3]int{2, 4, 6},
	},
	[][3]int{
		[3]int{6, 7, 8},
		[3]int{1, 4, 7},
	},
	[][3]int{
		[3]int{6, 7, 8},
		[3]int{2, 5, 8},
		[3]int{0, 4, 8},
	},
	// VictoryChecks
	[][3]int{
		[3]int{0, 1, 2},
		[3]int{3, 4, 5},
		[3]int{6, 7, 8},
		[3]int{0, 3, 6},
		[3]int{1, 4, 7},
		[3]int{2, 5, 8},
		[3]int{0, 4, 8},
		[3]int{2, 4, 6},
	},
}

var Rotation = [9]int{2, 5, 8, 1, 4, 7, 0, 3, 6}

func main() {
	stop := false

	gameInfo := bruteforce.GameInfo{
		"tic-tac-toe", "3x3",
		CheckWin,
		CheckRecursiveWin,
		GetNext,
	}
	db := ramdb.NewDatabase(true)
	startState := compress(GameState{})
	db.UpdateSteps(&startState, 0)

	gameInfo.RunSingleThreaded(db, &stop)
}

func CheckWin(compressedState *bruteforce.GameState) bruteforce.Player {
	state := decompress(compressedState)
	if state.count < 5 {
		return bruteforce.NoPlayer
	}

	checks := Lines[9]
	for i := range checks {
		victory := state.tiles[checks[i][0]]
		if victory == Empty || victory == Filled {
			continue
		}
		if state.tiles[checks[i][1]] != victory || state.tiles[checks[i][2]] != victory {
			continue
		}
		if victory == Player1 {
			return bruteforce.Player1
		}
		return bruteforce.Player2
	}

	if state.count == 9 {
		return bruteforce.BothPlayers
	}

	return bruteforce.NoPlayer
}

func CheckRecursiveWin(compressedState *bruteforce.GameState, p1, p2, draw uint8) bruteforce.Player {
	state := decompress(compressedState)
	if state.count%2 == 0 {
		if p1 > 0 {
			return bruteforce.Player1
		} else if draw > 0 {
			return bruteforce.BothPlayers
		}
		return bruteforce.Player2
	}

	if p2 > 0 {
		return bruteforce.Player2
	} else if draw > 0 {
		return bruteforce.BothPlayers
	}
	return bruteforce.Player1
}

func GetNext(compressedState *bruteforce.GameState) []bruteforce.StatePath {
	state := decompress(compressedState)
	statePaths := make([]bruteforce.StatePath, 9-state.count)
	stateIndex := 0

	var player Tile
	if state.count%2 == 0 {
		player = Player1
	} else {
		player = Player2
	}

	for i := 0; i < 9; i++ {
		if state.tiles[i] != Empty {
			continue
		}
		currentNextState := state.place(i, player)
		findDeadTiles(&currentNextState)
		reorder(&currentNextState)

		compressedNextState := compress(currentNextState)
		statePaths[stateIndex] = bruteforce.StatePath{
			&compressedNextState, 1, false,
		}
		stateIndex++
	}
	return statePaths
}

func decompress(state *bruteforce.GameState) GameState {
	tiles := [9]Tile{
		Tile((*state)[0]) & 3,
		Tile((*state)[0]>>2) & 3,
		Tile((*state)[0]>>4) & 3,
		Tile((*state)[0]>>6) & 3,
		Tile((*state)[1]) & 3,
		Tile((*state)[1]>>2) & 3,
		Tile((*state)[1]>>4) & 3,
		Tile((*state)[1]>>6) & 3,
		Tile((*state)[2]) & 3,
	}

	count := uint8(0)
	for i := range tiles {
		if tiles[i] != 0 {
			count++
		}
	}
	decompressed := GameState{
		tiles,
		count,
	}
	return decompressed
}

func compress(state GameState) bruteforce.GameState {
	tiles := state.tiles
	return []byte{
		byte(tiles[0] | tiles[1]<<2 | tiles[2]<<4 | tiles[3]<<6),
		byte(tiles[4] | tiles[5]<<2 | tiles[6]<<4 | tiles[7]<<6),
		byte(tiles[8]),
	}
}

func (gameState GameState) place(index int, tile Tile) GameState {
	newState := GameState{}
	for i := 0; i < 9; i++ {
		newState.tiles[i] = gameState.tiles[i]
	}
	newState.tiles[index] = tile
	newState.count++
	return newState
}

func findDeadTiles(gameState *GameState) {
	tiles := gameState.tiles
	for tileIndex := range tiles {
		if tiles[tileIndex] == Empty || tiles[tileIndex] == Filled {
			continue
		}
		tile := tiles[tileIndex]

		anyValid := false
		for lineIndex := range Lines[tileIndex] {
			line := Lines[tileIndex][lineIndex]
			valid := true
			for i := 0; i < 3; i++ {
				if tiles[line[i]] == tile {
					continue
				}
				if tiles[line[i]] == Empty {
					continue
				}
				valid = false
				break
			}

			if valid {
				anyValid = true
				break
			}
		}

		if !anyValid {
			gameState.tiles[tileIndex] = Filled
		}
	}
}

func reorder(gameState *GameState) {
	next := (*gameState).rotate()

	maybeReplace := func() {
		if !isLess(*gameState, next) {
			*gameState = next
		}
	}

	maybeReplace()
	next = next.rotate()
	maybeReplace()
	next = next.rotate()
	maybeReplace()

	next = next.transpose()
	maybeReplace()
	next = next.rotate()
	maybeReplace()
	next = next.rotate()
	maybeReplace()
	next = next.rotate()
	maybeReplace()
}

func (gameState GameState) rotate() GameState {
	newState := GameState{}
	newState.count = gameState.count
	for i := range newState.tiles {
		newState.tiles[i] = gameState.tiles[Rotation[i]]
	}

	return newState
}

func (gameState GameState) transpose() GameState {
	newState := GameState{}
	newState.count = gameState.count

	for i := range newState.tiles {
		newState.tiles[(i%3)*3+i/3] = gameState.tiles[i]
	}

	return newState
}

func isLess(first, second GameState) bool {
	for i := range first.tiles {
		if first.tiles[i] < second.tiles[i] {
			return true
		}
		if first.tiles[i] > second.tiles[i] {
			return false
		}
	}
	return false
}
