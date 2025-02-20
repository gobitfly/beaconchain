package types

type SSVExporterResponse struct {
	Type   string `json:"type"`
	Filter struct {
		From int `json:"from"`
		To   int `json:"to"`
	} `json:"filter"`
	Data []SSVExporterData `json:"data"`
}

type SSVExporterData struct {
	Index     int    `json:"index"`
	Publickey string `json:"publicKey"`
	Operators []struct {
		Nodeid    int    `json:"nodeId"`
		Publickey string `json:"publicKey"`
	} `json:"operators"`
}
