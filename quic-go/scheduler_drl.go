package quic

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"net/http"
)

var ModelID uint64 = 0

// SACmodel's config - start
const (
	serverURL  = "http://127.0.0.1:5000"
	stateDim   = 6
	actionDim  = 3
	maxAction  = 1.0
	bufferSize = 64
)

type SetModelRequest struct {
	StateDim     int     `json:"state_dim"`
	ActionDim    int     `json:"action_dim"`
	MaxAction    float64 `json:"max_action"`
	ConnectionID uint64  `json:"connection_id"`
}

type SelectActionRequest struct {
	State        []float64 `json:"state"`
	ConnectionID uint64    `json:"connection_id"`
}

type TrainModelRequest struct {
	ReplayBuffer string `json:"replay_buffer"`
	Iterations   int    `json:"iterations"`
	ConnectionID uint64 `json:"connection_id"`
}

type SelectActionResponse struct {
	ActionProbs []float64 `json:"action_probs"`
}

type Experience struct {
	State     []float64 `json:"state"`
	Action    int       `json:"action"`
	Reward    float64   `json:"reward"`
	NextState []float64 `json:"next_state"`
	Done      bool      `json:"done"`
}

//SACmodel's config - end

func sendRequest(endpoint string, data interface{}) error {
	url := serverURL + endpoint
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal request data: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("server returned non-OK status: %s, response: %s", resp.Status, body)
	}

	responseData, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("failed to read response: %w", err)
	}

	fmt.Printf("Response from %s: %s\n", endpoint, responseData)
	return nil
}

func selectAction(state []float64, connectionID uint64) ([]float64, error) {
	selectActionRequest := SelectActionRequest{
		State:        state,
		ConnectionID: connectionID,
	}

	url := serverURL + "/select_action"
	jsonData, err := json.Marshal(selectActionRequest)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request data: %w", err)
	}

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return nil, fmt.Errorf("server returned non-OK status: %s, response: %s", resp.Status, body)
	}

	var selectActionResponse SelectActionResponse
	if err := json.NewDecoder(resp.Body).Decode(&selectActionResponse); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return selectActionResponse.ActionProbs, nil
}

func trainModel(replayBuffer []Experience, iterations int, connectionID uint64) error {
	trainModelRequest := TrainModelRequest{
		ReplayBuffer: "replay_buffer_1", // Example identifier for replay buffer
		Iterations:   iterations,
		ConnectionID: connectionID,
	}
	fmt.Println(trainModelRequest)
	return sendRequest("/train_model", trainModelRequest)
}

func plotTrainingHistory(connectionID uint64) error {
	plotRequest := map[string]uint64{"connection_id": connectionID}
	return sendRequest("/plot_training_history", plotRequest)
}

func generateRandomState(dim int) []float64 {
	state := make([]float64, dim)
	for i := range state {
		state[i] = rand.Float64()
	}
	return state
}

func chooseAction(actionProbs []float64) int {
	// Choose action based on the probabilities
	r := rand.Float64()
	cumulativeProb := 0.0
	for i, prob := range actionProbs {
		cumulativeProb += prob
		if r < cumulativeProb {
			return i + 1 // Actions are 1-indexed
		}
	}
	return len(actionProbs) // In case of numerical issues, return the last action
}

func simulateEnvironment(state []float64, action int) (float64, []float64, bool) {
	// Simulate reward and next state based on the current state and action
	reward := rand.Float64() // Example reward
	nextState := generateRandomState(len(state))
	done := rand.Float64() < 0.1 // Example condition to end episode
	return reward, nextState, done
}
