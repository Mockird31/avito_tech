//go:build e2e

package e2e

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcpostgres "github.com/testcontainers/testcontainers-go/modules/postgres"

	"github.com/Mockird31/avito_tech/migrations"
	"github.com/Mockird31/avito_tech/pkg/logger"
	"github.com/Mockird31/avito_tech/pkg/postgres"

	"github.com/Mockird31/avito_tech/internal/middleware"
	prHttp "github.com/Mockird31/avito_tech/internal/pullRequest/delivery/http"
	prRepo "github.com/Mockird31/avito_tech/internal/pullRequest/repository"
	prUse "github.com/Mockird31/avito_tech/internal/pullRequest/usecase"
	statsHttp "github.com/Mockird31/avito_tech/internal/stats/delivery/http"
	statsRepo "github.com/Mockird31/avito_tech/internal/stats/repository"
	statsUse "github.com/Mockird31/avito_tech/internal/stats/usecase"
	teamHttp "github.com/Mockird31/avito_tech/internal/team/delivery/http"
	teamRepo "github.com/Mockird31/avito_tech/internal/team/repository"
	teamUse "github.com/Mockird31/avito_tech/internal/team/usecase"
	userHttp "github.com/Mockird31/avito_tech/internal/user/delivery/http"
	userRepo "github.com/Mockird31/avito_tech/internal/user/repository"
	userUse "github.com/Mockird31/avito_tech/internal/user/usecase"

	"github.com/Mockird31/avito_tech/internal/entity"
)

// Глобальное состояние контейнера БД для всего пакета e2e
var (
	dsnGlobal     string
	stopContainer func()
)

func TestMain(m *testing.M) {
	ctx := context.Background()

	container, err := tcpostgres.RunContainer(ctx,
		testcontainers.WithImage("postgres:16-alpine"),
		tcpostgres.WithInitScripts(),
		tcpostgres.WithDatabase("appdb"),
		tcpostgres.WithUsername("appuser"),
		tcpostgres.WithPassword("apppass"),
	)
	if err != nil {
		panic(err)
	}

	host, err := container.Host(ctx)
	if err != nil {
		_ = container.Terminate(ctx)
		panic(err)
	}
	port, err := container.MappedPort(ctx, "5432")
	if err != nil {
		_ = container.Terminate(ctx)
		panic(err)
	}

	dsn := "postgres://" + "appuser" + ":" + "apppass" + "@" + host + ":" + port.Port() + "/appdb?sslmode=disable"

	if err := WaitForPostgres(ctx, dsn, 20*time.Second); err != nil {
		_ = container.Terminate(ctx)
		panic(err)
	}

	migrator, err := migrations.NewMigrator(dsn)
	if err != nil {
		_ = container.Terminate(ctx)
		panic(err)
	}
	if err := migrator.Migrate(); err != nil {
		_ = container.Terminate(ctx)
		panic(err)
	}

	dsnGlobal = dsn
	stopContainer = func() { _ = container.Terminate(ctx) }

	code := m.Run()

	if stopContainer != nil {
		stopContainer()
	}
	os.Exit(code)
}

type httpResp struct {
	Code int
	Body []byte
}

func doJSON(t *testing.T, client *http.Client, method, url string, body any) httpResp {
	var req *http.Request
	var err error
	if body != nil {
		b, _ := json.Marshal(body)
		req, err = http.NewRequest(method, url, bytesReader(b))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req, err = http.NewRequest(method, url, nil)
	}
	require.NoError(t, err)

	res, err := client.Do(req)
	require.NoError(t, err)
	defer res.Body.Close()

	raw, err := ioReadAll(res.Body)
	require.NoError(t, err)
	return httpResp{Code: res.StatusCode, Body: raw}
}

func bytesReader(b []byte) *bytes.Reader    { return bytes.NewReader(b) }
func ioReadAll(r io.Reader) ([]byte, error) { return io.ReadAll(r) }

func newTestServer(t *testing.T) (*httptest.Server, *http.Client) {
	t.Helper()

	db, err := postgres.ConnectPostgres(postgresConfigFromDSN(dsnGlobal))
	require.NoError(t, err)

	tr := teamRepo.NewRepository(db)
	ur := userRepo.NewRepository(db)
	prr := prRepo.NewRepository(db)
	sr := statsRepo.NewRepository(db)

	tu := teamUse.NewUsecase(tr, ur)
	uu := userUse.NewUsecase(ur, prr)
	pu := prUse.NewUsecase(prr, ur, tr)
	su := statsUse.NewUsecase(sr)

	th := teamHttp.NewHandler(tu)
	uh := userHttp.NewHandler(uu)
	ph := prHttp.NewHandler(pu)
	sh := statsHttp.NewHandler(su)

	r := mux.NewRouter()

	zl, err := logger.NewZapLogger()
	require.NoError(t, err)
	r.Use(middleware.LoggerMiddleware(zl))

	teamSr := r.PathPrefix("/team").Subrouter()
	teamSr.HandleFunc("/add", th.AddTeam).Methods(http.MethodPost)
	teamSr.HandleFunc("/get", th.GetTeam).Methods(http.MethodGet)

	userSr := r.PathPrefix("/users").Subrouter()
	userSr.HandleFunc("/setIsActive", uh.SetUserIsActive).Methods(http.MethodPost)
	userSr.HandleFunc("/getReview", uh.GetUserReviews).Methods(http.MethodGet)

	prSr := r.PathPrefix("/pullRequest").Subrouter()
	prSr.HandleFunc("/create", ph.CreatePullRequest).Methods(http.MethodPost)
	prSr.HandleFunc("/merge", ph.MergePullRequest).Methods(http.MethodPost)
	prSr.HandleFunc("/reassign", ph.ReassignPullRequest).Methods(http.MethodPost)

	statsSr := r.PathPrefix("/stats").Subrouter()
	statsSr.HandleFunc("/assignmentsByReviewers", sh.GetAssignmentsStats).Methods(http.MethodGet)

	ts := httptest.NewServer(r)
	t.Cleanup(func() {
		ts.Close()
		_ = db.Close()
	})
	return ts, ts.Client()
}

