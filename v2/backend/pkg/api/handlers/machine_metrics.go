package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gobitfly/beaconchain/pkg/api/types"

	"github.com/gobitfly/beaconchain/pkg/commons/db"
	commontypes "github.com/gobitfly/beaconchain/pkg/commons/types"
	"github.com/mitchellh/mapstructure"
	"github.com/pkg/errors"

	"google.golang.org/protobuf/proto"
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

	userInfo, err := h.dai.GetUserInfo(r.Context(), userId)
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

	// validate limit and offset according to user's premium perks
	maxDataPoints := userInfo.PremiumPerks.MachineMonitoringHistorySeconds / 60 // one entry per minute
	timeframe := offset + limit
	if timeframe > maxDataPoints {
		limit = maxDataPoints
		offset = 0
	}

	data, err := h.dai.GetUserMachineMetrics(r.Context(), userId, int(limit), int(offset))
	if err != nil {
		handleErr(w, r, err)
		return
	}
	response := types.GetUserMachineMetricsRespone{
		Data: *data,
	}

	returnOk(w, r, response)
}

func (h *HandlerService) LegacyPostUserMachineMetrics(w http.ResponseWriter, r *http.Request) {
	q := r.URL.Query()
	apiKey := q.Get("apikey")
	machine := q.Get("machine")

	if apiKey == "" {
		apiKey = r.Header.Get("apikey")
	}

	if !h.isPostMachineMetricsEnabled {
		returnError(w, r, http.StatusServiceUnavailable, fmt.Errorf("machine metrics pushing is temporarily disabled"))
		return
	}

	userID, err := h.dai.GetUserIdByApiKey(r.Context(), apiKey)
	if err != nil {
		returnBadRequest(w, r, fmt.Errorf("no user found with api key"))
		return
	}

	userInfo, err := h.dai.GetUserInfo(r.Context(), userID)
	if err != nil {
		handleErr(w, r, err)
		return
	}

	if contentType := r.Header.Get("Content-Type"); !reJsonContentType.MatchString(contentType) {
		returnBadRequest(w, r, fmt.Errorf("invalid content type, expected application/json"))
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil {
		returnBadRequest(w, r, fmt.Errorf("could not read request body"))
		return
	}

	var jsonObjects []map[string]interface{}
	err = json.Unmarshal(body, &jsonObjects)
	if err != nil {
		var jsonObject map[string]interface{}
		err = json.Unmarshal(body, &jsonObject)
		if err != nil {
			returnBadRequest(w, r, errors.Wrap(err, "Invalid JSON format in request body"))
			return
		}
		jsonObjects = []map[string]interface{}{jsonObject}
	}

	if len(jsonObjects) >= 10 {
		returnBadRequest(w, r, fmt.Errorf("Max number of stat entries are 10"))
		return
	}

	var rateLimitErrs = 0
	var result bool = false
	for i := 0; i < len(jsonObjects); i++ {
		err := h.internal_processMachine(r.Context(), machine, &jsonObjects[i], userInfo)
		result = err == nil
		if err != nil {
			if strings.HasPrefix(err.Error(), "rate limit") {
				result = true
				rateLimitErrs++
				continue
			}
			break
		}
	}

	if rateLimitErrs >= len(jsonObjects) {
		returnTooManyRequests(w, r, fmt.Errorf("too many metric requests, max allowed is 1 per user per machine per process"))
		return
	}

	if !result {
		returnError(w, r, http.StatusInternalServerError, fmt.Errorf("could not insert stats"))
		return
	}

	returnOk(w, r, nil)
}

func (h *HandlerService) internal_processMachine(context context.Context, machine string, obj *map[string]interface{}, userInfo *types.UserInfo) error {
	var parsedMeta *commontypes.StatsMeta
	err := mapstructure.Decode(obj, &parsedMeta)
	if err != nil {
		return errors.Wrap(err, "could not parse meta")
	}

	parsedMeta.Machine = machine

	if parsedMeta.Version > 2 || parsedMeta.Version <= 0 {
		return newBadRequestErr("unsupported data format version")
	}

	if parsedMeta.Process != "validator" && parsedMeta.Process != "beaconnode" && parsedMeta.Process != "slasher" && parsedMeta.Process != "system" {
		return newBadRequestErr("unknown process")
	}

	maxNodes := userInfo.PremiumPerks.MonitorMachines

	count, err := db.BigtableClient.GetMachineMetricsMachineCount(commontypes.UserId(userInfo.Id))
	if err != nil {
		return errors.Wrap(err, "could not get machine count")
	}

	if count > maxNodes {
		return newForbiddenErr("user has reached max machine count")
	}

	// protobuf encode
	var data []byte
	if parsedMeta.Process == "system" {
		var parsedResponse *commontypes.MachineMetricSystem
		err = DecodeMapStructure(obj, &parsedResponse)
		if err != nil {
			return errors.Wrap(err, "could not parse stats (system stats)")
		}
		data, err = proto.Marshal(parsedResponse)
		if err != nil {
			return errors.Wrap(err, "could not parse stats (system stats)")
		}
	} else if parsedMeta.Process == "validator" {
		var parsedResponse *commontypes.MachineMetricValidator
		err = DecodeMapStructure(obj, &parsedResponse)
		if err != nil {
			return errors.Wrap(err, "could not parse stats (validator stats)")
		}
		data, err = proto.Marshal(parsedResponse)
		if err != nil {
			return errors.Wrap(err, "could not parse stats (validator stats)")
		}
	} else if parsedMeta.Process == "beaconnode" {
		var parsedResponse *commontypes.MachineMetricNode
		err = DecodeMapStructure(obj, &parsedResponse)
		if err != nil {
			return errors.Wrap(err, "could not parse stats (beaconnode stats)")
		}
		data, err = proto.Marshal(parsedResponse)
		if err != nil {
			return errors.Wrap(err, "could not parse stats (beaconnode stats)")
		}
	}

	return h.dai.PostUserMachineMetrics(context, userInfo.Id, machine, parsedMeta.Process, data)
}

func DecodeMapStructure(input interface{}, output interface{}) error {
	config := &mapstructure.DecoderConfig{
		Metadata: nil,
		Result:   output,
		TagName:  "json",
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	return decoder.Decode(input)
}
