package statemachine

import (
	"fmt"

	"baton/internal/storage"
)

// ValidTransitions defines the allowed state transitions
var ValidTransitions = map[storage.State][]storage.State{
	storage.ReadyForPlan: {
		storage.Planning,
	},
	storage.Planning: {
		storage.ReadyForImplementation,
		storage.NeedsFixes,
	},
	storage.ReadyForImplementation: {
		storage.Implementing,
	},
	storage.Implementing: {
		storage.ReadyForCodeReview,
		storage.NeedsFixes,
	},
	storage.ReadyForCodeReview: {
		storage.Reviewing,
	},
	storage.Reviewing: {
		storage.ReadyForCommit,
		storage.NeedsFixes,
	},
	storage.ReadyForCommit: {
		storage.Committing,
	},
	storage.Committing: {
		storage.Done,
		storage.NeedsFixes,
	},
	storage.NeedsFixes: {
		storage.Fixing,
	},
	storage.Fixing: {
		storage.ReadyForCodeReview,
		storage.NeedsFixes,
	},
	storage.Done: {
		// Terminal state - no transitions
	},
}

// ValidateTransition validates if a state transition is allowed
func ValidateTransition(from, to storage.State) error {
	// Normalize states (handle aliases)
	from = storage.NormalizeState(string(from))
	to = storage.NormalizeState(string(to))

	// Check if transition is valid
	allowedStates, exists := ValidTransitions[from]
	if !exists {
		return fmt.Errorf("invalid source state: %s", from)
	}

	for _, allowedState := range allowedStates {
		if to == allowedState {
			return nil
		}
	}

	return fmt.Errorf("invalid transition from %s to %s. Allowed transitions: %v", from, to, allowedStates)
}

// GetAllowedTransitions returns the list of allowed transitions from a given state
func GetAllowedTransitions(from storage.State) ([]storage.State, error) {
	from = storage.NormalizeState(string(from))

	allowedStates, exists := ValidTransitions[from]
	if !exists {
		return nil, fmt.Errorf("invalid state: %s", from)
	}

	return allowedStates, nil
}

// IsTerminalState checks if a state is terminal (no outgoing transitions)
func IsTerminalState(state storage.State) bool {
	state = storage.NormalizeState(string(state))
	allowedStates, exists := ValidTransitions[state]
	return exists && len(allowedStates) == 0
}

// GetAllStates returns all valid states
func GetAllStates() []storage.State {
	states := make([]storage.State, 0, len(ValidTransitions))
	for state := range ValidTransitions {
		states = append(states, state)
	}
	return states
}