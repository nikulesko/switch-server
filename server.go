package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"sync"

	"github.com/stianeikeland/go-rpio"
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
	err := rpio.Open()

	if err != nil {
		log.Fatal(err)
	}
	defer rpio.Close()

	// Inint GPIO here for the rasberry
	pin1_0 := rpio.Pin(3) // GPIO2
	pin1_0.Output()
	pin1_1 := rpio.Pin(5) // GPIO3
	pin1_1.Output()

	pin2_0 := rpio.Pin(7) // GPIO4
	pin2_0.Output()
	pin2_1 := rpio.Pin(11) // GPIO17
	pin2_1.Output()

	pin3_0 := rpio.Pin(13) // GPIO27
	pin3_0.Output()
	pin3_1 := rpio.Pin(15) // GPIO22
	pin3_1.Output()

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

		// "Clean" all the switches
		pin1_0.Low()
		pin1_1.Low()
		pin2_0.Low()
		pin2_1.Low()
		pin3_0.Low()
		pin3_1.Low()

		switch req.ID {
		case 0: //RF1
			pin1_0.High()
			pin2_0.High()
			pin3_0.High()
		case 1: //RF2
			pin1_1.High()
			pin2_0.High()
			pin3_0.High()
		case 2: //RF3
			pin1_0.High()
			pin2_1.High()
			pin3_0.High()
		case 3: //RF4
			pin1_1.High()
			pin2_1.High()
			pin3_0.High()
		case 4: //RF5
			pin1_0.High()
			pin2_0.High()
			pin3_1.High()
		case 5: //RF6
			pin1_1.High()
			pin2_0.High()
			pin3_1.High()
		case 6: //RF7
			pin1_0.High()
			pin2_1.High()
			pin3_1.High()
		case 7: //RF8
			pin1_1.High()
			pin2_1.High()
			pin3_1.High()
		}

		log.Println("ID toggled:", req.ID)

		log.Printf("State pin1_0: %v", pin1_0.Read())
		log.Printf("State pin1_1: %v", pin1_1.Read())
		log.Printf("State pin2_0: %v", pin2_0.Read())
		log.Printf("State pin2_1: %v", pin2_1.Read())
		log.Printf("State pin3_0: %v", pin3_0.Read())
		log.Printf("State pin3_1: %v", pin3_1.Read())

		log.Println("-----------------------------------")

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(newState)
	})

	log.Printf("Server starting on port %d", port)
	if err := http.ListenAndServe(fmt.Sprintf(":%d", port), nil); err != nil {
		log.Fatal(err)
	}
}
