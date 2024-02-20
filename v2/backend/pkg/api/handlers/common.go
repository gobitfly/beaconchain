package apihandlers

import (
	"net/http"

	dataaccess "github.com/gobitfly/beaconchain/pkg/api/data_access"
)

// All changes to common functions MUST NOT break any public handler behavior

type HandlerService struct {
	dai dataaccess.DataAccessInterface
}

func NewHandlerService(das dataaccess.DataAccessInterface) HandlerService {
	return HandlerService{dai: das}
}

func ReturnOk(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("OK"))
}

func ReturnCreated(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Created"))
}

func ReturnNoContent(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusNoContent)
}
