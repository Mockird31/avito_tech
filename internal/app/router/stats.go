package router

import (
	"database/sql"
	"net/http"

	statsRepository "github.com/Mockird31/avito_tech/internal/stats/repository"

	statsUsecase "github.com/Mockird31/avito_tech/internal/stats/usecase"

	statsDeliveryHttp "github.com/Mockird31/avito_tech/internal/stats/delivery/http"
	"github.com/gorilla/mux"
)

func StatsRouter(r *mux.Router, postgresConn *sql.DB) *mux.Router {
	statsRepo := statsRepository.NewRepository(postgresConn)

	statsUse := statsUsecase.NewUsecase(statsRepo)

	statsHttp := statsDeliveryHttp.NewHandler(statsUse)

	sr := r.PathPrefix("/stats").Subrouter()
	sr.HandleFunc("/assignmentsByReviewers", statsHttp.GetAssignmentsStats).Methods(http.MethodGet)
	return sr
}
