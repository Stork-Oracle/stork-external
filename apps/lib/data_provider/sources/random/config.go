package random

type randomConfig struct {
	UpdateFrequency string  `json:"updateFrequency"`
	MinValue        float64 `json:"minValue"`
	MaxValue        float64 `json:"maxValue"`
}
