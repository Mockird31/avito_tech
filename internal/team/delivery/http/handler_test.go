package http

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/Mockird31/avito_tech/internal/entity"
	mock_team "github.com/Mockird31/avito_tech/mocks/team"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandler_AddTeam(t *testing.T) {
	type args struct {
		body string
	}
	tests := []struct {
		name           string
		args           args
		mockSetup      func(m *mock_team.MockIUsecase)
		wantStatusCode int
		wantTeam       *entity.Team
		wantBody       string
	}{
		{
			name: "invalid_json_body",
			args: args{
				body: `{"team_name": "alpha", "members": [`, // broken json
			},
			mockSetup:      func(m *mock_team.MockIUsecase) {},
			wantStatusCode: http.StatusInternalServerError,
			wantBody:       `"failed to parse request"`,
		},
		{
			name: "usecase_error",
			args: args{
				body: `{"team_name": "alpha", "members": [{"user_id":"u1","username":"alice","is_active":true}]}`,
			},
			mockSetup: func(m *mock_team.MockIUsecase) {
				m.EXPECT().
					AddTeam(mock.Anything, mock.AnythingOfType("*entity.Team")).
					Return(nil, errors.New("team_name already exists"))
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       `"team_name already exists"`,
		},
		{
			name: "success",
			args: args{
				body: `{"team_name": "alpha", "members": [{"user_id":"u1","username":"alice","is_active":true},{"user_id":"u2","username":"bob","is_active":false}]}`,
			},
			mockSetup: func(m *mock_team.MockIUsecase) {
				res := &entity.Team{
					TeamName: "alpha",
					Members: []*entity.TeamMember{
						{UserID: "u1", Username: "alice", IsActive: true},
						{UserID: "u2", Username: "bob", IsActive: false},
					},
				}
				m.EXPECT().
					AddTeam(mock.Anything, mock.AnythingOfType("*entity.Team")).
					Return(res, nil)
			},
			wantStatusCode: http.StatusCreated,
			wantTeam: &entity.Team{
				TeamName: "alpha",
				Members: []*entity.TeamMember{
					{UserID: "u1", Username: "alice", IsActive: true},
					{UserID: "u2", Username: "bob", IsActive: false},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			m := mock_team.NewMockIUsecase(t)
			if tt.mockSetup != nil {
				tt.mockSetup(m)
			}

			h := NewHandler(m)

			req := httptest.NewRequest(http.MethodPost, "/team/add", bytes.NewBufferString(tt.args.body))
			rr := httptest.NewRecorder()

			http.HandlerFunc(h.AddTeam).ServeHTTP(rr, req)

			require.Equal(t, tt.wantStatusCode, rr.Code)

			if tt.wantTeam != nil {
				var got entity.Team
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &got))
				assert.Equal(t, tt.wantTeam, &got)
			}
			if tt.wantBody != "" {
				assert.JSONEq(t, tt.wantBody, rr.Body.String())
			}
		})
	}
}

func TestHandler_GetTeam(t *testing.T) {
	tests := []struct {
		name           string
		query          string
		mockSetup      func(m *mock_team.MockIUsecase)
		wantStatusCode int
		wantTeam       *entity.Team
		wantBody       string
	}{
		{
			name:           "missing_team_name",
			query:          "",
			mockSetup:      func(m *mock_team.MockIUsecase) {},
			wantStatusCode: http.StatusNotFound,
			wantBody:       `"NOT_FOUND"`,
		},
		{
			name:  "usecase_error",
			query: "?team_name=alpha",
			mockSetup: func(m *mock_team.MockIUsecase) {
				m.EXPECT().
					GetTeam(mock.Anything, "alpha").
					Return(nil, errors.New("NOT_FOUND"))
			},
			wantStatusCode: http.StatusNotFound,
			wantBody:       `"NOT_FOUND"`,
		},
		{
			name:  "success",
			query: "?team_name=alpha",
			mockSetup: func(m *mock_team.MockIUsecase) {
				res := &entity.Team{
					TeamName: "alpha",
					Members: []*entity.TeamMember{
						{UserID: "u1", Username: "alice", IsActive: true},
					},
				}
				m.EXPECT().
					GetTeam(mock.Anything, "alpha").
					Return(res, nil)
			},
			wantStatusCode: http.StatusOK,
			wantTeam: &entity.Team{
				TeamName: "alpha",
				Members: []*entity.TeamMember{
					{UserID: "u1", Username: "alice", IsActive: true},
				},
			},
		},
	}

	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			m := mock_team.NewMockIUsecase(t)
			if tt.mockSetup != nil {
				tt.mockSetup(m)
			}

			h := NewHandler(m)

			req := httptest.NewRequest(http.MethodGet, "/team/get"+tt.query, nil)
			rr := httptest.NewRecorder()

			http.HandlerFunc(h.GetTeam).ServeHTTP(rr, req)

			require.Equal(t, tt.wantStatusCode, rr.Code)

			if tt.wantTeam != nil {
				var got entity.Team
				require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &got))
				assert.Equal(t, tt.wantTeam, &got)
			}
			if tt.wantBody != "" {
				assert.JSONEq(t, tt.wantBody, rr.Body.String())
			}
		})
	}
}
