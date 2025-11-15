package router

import (
	"database/sql"
	"net/http"

	userRepository "github.com/Mockird31/avito_tech/internal/user/repository"

	userUsecase "github.com/Mockird31/avito_tech/internal/user/usecase"

	userDeliveryHttp "github.com/Mockird31/avito_tech/internal/user/delivery/http"
	"github.com/gorilla/mux"
)

func UserRouter(r *mux.Router, postgresConn *sql.DB) *mux.Router {
	userRepo := userRepository.NewRepository(postgresConn)

	userUse := userUsecase.NewUsecase(userRepo)

	userHttp := userDeliveryHttp.NewHandler(userUse)

	sr := r.PathPrefix("/users").Subrouter()
	sr.HandleFunc("/setIsActive", userHttp.SetUserIsActive).Methods(http.MethodPost)
	return sr
}
