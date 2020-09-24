package bruteforce

const (
	NoPlayer    Player = 0
	Player1     Player = 1
	Player2     Player = 2
	BothPlayers Player = 3
)

type Player uint8
type GameState []byte

type GameInfo struct {
	game, variant string
	checkWin      func(*GameState) Player
	getNext       func(*GameState) []*GameState
}

func (gameInfo *GameInfo) search(
	database Database,
	gameState *GameState,
	nextStep uint8,
) {
	win := gameInfo.checkWin(gameState)
	if win != NoPlayer {
		database.SetStateEnd(gameState, win)
		return
	}

	nextStates := gameInfo.getNext(gameState)
	for i := range nextStates {
		database.UpdateSteps(nextStates[i], nextStep)
	}

	database.SetStateSearched(gameState)
}

func (gameInfo *GameInfo) solve(
	database Database,
	gameState *GameState,
	currentStep uint8,
) {
	nextStates := gameInfo.getNext(gameState)
	p1, p2, draw := uint8(0), uint8(0), uint8(0)
	err := NoError

	for i := range nextStates {
		steps, win := database.GetStepsAndWinner(nextStates[i])
		if steps <= currentStep {
			err = NonLinear
			continue
		}

		switch win {
		case NoPlayer:
			err = NoPlayerChildren
			break
		case Player1:
			p1++
		case Player2:
			p2++
		case BothPlayers:
			draw++
		}
	}
	if p1 == 0 && p2 == 0 && draw == 0 {
		err = NoEndstate
		draw = 1
	}

	database.SetStateSolved(gameState, p1, p2, draw)
	if err != NoError {
		database.SetError(gameState, err)
	}
}
