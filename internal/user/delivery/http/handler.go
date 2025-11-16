package http

import (
	"errors"
	"net/http"

	"github.com/Mockird31/avito_tech/internal/entity"
	"github.com/Mockird31/avito_tech/internal/user"
	"github.com/asaskevich/govalidator"

	json "github.com/Mockird31/avito_tech/pkg/json"
)

type Handler struct {
	usecase user.IUsecase
}

func NewHandler(usecase user.IUsecase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

func (h *Handler) SetUserIsActive(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var userActiveRequest entity.UserUpdateActive

	err := json.ReadJSON(w, r, &userActiveRequest)
	if err != nil {
		json.WriteErrorJson(w, http.StatusInternalServerError, "failed to parse json")
		return
	}

	isValid, err := govalidator.ValidateStruct(userActiveRequest)
	if err != nil {
		json.WriteErrorJson(w, http.StatusBadRequest, "failed to parse request")
		return
	}

	if !isValid {
		json.WriteErrorJson(w, http.StatusBadRequest, "wrong json")
		return
	}

	user, err := h.usecase.SetIsActive(ctx, &userActiveRequest)
	if err != nil {
		var statusCode int
		switch {
		case errors.Is(err, entity.ErrUserNotFound):
			statusCode = http.StatusNotFound
		default:
			statusCode = http.StatusInternalServerError
		}
		json.WriteErrorJson(w, statusCode, err.Error())
		return
	}

	json.WriteJSON(w, http.StatusOK, &entity.UserResponse{User: user}, nil)
}

func (h *Handler) GetUserReviews(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	userId := r.URL.Query().Get("user_id")
	if userId == "" {
		json.WriteErrorJson(w, http.StatusNotFound, "NOT_FOUND")
		return
	}

	pullRequests, userId, err := h.usecase.GetUserReview(ctx, userId)
	if err != nil {
		var statusCode int
		switch {
		case errors.Is(err, entity.ErrUserNotFound):
			statusCode = http.StatusNotFound
		default:
			statusCode = http.StatusInternalServerError
		}

		json.WriteErrorJson(w, statusCode, err.Error())
		return
	}

	json.WriteJSON(w, http.StatusOK, &entity.ReviewerPullRequests{UserId: userId, PullRequests: pullRequests}, nil)
}

func (h *Handler) DeactivateTeamUsers(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var deactivateUsers entity.DeactivateUsers

	err := json.ReadJSON(w, r, &deactivateUsers)
	if err != nil {
		json.WriteErrorJson(w, http.StatusInternalServerError, "failed to parse json")
		return
	}

	isValid, err := govalidator.ValidateStruct(deactivateUsers)
	if err != nil {
		json.WriteErrorJson(w, http.StatusBadRequest, "failed to parse request")
		return
	}

	if !isValid {
		json.WriteErrorJson(w, http.StatusBadRequest, "wrong json")
		return
	}

	deactivateUsersResp, err := h.usecase.DeactivateTeamUsers(ctx, &deactivateUsers)
	if err != nil {
		var status int
		switch {
		case errors.Is(err, entity.ErrUserNotFound):
			status = http.StatusNotFound
		case errors.Is(err, entity.ErrUsersNotSameTeam):
			status = http.StatusBadRequest
		default:
			status = http.StatusInternalServerError
		}
		json.WriteErrorJson(w, status, err.Error())
		return
	}

	json.WriteJSON(w, http.StatusOK, &entity.DeactivateUsersResponse{DeactivateUsers: deactivateUsersResp}, nil)
}
