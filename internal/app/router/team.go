package router

import (
	"database/sql"
	"net/http"

	teamRepository "github.com/Mockird31/avito_tech/internal/team/repository"
	userRepository "github.com/Mockird31/avito_tech/internal/user/repository"

	teamUsecase "github.com/Mockird31/avito_tech/internal/team/usecase"

	teamDeliveryHttp "github.com/Mockird31/avito_tech/internal/team/delivery/http"
	"github.com/gorilla/mux"
)

func TeamRouter(r *mux.Router, postgresConn *sql.DB) *mux.Router {
	teamRepo := teamRepository.NewRepository(postgresConn)
	userRepo := userRepository.NewRepository(postgresConn)

	teamUse := teamUsecase.NewUsecase(teamRepo, userRepo)

	teamHttp := teamDeliveryHttp.NewHandler(teamUse)

	sr := r.PathPrefix("/team").Subrouter()
	sr.HandleFunc("/add", teamHttp.AddTeam).Methods(http.MethodPost)
	sr.HandleFunc("/get", teamHttp.GetTeam).Methods(http.MethodGet)
	return sr
}
