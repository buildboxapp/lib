package models

// тип ответа, который сервис отдает прокси при периодическом опросе (ping-е)
type Pong struct {
	Name string `json:"name"`
	Version string `json:"version"`
	Status string `json:"status"`
	Port int `json:"port"`
	Pid  string `json:"pid"`
	State string `json:"state"`
	Replicas int `json:"replicas"`
	Https bool `json:"https"`
}

