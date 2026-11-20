package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"go-backend/internal/auth"
	"go-backend/internal/http/middleware"
	"go-backend/internal/store/model"
	"go-backend/internal/store/repo"
)

func newIXPortHandlerTest(t *testing.T) (*Handler, *repo.Repository) {
	t.Helper()
	r, err := repo.Open(filepath.Join(t.TempDir(), "ix-port-handler.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = r.Close() })
	return New(r, "test-secret"), r
}

func ixPortRequest(t *testing.T, method, path string, body interface{}, roleID int) *http.Request {
	t.Helper()
	var raw []byte
	if body != nil {
		var err error
		raw, err = json.Marshal(body)
		if err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(raw))
	claims := auth.Claims{Sub: "1", RoleID: roleID}
	return req.WithContext(context.WithValue(req.Context(), middleware.ClaimsContextKey, claims))
}

func ixPortResponseCode(t *testing.T, recorder *httptest.ResponseRecorder) int {
	t.Helper()
	var payload struct {
		Code int `json:"code"`
	}
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v, body=%s", err, recorder.Body.String())
	}
	return payload.Code
}

func TestIXPortAPIRequiresAdministrator(t *testing.T) {
	h, _ := newIXPortHandlerTest(t)
	recorder := httptest.NewRecorder()
	h.ixPortPoolList(recorder, ixPortRequest(t, http.MethodGet, "/api/v1/ix-port/pool/list", nil, 1))
	if got := ixPortResponseCode(t, recorder); got != 403 {
		t.Fatalf("non-admin response code = %d, want 403", got)
	}
}

func TestIXPortAPIFullPoolAlwaysUsesCompleteRange(t *testing.T) {
	h, r := newIXPortHandlerTest(t)
	if err := r.DB().Create(&model.Node{ID: 7, Name: "gz2", Secret: "secret", ServerIP: "127.0.0.1", Port: "1-65535", CreatedTime: 1, Status: 1, TCPListenAddr: "[::]", UDPListenAddr: "[::]"}).Error; err != nil {
		t.Fatal(err)
	}
	recorder := httptest.NewRecorder()
	h.ixPortPoolCreate(recorder, ixPortRequest(t, http.MethodPost, "/api/v1/ix-port/pool/create", map[string]interface{}{
		"nodeId": 7, "name": "gz2", "kind": "full", "rangeStart": 38000, "rangeEnd": 38099, "status": 1,
	}, 0))
	if got := ixPortResponseCode(t, recorder); got != 0 {
		t.Fatalf("create response code = %d, body=%s", got, recorder.Body.String())
	}
	pools, err := r.ListIXPortPools()
	if err != nil {
		t.Fatal(err)
	}
	if len(pools) != 1 || pools[0].RangeStart != 1 || pools[0].RangeEnd != 65535 {
		t.Fatalf("full pool was narrowed by request: %+v", pools)
	}
}
