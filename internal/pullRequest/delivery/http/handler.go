package http

import (
	"errors"
	"net/http"

	"github.com/Mockird31/avito_tech/internal/entity"
	pullrequest "github.com/Mockird31/avito_tech/internal/pullRequest"
	json "github.com/Mockird31/avito_tech/pkg/json"
)

type Handler struct {
	usecase pullrequest.IUsecase
}

func NewHandler(usecase pullrequest.IUsecase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

func (h *Handler) CreatePullRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var createPullRequest entity.PullRequest

	err := json.ReadJSON(w, r, &createPullRequest)
	if err != nil {
		json.WriteErrorJson(w, http.StatusInternalServerError, "failed to parse request")
		return
	}

	pullRequest, err := h.usecase.CreatePullRequest(ctx, &createPullRequest)
	if err != nil {
		var statusCode int
		switch {
		case errors.Is(err, entity.ErrAuthorOrTeamNotExist):
			statusCode = http.StatusNotFound
		case errors.Is(err, entity.ErrPullRequestExist):
			statusCode = http.StatusConflict
		default:
			statusCode = http.StatusInternalServerError
		}
		json.WriteErrorJson(w, statusCode, err.Error())
		return
	}

	json.WriteJSON(w, http.StatusCreated, &entity.PullRequestResponse{PullRequest: pullRequest}, nil)
}

func (h *Handler) MergePullRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var mergePullRequest entity.PullRequest

	err := json.ReadJSON(w, r, &mergePullRequest)
	if err != nil {
		json.WriteErrorJson(w, http.StatusInternalServerError, "failed to parse request")
		return
	}

	pullRequest, err := h.usecase.MergePullRequest(ctx, &mergePullRequest)
	if err != nil {
		var statusCode int
		switch {
		case errors.Is(err, entity.ErrPullRequestNotExist):
			statusCode = http.StatusNotFound
		default:
			statusCode = http.StatusInternalServerError
		}
		json.WriteErrorJson(w, statusCode, err.Error())
		return
	}

	json.WriteJSON(w, http.StatusOK, &entity.PullRequestResponse{PullRequest: pullRequest}, nil)
}

func (h *Handler) ReassignPullRequest(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	var reassignPullRequest entity.PullRequestReassignRequest

	err := json.ReadJSON(w, r, &reassignPullRequest)
	if err != nil {
		json.WriteErrorJson(w, http.StatusInternalServerError, "failed to parse request")
		return
	}

	pullRequest, newReviewer, err := h.usecase.ReassignPullRequest(ctx, &reassignPullRequest)
	if err != nil {
		var statusCode int
		switch {
		case errors.Is(err, entity.ErrPullRequestNotExist):
			statusCode = http.StatusNotFound
		case errors.Is(err, entity.ErrUserNotFound):
			statusCode = http.StatusNotFound
		case errors.Is(err, entity.ErrRequestAlreadyMerged):
			statusCode = http.StatusConflict
		default:
			statusCode = http.StatusInternalServerError
		}
		json.WriteErrorJson(w, statusCode, err.Error())
		return
	}

	json.WriteJSON(w, http.StatusOK, &entity.PullRequestReassignResponse{PullRequest: pullRequest, ReplacedBy: newReviewer}, nil)
}
