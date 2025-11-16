package http

import (
	"net/http"

	"github.com/Mockird31/avito_tech/internal/entity"
	"github.com/Mockird31/avito_tech/internal/stats"
	json "github.com/Mockird31/avito_tech/pkg/json"
)

type Handler struct {
	statsUsecase stats.IUsecase
}

func NewHandler(statsUsecase stats.IUsecase) *Handler {
	return &Handler{
		statsUsecase: statsUsecase,
	}
}

func (h *Handler) GetAssignmentsStats(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	assignmentStats, err := h.statsUsecase.GetAssignmentsStatsByReviewers(ctx)
	if err != nil {
		json.WriteErrorJson(w, http.StatusBadRequest, "failed to get stats")
		return
	}

	json.WriteJSON(w, http.StatusOK, &entity.AssignmentStatsResponse{Statistics: assignmentStats}, nil)
}
