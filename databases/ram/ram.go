package ramdb

import (
	"fmt"

	"github.com/ITR13/bruteforce"
)

type Database struct {
	logAll       bool
	metaData     bruteforce.MetaData
	stateById    map[string]*bruteforce.StateData
	statesByStep map[uint8][]*bruteforce.StateData
}

func NewDatabase(logAll bool) bruteforce.Database {
	db := Database{
		logAll,
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
			panic(
				fmt.Errorf(
					"State '%x' reachable in both %d and %d steps",
					stateStr, stateData.MinimumSteps, minSteps,
				),
			)
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

	if db.logAll {
		fmt.Printf("State '%x' has steps %d\n", stateStr, minSteps)
	}
}
func (db *Database) SetStateSearched(state *bruteforce.GameState) {
	stateStr := string(*state)
	stateData, ok := db.stateById[stateStr]
	if !ok {
		panic(fmt.Errorf("State '%x' doesn't exist", stateStr))
		return
	}
	if stateData.SearchState != bruteforce.Unsearched {
		panic(fmt.Errorf("State '%x' already searched", stateStr))
		return
	}
	stateData.SearchState = bruteforce.Searched
	if db.logAll {
		fmt.Printf("State '%x' was searched\n", stateStr)
	}
}
func (db *Database) SetStateEnd(state *bruteforce.GameState, winningPlayer bruteforce.Player) {
	stateStr := string(*state)
	stateData, ok := db.stateById[stateStr]
	if !ok {
		panic(fmt.Errorf("State '%x' doesn't exist", stateStr))
		return
	}
	if stateData.SearchState != bruteforce.Unsearched {
		panic(fmt.Errorf("State '%x' not unsearched", stateStr))
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

	if db.logAll {
		fmt.Printf("State '%x' ends with winner %v\n", stateStr, winningPlayer)
	}
}

func (db *Database) SetStateSolved(state *bruteforce.GameState, p1Wins, p2Wins, draw uint8, winningPlayer bruteforce.Player) {
	stateStr := string(*state)
	stateData, ok := db.stateById[stateStr]
	if !ok {
		panic(fmt.Errorf("State '%x' doesn't exist", stateStr))
		return
	}
	if stateData.SearchState != bruteforce.Searched {
		panic(fmt.Errorf("State '%x' not searched", stateStr))
		return
	}
	stateData.P1Wins = p1Wins
	stateData.P2Wins = p2Wins
	stateData.Draw = draw
	stateData.Winner = winningPlayer
	stateData.SearchState = bruteforce.End

	if db.logAll {
		fmt.Printf("State '%x' was solved with winner %v (%d, %d, %d)\n",
			stateStr,
			winningPlayer,
			p1Wins, p2Wins, draw,
		)
	}
}

func (db *Database) UpdateCurrentStep(step uint8, exiting bool) {
	db.metaData.CurrentStep = step
	db.metaData.Exiting = exiting
	if db.logAll {
		fmt.Printf("Updated to step %d exiting %v\n", step, exiting)
	}
}

func (db *Database) GetStepsAndWinner(state *bruteforce.GameState) (uint8, bruteforce.Player) {
	stateStr := string(*state)
	stateData, ok := db.stateById[stateStr]
	if !ok {
		panic(fmt.Errorf("State '%x' doesn't exist", stateStr))
	}
	return stateData.MinimumSteps, stateData.Winner
}
func (db *Database) SetError(state *bruteforce.GameState, err bruteforce.Error) {
	stateStr := string(*state)
	panic(fmt.Errorf("State '%x' has error %v", stateStr, err))
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

	if db.logAll {
		fmt.Printf("Getting %d states with step %d and searchState %v\n",
			foundCount,
			step,
			searchState,
		)
	}

	return foundStates[:foundCount]
}
