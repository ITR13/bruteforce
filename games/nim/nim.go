package main

import (
	"github.com/ITR13/bruteforce"
	"github.com/ITR13/bruteforce/databases/ram"
)

func main() {
	stop := false

	gameInfo := bruteforce.GameInfo{
		"nim", "21",
		CheckWin,
		CheckRecursiveWin,
		GetNext,
	}
	db := ramdb.NewDatabase(true)
	startState := bruteforce.GameState{21}
	db.UpdateSteps(&startState, 0)

	gameInfo.RunSingleThreaded(db, &stop)
}

func CheckWin(state *bruteforce.GameState) bruteforce.Player {
	value := int8((*state)[0])
	if value > 0 {
		return bruteforce.NoPlayer
	}
	return bruteforce.Player1
}

func CheckRecursiveWin(_ *bruteforce.GameState, p1, p2, draw uint8) bruteforce.Player {
	if p1 > 0 {
		return bruteforce.Player1
	} else if draw > 0 {
		return bruteforce.BothPlayers
	}
	return bruteforce.Player2
}

func GetNext(state *bruteforce.GameState) []bruteforce.StatePath {
	value := int8((*state)[0])

	statePaths := make([]bruteforce.StatePath, 3)
	stateIndex := 0

	for i := int8(1); i <= 3; i++ {
		if value < i {
			continue
		}
		nextState := bruteforce.GameState{byte(value - i)}

		statePaths[stateIndex] = bruteforce.StatePath{
			&nextState, uint8(i), true,
		}
		stateIndex++
	}
	return statePaths[:stateIndex]
}
