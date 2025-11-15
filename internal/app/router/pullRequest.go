package router

import (
	"database/sql"
	"net/http"

	prRepository "github.com/Mockird31/avito_tech/internal/pullRequest/repository"
	teamRepository "github.com/Mockird31/avito_tech/internal/team/repository"
	userRepository "github.com/Mockird31/avito_tech/internal/user/repository"

	prUsecase "github.com/Mockird31/avito_tech/internal/pullRequest/usecase"

	prDeliveryHttp "github.com/Mockird31/avito_tech/internal/pullRequest/delivery/http"
	"github.com/gorilla/mux"
)

func PullRequestRouter(r *mux.Router, postgresConn *sql.DB) *mux.Router {
	teamRepo := teamRepository.NewRepository(postgresConn)
	userRepo := userRepository.NewRepository(postgresConn)
	prRepo := prRepository.NewRepository(postgresConn)

	prUse := prUsecase.NewUsecase(prRepo, userRepo, teamRepo)

	prHttp := prDeliveryHttp.NewHandler(prUse)

	sr := r.PathPrefix("/pullRequest").Subrouter()
	sr.HandleFunc("/create", prHttp.CreatePullRequest).Methods(http.MethodPost)
	sr.HandleFunc("/merge", prHttp.MergePullRequest).Methods(http.MethodPost)
	sr.HandleFunc("/reassign", prHttp.ReassignPullRequest).Methods(http.MethodPost)
	return sr
}
