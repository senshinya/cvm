package lexer

import (
	"fmt"
	"github.com/thoas/go-funk"
	"shinya.click/cvm/common"
	"shinya.click/cvm/lexer/util"
)

type state string
type stateTable map[state][]Edge
type condition string
type conditionFunc func(byte) bool
type conditionTable map[condition]conditionFunc
type tokenConstructor func(lexeme string, line, sc, ec int, endState state, literal interface{}) (common.Token, error)
type transferInterceptor func(before, after state, char byte, store interface{}, l, sc, ec int) error
type Edge struct {
	condition condition
	state     state
}

func (s state) in(group []state) bool {
	for _, c := range group {
		if c == s {
			return true
		}
	}
	return false
}

type ScannerBuilder struct {
	scanner *Scanner
}

type Scanner struct {
	name                string // just for error message
	stateTable          stateTable
	conditionTable      conditionTable
	tokenConstructor    tokenConstructor
	startState          state
	endState            map[state]struct{}
	transferInterceptor transferInterceptor
	store               interface{}
}

func NewScannerBuilder(name string) *ScannerBuilder {
	return &ScannerBuilder{
		scanner: &Scanner{name: name},
	}
}

func (b *ScannerBuilder) StateTable(stateTable stateTable) *ScannerBuilder {
	b.scanner.stateTable = stateTable
	return b
}

func (b *ScannerBuilder) ConditionTable(conditionTable conditionTable) *ScannerBuilder {
	b.scanner.conditionTable = conditionTable
	return b
}

func (b *ScannerBuilder) TokenConstructor(tokenConstructor tokenConstructor) *ScannerBuilder {
	b.scanner.tokenConstructor = tokenConstructor
	return b
}

func (b *ScannerBuilder) StartState(startState state) *ScannerBuilder {
	b.scanner.startState = startState
	return b
}

func (b *ScannerBuilder) EndState(endState []state) *ScannerBuilder {
	b.scanner.endState = funk.Map(endState, func(s state) (state, struct{}) {
		return s, struct{}{}
	}).(map[state]struct{})
	return b
}

func (b *ScannerBuilder) transferInterceptor(transferInterceptor transferInterceptor) *ScannerBuilder {
	b.scanner.transferInterceptor = transferInterceptor
	return b
}

func (b *ScannerBuilder) Build() *Scanner {
	s := b.scanner
	if err := checkScannerValid(s.stateTable, s.conditionTable, s.startState, s.endState); err != nil {
		panic(common.NewInitError(common.ErrInvalidStateMachine, "Invalid %s Scanner: %s", b.scanner.name, err.Error()))
	}
	return s
}

var emptyToken = common.Token{}

func (s *Scanner) scan(lexState *Lexer) (common.Token, error) {
	cState := s.startState
	for !lexState.isAtEnd() {
		cByte := lexState.peek()
		canTransfer := false
		edges := s.stateTable[cState]
		for _, edge := range edges {
			condFunc := s.conditionTable[edge.condition]
			if condFunc(cByte) {
				// bingo, transfer state
				canTransfer = true
				nState := edge.state
				if s.transferInterceptor != nil {
					err := s.transferInterceptor(cState, nState, cByte, s.store, lexState.line, lexState.sColumn, lexState.cColumn)
					if err != nil {
						return emptyToken, err
					}
				}
				cState = nState
				break
			}
		}
		if canTransfer {
			lexState.moveNext()
			continue
		}
		// cannot transfer, see if cState is an end state
		if _, canEnd := s.endState[cState]; canEnd {
			return s.tokenConstructor(lexState.source[lexState.start:lexState.current], lexState.line, lexState.sColumn, lexState.cColumn, cState, s.store)
		}
		// unknown token
		return emptyToken, util.NewLexerError(util.ErrUnidentifiableToken, lexState.line, lexState.sColumn, lexState.cColumn, lexState.source[lexState.start:lexState.current])
	}

	// read to the end, see if cState is an end state
	if _, canEnd := s.endState[cState]; canEnd {
		return s.tokenConstructor(lexState.source[lexState.start:lexState.current], lexState.line, lexState.sColumn, lexState.cColumn, cState, s.store)
	}
	// unknown token
	return emptyToken, util.NewLexerError(util.ErrUnidentifiableToken, lexState.line, lexState.sColumn, lexState.cColumn, lexState.source[lexState.start:lexState.current])
}

func (s *Scanner) Store(store interface{}) {
	s.store = store
}

func checkScannerValid(stateTable stateTable, conditionTable conditionTable, startState state, endState map[state]struct{}) error {
	states := funk.Map(stateTable, func(s state, _ []Edge) (state, struct{}) {
		return s, struct{}{}
	}).(map[state]struct{})
	conditions := map[condition]struct{}{}

	// check if all states in stateTable are defined
	for _, edges := range stateTable {
		for _, edge := range edges {
			conditions[edge.condition] = struct{}{}
			if _, ok := states[edge.state]; !ok {
				return fmt.Errorf("unknown state: %s", edge.state)
			}
		}
	}

	// check if all conditions in stateTable are defined
	for cond := range conditions {
		if _, ok := conditionTable[cond]; !ok {
			return fmt.Errorf("unknown condition: %s", cond)
		}
	}

	// check if start state and end states are defined
	for s := range endState {
		if _, ok := states[s]; !ok {
			return fmt.Errorf("unknown end state: %s", s)
		}
	}
	if _, ok := states[startState]; !ok {
		return fmt.Errorf("unknown start state: %s", startState)
	}

	return nil
}
