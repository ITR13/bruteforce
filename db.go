package bruteforce

const (
	Unsearched SearchState = 0
	Searched   SearchState = 1
	Solved     SearchState = 2
	End        SearchState = 3
)
const (
	NoError          Error = 0
	NonLinear        Error = 1 << 0 //The step logic won't work
	NoPlayerChildren Error = 1 << 1 //At least one child doesn't have an ending
	NoEndstate       Error = 1 << 2 //State has no ending
)

type SearchState uint8
type Error uint8

type Database interface {
	//If minSteps is equal to the stored value: Everything OK
	//If minSteps is higher than the stored value: Game is not one-way
	//If minSteps is lower than the stored value: Something is wrong
	UpdateSteps(state *GameState, minSteps uint8)
	//Only valid if SearchState was Unsearched or everything is the same
	SetStateSearched(state *GameState)
	//Should set winner based on value too
	//Only valid if SearchState was Unsearched and winningPlayer isn't NoPlayer
	// or everything is the same
	SetStateEnd(state *GameState, winningPlayer Player)
	//Should set winner based on values too
	//Only valid if SearchState was Unsearched or everything is the same
	SetStateSolved(state *GameState, p1Wins, p2Wins, draw uint8)
	//Only valid in any of the following cases:
	//  step > currentStep and exiting is false
	//  step = currentStep and exiting goes from false to true
	//  step < currentStep and exiting is true
	UpdateCurrentStep(step uint8, exiting bool)
	//Should return minimum of a gamestate
	GetStepsAndWinner(state *GameState) (uint8, Player)
	//Used to find out where something is amiss
	SetError(state *GameState, err Error)
	//Used to find out what tasks to do next
	GetMetaData() MetaData
	//Returns all gamestates of a step.
	// Will be replaced later to not have to load all states
	GetAllWithStepAndSearchState(step uint8, searchState SearchState) []*GameState
}

type MetaData struct {
	CurrentStep uint8
	Exiting     bool
}

type StateData struct {
	Id                   GameState   // Must be set when first creating State
	MinimumSteps         uint8       // Must be set when first creating Step
	P1Wins, P2Wins, Draw uint8       // All defaults to 0
	Winner               Player      // Defaults to NoPlayer
	SearchState          SearchState //Defaults to Unsearched
}
