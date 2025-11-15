package http

import (
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
		switch err {
		case entity.ErrAuthorOrTeamNotExist:
			statusCode = http.StatusNotFound
		case entity.ErrPullRequestExist:
			statusCode = http.StatusConflict
		default:
			statusCode = http.StatusInternalServerError
		}
		json.WriteErrorJson(w, statusCode, err.Error())
		return
	}

	json.WriteJSON(w, http.StatusCreated, &entity.PullRequestResponse{PullRequest: pullRequest}, nil)
}
