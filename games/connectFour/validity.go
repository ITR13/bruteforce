package main

type ValidState uint8

const (
	ValidHorizontal ValidState = 1 << iota
	ValidVertical   ValidState = 1 << iota
	ValidDiagonal1  ValidState = 1 << iota
	ValidDiagonal2  ValidState = 1 << iota
)

func findDeadTiles(gameState *GameState) {
	validity := [7][6]ValidState{}

	for row := 0; row < 6; row++ {
		checkValidLine(gameState.tiles, &validity, ValidHorizontal, func(_ int) int { return row }, false)

		checkValidLine(gameState.tiles, &validity, ValidDiagonal1, func(x int) int { return 3 + row - x }, true)
		checkValidLine(gameState.tiles, &validity, ValidDiagonal2, func(x int) int { return 2 + x - row }, false)
	}

	for column := 0; column < 7; column++ {
		checkValidVertical(gameState.tiles, &validity, column)

		for y := 0; y < 6; y++ {
			if gameState.tiles[column][y] == Empty {
				break
			}

			if validity[column][y] == 0 {
				gameState.tiles[column][y] = Filled
			}
		}
	}
}

func checkValidLine(tiles *[7][6]Tile, validity *[7][6]ValidState, stateToSet ValidState, transform func(int) int, declining bool) {
	currentTile := Filled
	empty := 0
	currentCount := 0

	for x := 0; x < 8; x++ {
		y := transform(x)

		if (!declining && y < 0) || (declining && y >= 6) {
			continue
		}
		if x == 7 || (declining && y < 0) || (!declining && y >= 6) {
			if currentTile != Filled && currentCount+empty >= 4 {
				for x2 := x - currentCount - empty; x2 < x; x2++ {
					validity[x2][transform(x2)] |= stateToSet
				}
			}
			break
		}

		if tiles[x][y] == Empty {
			empty++
			continue
		}
		if tiles[x][y] == currentTile && currentTile != Filled {
			currentCount += empty + 1
			empty = 0
			continue
		}

		if currentTile != Filled && currentCount+empty >= 4 {
			for x2 := x - currentCount - empty; x2 < x; x2++ {
				validity[x2][transform(x2)] |= stateToSet
			}
		}

		currentCount = empty + 1
		currentTile = tiles[x][y]
		empty = 0
	}
}

func checkValidVertical(tiles *[7][6]Tile, validity *[7][6]ValidState, x int) {
	currentTile := Filled
	currentCount := 0

	if tiles[x][5] != Empty {
		return
	}

	for y := 0; y < 7; y++ {
		if y == 6 || tiles[x][y] == Empty {

			if currentCount+(6-y) < 4 || currentTile == Filled {
				return
			}

			for y2 := y - currentCount; y2 < 6; y2++ {
				validity[x][y2] |= ValidVertical
			}

			break
		}

		if currentTile != tiles[x][y] {
			currentTile = tiles[x][y]
			currentCount = 1
		} else {
			currentCount++
		}

	}
}
