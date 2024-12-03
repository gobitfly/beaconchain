package database

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

const (
	routeBulkAdd         = "/bulkAdd"
	routeGetRowsRange    = "/rowRange"
	routeGetRow          = "/row"
	routeRead            = "/read"
	routeGetRowsWithKeys = "/rowsWithKeys"
)

type RemoteServer struct {
	db Database
}

func NewRemote(db Database) RemoteServer {
	return RemoteServer{db: db}
}

func (api RemoteServer) Routes() http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc(routeBulkAdd, api.BulkAdd)
	mux.HandleFunc(routeGetRowsRange, api.GetRowsRange)
	mux.HandleFunc(routeGetRow, api.GetRow)
	mux.HandleFunc(routeRead, api.Read)
	mux.HandleFunc(routeGetRowsWithKeys, api.GetRowsWithKeys)

	return mux
}

type ParamsBulkAdd struct {
	Items map[string][]Item `json:"items"`
}

func (api RemoteServer) BulkAdd(w http.ResponseWriter, r *http.Request) {
	var args ParamsBulkAdd
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		respondWithErr(w, http.StatusBadRequest, err)
		return
	}
	err := api.db.BulkAdd(args.Items)
	if err != nil {
		respondWithErr(w, http.StatusInternalServerError, err)
		return
	}
	respond(w, nil)
}

type ParamsGetRowsRange struct {
	High      string `json:"high"`
	Low       string `json:"low"`
	Limit     int64  `json:"limit"`
	OpenRange bool   `json:"open_range"`
}

func (api RemoteServer) GetRowsRange(w http.ResponseWriter, r *http.Request) {
	var args ParamsGetRowsRange
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		respondWithErr(w, http.StatusBadRequest, err)
		return
	}
	rows, err := api.db.GetRowsRange(args.High, args.Low, WithOpenRange(args.OpenRange), WithLimit(args.Limit))
	if err != nil {
		respondWithErr(w, http.StatusInternalServerError, err)
		return
	}
	respond(w, rows)
}

type ParamsGetRow struct {
	Key string `json:"key"`
}

func (api RemoteServer) GetRow(w http.ResponseWriter, r *http.Request) {
	var args ParamsGetRow
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		respondWithErr(w, http.StatusBadRequest, err)
		return
	}
	row, err := api.db.GetRow(args.Key)
	if err != nil {
		respondWithErr(w, http.StatusInternalServerError, err)
		return
	}
	respond(w, row)
}

type ParamsRead struct {
	Prefix string `json:"prefix"`
}

func (api RemoteServer) Read(w http.ResponseWriter, r *http.Request) {
	var args ParamsRead
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		respondWithErr(w, http.StatusBadRequest, err)
		return
	}
	rows, err := api.db.Read(args.Prefix)
	if err != nil {
		respondWithErr(w, http.StatusInternalServerError, err)
		return
	}
	respond(w, rows)
}

type ParamsGetRowsWithKeys struct {
	Keys []string `json:"keys"`
}

func (api RemoteServer) GetRowsWithKeys(w http.ResponseWriter, r *http.Request) {
	var args ParamsGetRowsWithKeys
	if err := json.NewDecoder(r.Body).Decode(&args); err != nil {
		respondWithErr(w, http.StatusBadRequest, err)
		return
	}
	rows, err := api.db.GetRowsWithKeys(args.Keys)
	if err != nil {
		respondWithErr(w, http.StatusInternalServerError, err)
		return
	}
	respond(w, rows)
}

func respond(w http.ResponseWriter, data any) {
	b, err := json.Marshal(data)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte(err.Error()))
		return
	}
	_, _ = w.Write(b)
}

func respondWithErr(w http.ResponseWriter, statusCode int, err error) {
	w.WriteHeader(statusCode)
	_, _ = w.Write([]byte(err.Error()))
}

type RemoteClient struct {
	url string
}

func NewRemoteClient(url string) *RemoteClient {
	return &RemoteClient{url: url}
}

func (r RemoteClient) Add(key string, item Item, allowDuplicate bool) error {
	//TODO implement me
	panic("implement me")
}

func (r RemoteClient) BulkAdd(itemsByKey map[string][]Item, opts ...Option) error {
	b, err := json.Marshal(ParamsBulkAdd{Items: itemsByKey})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", r.url, routeBulkAdd), bytes.NewReader(b))
	if err != nil {
		return err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		if ErrNotFound.Error() == string(b) {
			return ErrNotFound
		}
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, b)
	}
	return nil
}

func (r RemoteClient) Read(prefix string) ([]Row, error) {
	b, err := json.Marshal(ParamsRead{Prefix: prefix})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", r.url, routeRead), bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, b)
	}
	var rows []Row
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r RemoteClient) GetRow(key string) (*Row, error) {
	b, err := json.Marshal(ParamsGetRow{Key: key})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", r.url, routeGetRow), bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		if ErrNotFound.Error() == string(b) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, b)
	}
	var row Row
	if err := json.NewDecoder(resp.Body).Decode(&row); err != nil {
		return nil, err
	}
	return &row, nil
}

func (r RemoteClient) GetRowKeys(prefix string, opts ...Option) ([]string, error) {
	//TODO implement me
	panic("implement me")
}

func (r RemoteClient) GetLatestValue(key string) (*Row, error) {
	//TODO implement me
	panic("implement me")
}

func (r RemoteClient) GetRowsRange(high, low string, opts ...Option) ([]Row, error) {
	options := apply(opts)
	b, err := json.Marshal(ParamsGetRowsRange{
		High:      high,
		Low:       low,
		Limit:     options.Limit,
		OpenRange: options.OpenRange,
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", r.url, routeGetRowsRange), bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		if ErrNotFound.Error() == string(b) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, b)
	}
	var rows []Row
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r RemoteClient) GetRowsWithKeys(keys []string) ([]Row, error) {
	b, err := json.Marshal(ParamsGetRowsWithKeys{
		Keys: keys,
	})
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequest(http.MethodPost, fmt.Sprintf("%s%s", r.url, routeGetRowsWithKeys), bytes.NewReader(b))
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, b)
	}
	var rows []Row
	if err := json.NewDecoder(resp.Body).Decode(&rows); err != nil {
		return nil, err
	}
	return rows, nil
}

func (r RemoteClient) Close() error {
	//TODO implement me
	panic("implement me")
}

func (r RemoteClient) Clear() error {
	//TODO implement me
	panic("implement me")
}
