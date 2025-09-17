package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"
)

const inputCount = 8
const port = 8080

type ButtonState struct {
	mu     *sync.Mutex
	states []bool
}

func NewButtonState(count int, mu *sync.Mutex) *ButtonState {
	return &ButtonState{
		states: make([]bool, count),
		mu:     mu,
	}
}

func (bs *ButtonState) GetStates() []bool {
	bs.mu.Lock()
	defer bs.mu.Unlock()
	result := make([]bool, len(bs.states))
	copy(result, bs.states)
	return result
}

// Toggle there are no situations where more than one button can be ON at the same time and all buttons cannot be OFF
func (bs *ButtonState) Toggle(id int) ([]bool, error) {
	bs.mu.Lock()
	defer bs.mu.Unlock()

	if id < 0 || id >= len(bs.states) {
		return nil, fmt.Errorf("invalid button ID")
	}

	for i := 0; i < len(bs.states); i++ {
		if i == id && !bs.states[i] {
			continue
		}
		bs.states[i] = false
	}

	bs.states[id] = !bs.states[id]

	result := make([]bool, len(bs.states))
	copy(result, bs.states)
	return result, nil
}

func main() {
	tmpl, err := template.ParseFiles("templates/index.html")
	if err != nil {
		log.Fatal(err)
	}

	var mutex sync.Mutex

	buttonState := NewButtonState(inputCount, &mutex)
	//init button states by the raspberry pi GPIO here

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		tmpl.Execute(w, nil)
	})

	http.HandleFunc("/api/state", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(buttonState.GetStates())
	})

	http.HandleFunc("/api/toggle", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}

		var req struct {
			ID int `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "Invalid request", http.StatusBadRequest)
			return
		}

		newState, err := buttonState.Toggle(req.ID)

		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(newState)
	})

	log.Printf("Server starting on port %d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal(err)
	}
}
