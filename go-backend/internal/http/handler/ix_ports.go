package handler

import (
	"net/http"
	"strconv"
	"strings"

	"go-backend/internal/http/response"
	"go-backend/internal/store/model"
)

func requireIXPortAdmin(w http.ResponseWriter, r *http.Request) bool {
	if isAdminRequest(r) {
		return true
	}
	response.WriteJSON(w, response.Err(403, "仅管理员可以管理 IX 端口资源"))
	return false
}

func (h *Handler) ixPortPoolList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || !requireIXPortAdmin(w, r) {
		return
	}
	pools, err := h.repo.ListIXPortPools()
	if err != nil {
		response.WriteJSON(w, response.Err(-2, err.Error()))
		return
	}
	response.WriteJSON(w, response.OK(pools))
}

func (h *Handler) ixPortPoolCreate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !requireIXPortAdmin(w, r) {
		return
	}
	var pool model.IXPortPool
	if err := decodeJSON(r.Body, &pool); err != nil {
		response.WriteJSON(w, response.Err(500, "请求参数错误"))
		return
	}
	if strings.EqualFold(strings.TrimSpace(pool.Kind), "full") {
		pool.RangeStart, pool.RangeEnd = 1, 65535
	}
	if err := h.repo.CreateIXPortPool(&pool); err != nil {
		response.WriteJSON(w, response.Err(-2, err.Error()))
		return
	}
	response.WriteJSON(w, response.OK(pool))
}

func (h *Handler) ixPortPoolUpdate(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !requireIXPortAdmin(w, r) {
		return
	}
	var pool model.IXPortPool
	if err := decodeJSON(r.Body, &pool); err != nil {
		response.WriteJSON(w, response.Err(500, "请求参数错误"))
		return
	}
	if err := h.repo.UpdateIXPortPool(&pool); err != nil {
		response.WriteJSON(w, response.Err(-2, err.Error()))
		return
	}
	response.WriteJSON(w, response.OK(nil))
}

func (h *Handler) ixPortResourceList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || !requireIXPortAdmin(w, r) {
		return
	}
	poolID, err := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("poolId")), 10, 64)
	if err != nil || poolID <= 0 {
		response.WriteJSON(w, response.Err(500, "端口池不能为空"))
		return
	}
	resources, err := h.repo.ListIXPortResources(poolID)
	if err != nil {
		response.WriteJSON(w, response.Err(-2, err.Error()))
		return
	}
	response.WriteJSON(w, response.OK(resources))
}

func (h *Handler) ixPortResourceUpsert(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost || !requireIXPortAdmin(w, r) {
		return
	}
	var resource model.IXPortResource
	if err := decodeJSON(r.Body, &resource); err != nil {
		response.WriteJSON(w, response.Err(500, "请求参数错误"))
		return
	}
	if err := h.repo.SaveIXPortResource(&resource); err != nil {
		response.WriteJSON(w, response.Err(-2, err.Error()))
		return
	}
	response.WriteJSON(w, response.OK(resource))
}

func (h *Handler) ixPortLeaseList(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet || !requireIXPortAdmin(w, r) {
		return
	}
	ownerType := strings.TrimSpace(r.URL.Query().Get("ownerType"))
	ownerID, _ := strconv.ParseInt(strings.TrimSpace(r.URL.Query().Get("ownerId")), 10, 64)
	leases, err := h.repo.ListIXPortLeases(ownerType, ownerID)
	if err != nil {
		response.WriteJSON(w, response.Err(-2, err.Error()))
		return
	}
	response.WriteJSON(w, response.OK(leases))
}
