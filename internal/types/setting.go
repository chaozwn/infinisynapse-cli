package types

type EngineConfigItem struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type UpdateEngineConfigParams struct {
	Config []EngineConfigItem `json:"config"`
}
