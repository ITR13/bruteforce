package bruteforce

const (
	NoPlayer    Player = 0
	Player1     Player = 1
	Player2     Player = 2
	BothPlayers Player = 3
)

type Player uint8
type GameState []byte

type StatePath struct {
	GameState  *GameState
	StepDelta  uint8
	PlayerSwap bool
}

type GameInfo struct {
	Game, Variant     string
	CheckWin          func(state *GameState) Player
	CheckRecursiveWin func(state *GameState, p1, p2, draw uint8) Player
	GetNext           func(state *GameState) []StatePath
}

func (gameInfo *GameInfo) RunSingleThreaded(database Database, stop *bool) {
	metaData := database.GetMetaData()

	for !*stop {
		searchState := Unsearched
		if metaData.Exiting {
			searchState = Searched
		}

		states := database.GetAllWithStepAndSearchState(
			metaData.CurrentStep,
			searchState,
		)
		if len(states) == 0 {
			if metaData.Exiting {
				if metaData.CurrentStep <= 0 {
					return
				}

				metaData.CurrentStep -= 1
				database.UpdateCurrentStep(metaData.CurrentStep, true)
				continue
			}
			metaData.Exiting = true
			database.UpdateCurrentStep(metaData.CurrentStep, true)
			continue
		}

		for i := range states {
			if *stop {
				return
			}

			if !metaData.Exiting {
				gameInfo.search(database, states[i], metaData.CurrentStep)
			} else {
				gameInfo.solve(database, states[i], metaData.CurrentStep)
			}
		}

		if !metaData.Exiting {
			metaData.CurrentStep += 1
		} else {
			if metaData.CurrentStep == 0 {
				return
			}
			metaData.CurrentStep -= 1
		}

		database.UpdateCurrentStep(metaData.CurrentStep, metaData.Exiting)
	}
}

func (gameInfo *GameInfo) search(
	database Database,
	gameState *GameState,
	currentStep uint8,
) {
	win := gameInfo.CheckWin(gameState)
	if win != NoPlayer {
		database.SetStateEnd(gameState, win)
		return
	}

	statePaths := gameInfo.GetNext(gameState)
	for i := range statePaths {
		database.UpdateSteps(
			statePaths[i].GameState,
			currentStep+statePaths[i].StepDelta,
		)
	}

	database.SetStateSearched(gameState)
}

func (gameInfo *GameInfo) solve(
	database Database,
	gameState *GameState,
	currentStep uint8,
) {
	statePaths := gameInfo.GetNext(gameState)
	p1, p2, draw := uint8(0), uint8(0), uint8(0)
	err := NoError

	for i := range statePaths {
		steps, win := database.GetStepsAndWinner(statePaths[i].GameState)
		if steps <= currentStep {
			err = NonLinear
			continue
		}

		switch win {
		case NoPlayer:
			err = NoPlayerChildren
			break
		case Player1:
			if !statePaths[i].PlayerSwap {
				p1++
			} else {
				p2++
			}
		case Player2:
			if !statePaths[i].PlayerSwap {
				p2++
			} else {
				p1++
			}
		case BothPlayers:
			draw++
		}
	}
	if p1 == 0 && p2 == 0 && draw == 0 {
		err = NoEndstate
		draw = 1
	}
	winner := gameInfo.CheckRecursiveWin(gameState, p1, p2, draw)

	database.SetStateSolved(gameState, p1, p2, draw, winner)
	if err != NoError {
		database.SetError(gameState, err)
	}
}
