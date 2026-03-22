package handler

import (
	"log"
	"net/http"
	"sensor-service/config"
	"sensor-service/internal/module/user/dto"
	userUsecase "sensor-service/internal/module/user/usecase"
	"sensor-service/internal/platform/helper"
	"sensor-service/internal/platform/httpengine/httpresponse"
	"strings"
)

type UserHandler struct {
	UserUseCase userUsecase.IUserUseCase
}

// GetUserByEmail handles GET /api/user/:email
// O(1) handler delegation to usecase
func (h *UserHandler) GetUserByEmail(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimPrefix(r.URL.Path, "/api/user/")
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	user, err := h.UserUseCase.GetUserByEmail(email)
	if err != nil {
		log.Println("{GetUserByEmail}{Error} : ", err)
		http.Error(w, err.Message, err.Code)
		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status_code": http.StatusOK,
		"message":     "User retrieved successfully",
		"data":        user,
	})
}

// UpdateUser handles PUT /api/user/:email
// O(1) handler delegation to usecase
func (h *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimPrefix(r.URL.Path, "/api/user/")
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Extract caller role from JWT token
	callerRole := helper.GetRoleFromToken(r)
	if callerRole == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var request config.HTTPRequest
	request.Body = make(map[string]interface{})

	if err := helper.DecodeJSON(r.Body, &request.Body); err != nil {
		log.Println("{UpdateUser}{Decode}{Error} : ", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := h.UserUseCase.UpdateUser(email, request, callerRole)
	if err != nil {
		log.Println("{UpdateUser}{Error} : ", err)
		http.Error(w, err.Message, err.Code)
		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status_code": http.StatusOK,
		"message":     "User updated successfully",
		"data":        user,
	})
}

// UpdateUserStatus handles PATCH /api/user/:email/status
// O(1) handler delegation to usecase
func (h *UserHandler) UpdateUserStatus(w http.ResponseWriter, r *http.Request) {
	email := strings.TrimPrefix(r.URL.Path, "/api/user/")
	email = strings.TrimSuffix(email, "/status")
	if email == "" {
		http.Error(w, "Email is required", http.StatusBadRequest)
		return
	}

	// Extract caller role from JWT token
	callerRole := helper.GetRoleFromToken(r)
	if callerRole == "" {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var requestBody map[string]string
	if err := helper.DecodeJSON(r.Body, &requestBody); err != nil {
		log.Println("{UpdateUserStatus}{Decode}{Error} : ", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	newStatus := requestBody["status"]
	err := h.UserUseCase.UpdateUserStatus(email, callerRole, newStatus)
	if err != nil {
		log.Println("{UpdateUserStatus}{Error} : ", err)
		http.Error(w, err.Message, err.Code)
		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status_code": http.StatusOK,
		"message":     "User status updated successfully",
	})
}

// GetAllUsers handles GET /api/users
// O(1) handler delegation to usecase
func (h *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	users, err := h.UserUseCase.GetAllUsers()
	if err != nil {
		log.Println("{GetAllUsers}{Error} : ", err)
		http.Error(w, err.Message, err.Code)
		return
	}

	helper.WriteJSON(w, http.StatusOK, map[string]interface{}{
		"status_code": http.StatusOK,
		"message":     "Users retrieved successfully",
		"data":        users,
	})
}
