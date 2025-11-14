package http

import (
	"fmt"
	"net/http"

	"github.com/Mockird31/avito_tech/internal/entity"
	"github.com/Mockird31/avito_tech/internal/team"
	json "github.com/Mockird31/avito_tech/pkg/json"
)

type Handler struct {
	usecase team.IUsecase
}

func NewHandler(usecase team.IUsecase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

func (h *Handler) AddTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	var addTeamRequest entity.Team
	err := json.ReadJSON(w, r, &addTeamRequest)
	if err != nil {
		json.WriteErrorJson(w, http.StatusInternalServerError, "failed to parse request")
		return
	}
	fmt.Println(addTeamRequest.TeamName)
	resultTeam, err := h.usecase.AddTeam(ctx, &addTeamRequest)
	if err != nil {
		json.WriteErrorJson(w, http.StatusNotFound, err.Error())
		return
	}

	json.WriteJSON(w, http.StatusCreated, resultTeam, nil)
}

func (h *Handler) GetTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	teamName := r.URL.Query().Get("team_name")
	if teamName == "" {
		json.WriteErrorJson(w, http.StatusNotFound, "NOT_FOUND")
		return
	}

	resultTeam, err := h.usecase.GetTeam(ctx, teamName)
	if err != nil {
		json.WriteErrorJson(w, http.StatusNotFound, "NOT_FOUND")
		return
	}

	json.WriteJSON(w, http.StatusOK, resultTeam, nil)
}
