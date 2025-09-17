package main

import (
	"reflect"
	"sync"
	"testing"
)

func TestButtonState_Toggle(t *testing.T) {
	tests := []struct {
		name         string
		initialState []bool
		toggleID     int
		wantState    []bool
		wantErr      bool
	}{
		{
			name:         "toggle off button when another is on",
			initialState: []bool{true, false, false},
			toggleID:     1,
			wantState:    []bool{false, true, false},
			wantErr:      false,
		},
		{
			name:         "toggle on button when it's off",
			initialState: []bool{true, false, false},
			toggleID:     2,
			wantState:    []bool{false, false, true},
			wantErr:      false,
		},
		{
			name:         "toggle off the only on button should toggle it back on",
			initialState: []bool{true, false, false},
			toggleID:     0,
			wantState:    []bool{true, false, false},
			wantErr:      false,
		},
		{
			name:         "invalid button ID negative",
			initialState: []bool{true, false, false},
			toggleID:     -1,
			wantState:    nil,
			wantErr:      true,
		},
		{
			name:         "invalid button ID too large",
			initialState: []bool{true, false, false},
			toggleID:     3,
			wantState:    nil,
			wantErr:      true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mu := &sync.Mutex{}
			bs := &ButtonState{
				states: make([]bool, len(tt.initialState)),
				mu:     mu,
			}
			copy(bs.states, tt.initialState)

			got, err := bs.Toggle(tt.toggleID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Toggle() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.wantState) {
				t.Errorf("Toggle() got = %v, want %v", got, tt.wantState)
			}
		})
	}
}