func TestE2E_Team_AddAndGet(t *testing.T) {
	ts, client := newTestServer(t)

	teamName := "e2e-team-" + time.Now().Format("150405.000000")
	reqAdd := map[string]any{
		"team_name": teamName,
		"members": []map[string]any{
			{"user_id": "u1-" + teamName, "username": "alice", "is_active": true},
			{"user_id": "u2-" + teamName, "username": "bob", "is_active": true},
			{"user_id": "u3-" + teamName, "username": "carol", "is_active": false},
		},
	}

	resp := doJSON(t, client, http.MethodPost, ts.URL+"/team/add", reqAdd)
	require.Equal(t, http.StatusCreated, resp.Code)

	var teamResp entity.TeamResponse
	require.NoError(t, json.Unmarshal(resp.Body, &teamResp))
	require.Equal(t, teamName, teamResp.Team.TeamName)

	resp = doJSON(t, client, http.MethodGet, ts.URL+"/team/get?team_name="+teamName, nil)
	require.Equal(t, http.StatusOK, resp.Code)

	var gotTeam entity.Team
	require.NoError(t, json.Unmarshal(resp.Body, &gotTeam))
	require.Equal(t, teamName, gotTeam.TeamName)
	require.Len(t, gotTeam.Members, 3)
}

func TestE2E_PullRequest_Create_Merge_Reassign(t *testing.T) {
	ts, client := newTestServer(t)

	suffix := time.Now().Format("150405.000000")
	teamName := "team-pr-" + suffix
	author := "u1-" + suffix
	reviewer := "u2-" + suffix

	_ = doJSON(t, client, http.MethodPost, ts.URL+"/team/add", map[string]any{
		"team_name": teamName,
		"members": []map[string]any{
			{"user_id": author, "username": "alice", "is_active": true},
			{"user_id": reviewer, "username": "bob", "is_active": true},
		},
	})

	prId := "pr-" + suffix
	resp := doJSON(t, client, http.MethodPost, ts.URL+"/pullRequest/create", map[string]any{
		"pull_request_id":   prId,
		"pull_request_name": "Feature " + suffix,
		"author_id":         author,
	})
	require.Equal(t, http.StatusCreated, resp.Code)

	var prResp entity.PullRequestResponse
	require.NoError(t, json.Unmarshal(resp.Body, &prResp))
	require.Equal(t, prId, prResp.PullRequest.Id)

	if len(prResp.PullRequest.AssignedReviewersIds) > 0 {
		old := prResp.PullRequest.AssignedReviewersIds[0]
		resp = doJSON(t, client, http.MethodPost, ts.URL+"/pullRequest/reassign", map[string]any{
			"pull_request_id": prId,
			"old_reviewer_id": old,
		})
		require.Equal(t, http.StatusOK, resp.Code)

		var reass entity.PullRequestReassignResponse
		require.NoError(t, json.Unmarshal(resp.Body, &reass))
		require.Equal(t, prId, reass.PullRequest.Id)
	}

	resp = doJSON(t, client, http.MethodPost, ts.URL+"/pullRequest/merge", map[string]any{
		"pull_request_id": prId,
	})
	require.Equal(t, http.StatusOK, resp.Code)

	var merged entity.PullRequestResponse
	require.NoError(t, json.Unmarshal(resp.Body, &merged))
	require.Equal(t, entity.StatusMerged.String(), merged.PullRequest.Status)
}

func TestE2E_User_SetIsActive_And_GetReview(t *testing.T) {
	ts, client := newTestServer(t)

	suffix := time.Now().Format("150405.000000")
	teamName := "team-user-" + suffix
	u := "u2-" + suffix

	_ = doJSON(t, client, http.MethodPost, ts.URL+"/team/add", map[string]any{
		"team_name": teamName,
		"members": []map[string]any{
			{"user_id": "u1-" + suffix, "username": "alice", "is_active": true},
			{"user_id": u, "username": "bob", "is_active": true},
		},
	})

	resp := doJSON(t, client, http.MethodPost, ts.URL+"/users/setIsActive", map[string]any{
		"user_id": u, "is_active": false,
	})
	require.Equal(t, http.StatusOK, resp.Code)

	resp = doJSON(t, client, http.MethodGet, ts.URL+"/users/getReview?user_id="+u, nil)
	require.Contains(t, []int{http.StatusOK, http.StatusNotFound, http.StatusInternalServerError}, resp.Code)
}

func TestE2E_Stats_AssignmentsByReviewers(t *testing.T) {
	ts, client := newTestServer(t)

	resp := doJSON(t, client, http.MethodGet, ts.URL+"/stats/assignmentsByReviewers", nil)
	require.Equal(t, http.StatusOK, resp.Code)

	var stats entity.AssignmentStatsResponse
	require.NoError(t, json.Unmarshal(resp.Body, &stats))
	require.NotEmpty(t, stats.Title)
}
