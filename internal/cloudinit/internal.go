package cloudinit

type statusJSON struct {
	V1 statusJSONv1 `json:"v1"`
}

type statusJSONv1 struct {
	Datasource    string            `json:"datasource"`
	Stage         *string           `json:"stage"`
	Init          statusJSONv1Stage `json:"init"`
	InitLocal     statusJSONv1Stage `json:"init-local"`
	ModulesConfig statusJSONv1Stage `json:"modules-config"`
	ModulesFinal  statusJSONv1Stage `json:"modules-final"`
}

type statusJSONv1Stage struct {
	Errors            []string            `json:"errors"`
	Finished          float64             `json:"finished"`
	Start             float64             `json:"start"`
	RecoverableErrors map[string][]string `json:"recoverable_errors"`
}
