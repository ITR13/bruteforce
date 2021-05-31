package main

import (
	"fmt"

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
	victory bruteforce.Player
	tiles   *[7][6]Tile
	count   uint8
	empty   uint8
}

func main() {
	stop := false

	gameInfo := bruteforce.GameInfo{
		"connect-4", "7x6",
		CheckWin,
		CheckRecursiveWin,
		GetNext,
	}
	db := ramdb.NewDatabase(false)
	startState := compress(GameState{
		bruteforce.NoPlayer,
		&[7][6]Tile{},
		0,
		7,
	})
	db.UpdateSteps(&startState, 0)

	gameInfo.RunSingleThreaded(db, &stop)
}

func CheckWin(compressedState *bruteforce.GameState) bruteforce.Player {
	return decompressWinner(compressedState)
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
	statePaths := make([]bruteforce.StatePath, state.empty)
	stateIndex := 0

	var player Tile
	if state.count%2 == 0 {
		player = Player1
	} else {
		player = Player2
	}

	for i := 0; i < 7; i++ {
		if state.tiles[i][5] != Empty {
			continue
		}

		for j := 0; j < 6; j++ {
			if state.tiles[i][j] != Empty {
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
			break
		}
	}

	if stateIndex != int(state.empty) {
		panic(fmt.Errorf("State has %d empty rows, but only %d pieces were placed", state.empty, stateIndex))
	}

	return statePaths
}

func decompress(compressed *bruteforce.GameState) GameState {
	tiles := [7][6]Tile{}

	for i := 0; i <= 3; i++ {
		tiles[i*2] = [6]Tile{
			Tile((*compressed)[i*3]) & 3,
			Tile((*compressed)[i*3]>>2) & 3,
			Tile((*compressed)[i*3]>>4) & 3,
			Tile((*compressed)[i*3]>>6) & 3,
			Tile((*compressed)[i*3+1]) & 3,
			Tile((*compressed)[i*3+1]>>2) & 3,
		}

		if i == 3 {
			break
		}

		tiles[i*2+1] = [6]Tile{
			Tile((*compressed)[i*3+1]>>4) & 3,
			Tile((*compressed)[i*3+1]>>6) & 3,
			Tile((*compressed)[i*3+2]) & 3,
			Tile((*compressed)[i*3+2]>>2) & 3,
			Tile((*compressed)[i*3+2]>>4) & 3,
			Tile((*compressed)[i*3+2]>>6) & 3,
		}
	}

	victory := decompressWinner(compressed)

	count, empty := uint8(0), uint8(0)
	for i := range tiles {
		for j := range tiles[i] {
			if tiles[i][j] != Empty {
				count++
			}
		}
		if tiles[i][5] == Empty {
			empty++
		}
	}

	decompressed := GameState{
		victory,
		&tiles,
		count,
		empty,
	}
	return decompressed
}

func decompressWinner(state *bruteforce.GameState) bruteforce.Player {
	return bruteforce.Player((*state)[10]>>4) & 3
}

func compress(decompressed GameState) bruteforce.GameState {
	compressed := make([]byte, 11)

	for i := 0; i <= 3; i++ {
		compressed[i*3] |= byte(decompressed.tiles[i*2][0])
		compressed[i*3] |= byte(decompressed.tiles[i*2][1] << 2)
		compressed[i*3] |= byte(decompressed.tiles[i*2][2] << 4)
		compressed[i*3] |= byte(decompressed.tiles[i*2][3] << 6)
		compressed[i*3+1] |= byte(decompressed.tiles[i*2][4])
		compressed[i*3+1] |= byte(decompressed.tiles[i*2][5] << 2)

		if i == 3 {
			break
		}

		compressed[i*3+1] |= byte(decompressed.tiles[i*2+1][0] << 4)
		compressed[i*3+1] |= byte(decompressed.tiles[i*2+1][1] << 6)
		compressed[i*3+2] |= byte(decompressed.tiles[i*2+1][2])
		compressed[i*3+2] |= byte(decompressed.tiles[i*2+1][3] << 2)
		compressed[i*3+2] |= byte(decompressed.tiles[i*2+1][4] << 4)
		compressed[i*3+2] |= byte(decompressed.tiles[i*2+1][5] << 6)
	}
	compressed[10] |= byte(decompressed.victory << 4)

	return compressed
}

func (gameState GameState) place(column int, tile Tile) GameState {
	tiles := *gameState.tiles

	verticalWin := 0
	var row int
	for row = range tiles[column] {
		if tiles[column][row] == Empty {
			tiles[column][row] = tile
			verticalWin++
			break
		} else if tiles[column][row] == tile {
			verticalWin++
		} else {
			verticalWin = 0
		}
	}
	empty := gameState.empty

	if row == 7 {
		panic("Failed to place tile")
	} else if row == 6 {
		empty--
	}

	horizontalWin := lineWin(
		&tiles,
		column,
		tile,
		func(_ int) int { return row },
	)
	diagonal1Win := lineWin(
		&tiles,
		column,
		tile,
		func(x int) int { return row + column - x },
	)
	diagonal2Win := lineWin(
		&tiles,
		column,
		tile,
		func(x int) int { return row + x - column },
	)

	victory := bruteforce.NoPlayer
	if verticalWin > 3 || horizontalWin || diagonal1Win || diagonal2Win {
		victory = bruteforce.Player(tile)
	}

	return GameState{
		victory,
		&tiles,
		gameState.count + 1,
		empty,
	}
}

func lineWin(
	tiles *[7][6]Tile,
	current int,
	tile Tile,
	transform func(int) int,
) bool {
	combo := 0
	for x := 0; x < 7; x++ {
		y := transform(x)
		if y < 0 || y >= 6 {
			continue
		}

		if tiles[x][y] == tile {
			combo++
		} else if x < current {
			combo = 0
		} else {
			break
		}
	}
	return combo > 3
}

// func findDeadTiles(gameState *GameState) // Check validity.go

func reorder(gameState *GameState) {
	next := (*gameState).mirror()

	maybeReplace := func() {
		ok := !isLess(*gameState, next)
		if ok {
			*gameState = next
		}
	}

	maybeReplace()
}

func (gameState GameState) mirror() GameState {
	newState := GameState{
		gameState.victory,
		&[7][6]Tile{},
		gameState.count,
		gameState.empty,
	}

	for x := 0; x < 7; x++ {
		for y := 0; y < 6; y++ {
			newState.tiles[x][y] = gameState.tiles[6-x][y]
		}
	}

	return newState
}

func (gameState GameState) colorSwap() GameState {
	if gameState.count%2 != 0 {
		panic("Swapping colors is only valid if it doesn't change the starting player")
	}

	newState := GameState{
		gameState.victory,
		&[7][6]Tile{},
		gameState.count,
		gameState.empty,
	}

	for x := 0; x < 7; x++ {
		for y := 0; y < 6; y++ {
			switch gameState.tiles[x][y] {
			case Player1:
				newState.tiles[x][y] = Player2
				break
			case Player2:
				newState.tiles[x][y] = Player1
			default:
				newState.tiles[x][y] = gameState.tiles[x][y]
				break
			}
		}
	}

	return newState
}

func isLess(first, second GameState) bool {
	for x := 0; x < 7; x++ {
		for y := 0; y < 6; y++ {
			if first.tiles[x][y] < second.tiles[x][y] {
				return true
			}
			if first.tiles[x][y] > second.tiles[x][y] {
				return false
			}
		}
	}
	return false
}

func isEqual(first, second GameState) bool {
	for x := 0; x < 7; x++ {
		for y := 0; y < 6; y++ {
			if first.tiles[x][y] != second.tiles[x][y] {
				return false
			}
		}
	}
	return true
}
