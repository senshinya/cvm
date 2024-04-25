package lexer

import (
	"fmt"
	"github.com/thoas/go-funk"
	"shinya.click/cvm/common"
)

type state string
type stateTable map[state]map[condition]state
type condition string
type conditionFunc func(byte) bool
type conditionTable map[condition]conditionFunc
type tokenConstructor func(string, int) common.Token

type ScannerBuilder struct {
	scanner *Scanner
}

type Scanner struct {
	stateTable       stateTable
	conditionTable   conditionTable
	tokenConstructor tokenConstructor
	startState       state
	endState         map[state]struct{}
}

func NewScannerBuilder() *ScannerBuilder {
	return &ScannerBuilder{
		scanner: &Scanner{},
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

func (b *ScannerBuilder) Build() *Scanner {
	s := b.scanner
	if err := checkScannerValid(s.stateTable, s.conditionTable, s.startState, s.endState); err != nil {
		panic(err)
	}
	return s
}

func (s *Scanner) scan(lexState *Lexer) common.Token {
	cState := s.startState
	for !lexState.isAtEnd() {
		cByte := lexState.peek()
		canTransfer := false
		transferMap := s.stateTable[cState]
		for cond, nState := range transferMap {
			condFunc := s.conditionTable[cond]
			if condFunc(cByte) {
				// bingo, transfer state
				canTransfer = true
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
			return s.tokenConstructor(lexState.source[lexState.start:lexState.current], lexState.line)
		}
		// panic!
		panic(lexState.source[lexState.start:lexState.current])
	}

	// read to the end, see if cState is an end state
	if _, canEnd := s.endState[cState]; canEnd {
		return s.tokenConstructor(lexState.source[lexState.start:lexState.current], lexState.line)
	}
	// panic!
	panic(lexState.source[lexState.start:lexState.current])
}

func checkScannerValid(stateTable stateTable, conditionTable conditionTable, startState state, endState map[state]struct{}) error {
	states := funk.Map(stateTable, func(s state, _ map[condition]state) (state, struct{}) {
		return s, struct{}{}
	}).(map[state]struct{})
	conditions := map[condition]struct{}{}

	// check if all states in stateTable are defined
	for _, conditionTransfer := range stateTable {
		for cond, newState := range conditionTransfer {
			conditions[cond] = struct{}{}
			if _, ok := states[newState]; !ok {
				return fmt.Errorf("unknown state: %s", newState)
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
