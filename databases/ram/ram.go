package ramdb

import (
	"github.com/ITR13/bruteforce"
)

type Database struct {
	metaData     bruteforce.MetaData
	stateById    map[string]*bruteforce.StateData
	statesByStep map[uint8][]*bruteforce.StateData
}

func NewDatabase() bruteforce.Database {
	db := Database{
		bruteforce.MetaData{},
		make(map[string]*bruteforce.StateData),
		make(map[uint8][]*bruteforce.StateData),
	}
	return &db
}

func (db *Database) UpdateSteps(state *bruteforce.GameState, minSteps uint8) {
	stateStr := string(*state)
	if stateData, ok := db.stateById[stateStr]; ok {
		if stateData.MinimumSteps != minSteps {
			panic("State reachable in different amount of steps")
		}
		return
	} else {
		db.stateById[stateStr] = &bruteforce.StateData{
			*state,
			minSteps,
			0, 0, 0,
			bruteforce.NoPlayer,
			bruteforce.Unsearched,
		}
	}
	stateList, ok := db.statesByStep[minSteps]
	if ok {
		stateList = append(stateList, db.stateById[stateStr])
	} else {
		stateList = []*bruteforce.StateData{db.stateById[stateStr]}
	}
	db.statesByStep[minSteps] = stateList
}
func (db *Database) SetStateSearched(state *bruteforce.GameState) {
	stateStr := string(*state)
	stateData, ok := db.stateById[stateStr]
	if !ok {
		panic("State doesn't exist")
		return
	}
	if stateData.SearchState != bruteforce.Unsearched {
		panic("State already searched")
		return
	}
	stateData.SearchState = bruteforce.Searched
}
func (db *Database) SetStateEnd(state *bruteforce.GameState, winningPlayer bruteforce.Player) {
	stateStr := string(*state)
	stateData, ok := db.stateById[stateStr]
	if !ok {
		panic("State doesn't exist")
		return
	}
	if stateData.SearchState != bruteforce.Searched {
		panic("State not searched")
		return
	}
	switch winningPlayer {
	case bruteforce.Player1:
		stateData.P1Wins = 1
		break
	case bruteforce.Player2:
		stateData.P2Wins = 1
		break
	case bruteforce.BothPlayers:
		stateData.Draw = 1
		break
	default:
		return
	}
	stateData.Winner = winningPlayer
	stateData.SearchState = bruteforce.End
}

func (db *Database) SetStateSolved(state *bruteforce.GameState, p1Wins, p2Wins, draw uint8, winningPlayer bruteforce.Player) {
	stateStr := string(*state)
	stateData, ok := db.stateById[stateStr]
	if !ok {
		panic("State doesn't exist")
		return
	}
	if stateData.SearchState != bruteforce.Searched {
		panic("State not searched")
		return
	}
	stateData.P1Wins = p1Wins
	stateData.P2Wins = p2Wins
	stateData.Draw = draw
	stateData.Winner = winningPlayer
	stateData.SearchState = bruteforce.End
}

func (db *Database) UpdateCurrentStep(step uint8, exiting bool) {
	db.metaData.CurrentStep = step
	db.metaData.Exiting = exiting
}

func (db *Database) GetStepsAndWinner(state *bruteforce.GameState) (uint8, bruteforce.Player) {
	stateStr := string(*state)
	stateData, ok := db.stateById[stateStr]
	if !ok {
		panic("State doesn't exist")
	}
	return stateData.MinimumSteps, stateData.Winner
}
func (db *Database) SetError(state *bruteforce.GameState, err bruteforce.Error) {
	panic("Got error")
}
func (db *Database) GetMetaData() bruteforce.MetaData {
	return db.metaData
}
func (db *Database) GetAllWithStepAndSearchState(step uint8, searchState bruteforce.SearchState) []*bruteforce.GameState {
	toSearch := db.statesByStep[step]
	foundStates := make([]*bruteforce.GameState, len(toSearch))
	foundCount := 0
	for i := range toSearch {
		if toSearch[i].SearchState != searchState {
			continue
		}
		foundStates[foundCount] = &toSearch[i].Id
		foundCount++
	}
	return foundStates[:foundCount]
}
