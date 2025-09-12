package app

import (
	"encoding/json"
	"net/http"

	"github.com/gobitfly/beaconchain-backend/api/external/model"
)

// (GET /ping)
func (ApiService) GetPing(w http.ResponseWriter, r *http.Request) {
	resp := model.Pong{
		Ping: "pong",
	}

	w.WriteHeader(http.StatusOK)
	err := json.NewEncoder(w).Encode(resp)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
