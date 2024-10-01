package handlers

import (
	"net/http"

	"github.com/gobitfly/beaconchain/pkg/api/types"
)

func (h *HandlerService) InternalGetUserMachineMetrics(w http.ResponseWriter, r *http.Request) {
	h.PublicGetUserMachineMetrics(w, r)
}

func (h *HandlerService) PublicGetUserMachineMetrics(w http.ResponseWriter, r *http.Request) {
	var v validationError
	userId, err := GetUserIdByContext(r)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	q := r.URL.Query()
	offset := v.checkUint(q.Get("offset"), "offset")
	limit := uint64(180)
	if limitParam := q.Get("limit"); limitParam != "" {
		limit = v.checkUint(limitParam, "limit")
	}

	data, err := h.dai.GetUserMachineMetrics(r.Context(), userId, limit, offset)
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.GetUserMachineMetricsRespone{
		Data: *data,
	}

	returnOk(w, r, response)
}
