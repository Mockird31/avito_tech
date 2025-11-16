package http

import (
    "bytes"
    "encoding/json"
    "errors"
    "net/http"
    "net/http/httptest"
    "testing"

    "github.com/Mockird31/avito_tech/internal/entity"
    mock_user "github.com/Mockird31/avito_tech/mocks/user"
    "github.com/stretchr/testify/assert"
    "github.com/stretchr/testify/mock"
    "github.com/stretchr/testify/require"
)

func TestHandler_SetUserIsActive(t *testing.T) {
    type args struct {
        body string
    }
    tests := []struct {
        name           string
        args           args
        mockSetup      func(m *mock_user.MockIUsecase)
        wantStatusCode int
        wantUser       *entity.UserResponse
        wantBody       string
    }{
        {
            name: "invalid_json_body",
            args: args{
                body: `{"user_id":"u1","is_active":`, // broken json
            },
            mockSetup:      func(m *mock_user.MockIUsecase) {},
            wantStatusCode: http.StatusInternalServerError,
            wantBody:       `{"code":500,"message":"failed to parse json"}`,
        },
        {
            name: "usecase_user_not_found",
            args: args{
                body: `{"user_id":"missing","is_active":true}`,
            },
            mockSetup: func(m *mock_user.MockIUsecase) {
                m.EXPECT().
                    SetIsActive(mock.Anything, mock.AnythingOfType("*entity.UserUpdateActive")).
                    Return(nil, entity.ErrUserNotFound)
            },
            wantStatusCode: http.StatusNotFound,
            wantBody:       `{"code":404,"message":"resource not found"}`,
        },
        {
            name: "usecase_internal_error",
            args: args{
                body: `{"user_id":"u1","is_active":false}`,
            },
            mockSetup: func(m *mock_user.MockIUsecase) {
                m.EXPECT().
                    SetIsActive(mock.Anything, mock.AnythingOfType("*entity.UserUpdateActive")).
                    Return(nil, errors.New("db failure"))
            },
            wantStatusCode: http.StatusInternalServerError,
            wantBody:       `{"code":500,"message":"db failure"}`,
        },
        {
            name: "success",
            args: args{
                body: `{"user_id":"u1","is_active":true}`,
            },
            mockSetup: func(m *mock_user.MockIUsecase) {
                res := &entity.User{
                    UserId:   "u1",
                    Username: "alice",
                    TeamName: "teamA",
                    IsActive: true,
                }
                m.EXPECT().
                    SetIsActive(mock.Anything, mock.AnythingOfType("*entity.UserUpdateActive")).
                    Return(res, nil)
            },
            wantStatusCode: http.StatusOK,
            wantUser: &entity.UserResponse{User: &entity.User{
                UserId:   "u1",
                Username: "alice",
                TeamName: "teamA",
                IsActive: true,
            }},
        },
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            m := mock_user.NewMockIUsecase(t)
            if tt.mockSetup != nil {
                tt.mockSetup(m)
            }

            h := NewHandler(m)

            req := httptest.NewRequest(http.MethodPost, "/users/setIsActive", bytes.NewBufferString(tt.args.body))
            rr := httptest.NewRecorder()

            http.HandlerFunc(h.SetUserIsActive).ServeHTTP(rr, req)

            require.Equal(t, tt.wantStatusCode, rr.Code)

            if tt.wantUser != nil {
                var got entity.UserResponse
                require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &got))
                assert.Equal(t, tt.wantUser, &got)
            }
            if tt.wantBody != "" {
                assert.JSONEq(t, tt.wantBody, rr.Body.String())
            }
        })
    }
}

func TestHandler_GetUserReviews(t *testing.T) {
    tests := []struct {
        name           string
        query          string
        mockSetup      func(m *mock_user.MockIUsecase)
        wantStatusCode int
        wantResp       *entity.ReviewerPullRequests
        wantBody       string
    }{
        {
            name:           "missing_user_id",
            query:          "",
            mockSetup:      func(m *mock_user.MockIUsecase) {},
            wantStatusCode: http.StatusNotFound,
            wantBody:       `{"code":404,"message":"NOT_FOUND"}`,
        },
        {
            name:  "usecase_user_not_found",
            query: "?user_id=missing",
            mockSetup: func(m *mock_user.MockIUsecase) {
                m.EXPECT().
                    GetUserReview(mock.Anything, "missing").
                    Return(nil, "", entity.ErrUserNotFound)
            },
            wantStatusCode: http.StatusNotFound,
            wantBody:       `{"code":404,"message":"resource not found"}`,
        },
        {
            name:  "usecase_internal_error",
            query: "?user_id=u1",
            mockSetup: func(m *mock_user.MockIUsecase) {
                m.EXPECT().
                    GetUserReview(mock.Anything, "u1").
                    Return(nil, "", errors.New("db failure"))
            },
            wantStatusCode: http.StatusInternalServerError,
            wantBody:       `{"code":500,"message":"db failure"}`,
        },
        {
            name:  "success",
            query: "?user_id=u1",
            mockSetup: func(m *mock_user.MockIUsecase) {
                prs := []*entity.PullRequestShort{
                    {Id: "pr1", PrName: "Fix bug", AuthorId: "a1", Status: "OPEN"},
                    {Id: "pr2", PrName: "Add feature", AuthorId: "a2", Status: "MERGED"},
                }
                m.EXPECT().
                    GetUserReview(mock.Anything, "u1").
                    Return(prs, "u1", nil)
            },
            wantStatusCode: http.StatusOK,
            wantResp: &entity.ReviewerPullRequests{
                UserId: "u1",
                PullRequests: []*entity.PullRequestShort{
                    {Id: "pr1", PrName: "Fix bug", AuthorId: "a1", Status: "OPEN"},
                    {Id: "pr2", PrName: "Add feature", AuthorId: "a2", Status: "MERGED"},
                },
            },
        },
    }

    for _, tt := range tests {
        tt := tt
        t.Run(tt.name, func(t *testing.T) {
            m := mock_user.NewMockIUsecase(t)
            if tt.mockSetup != nil {
                tt.mockSetup(m)
            }

            h := NewHandler(m)

            req := httptest.NewRequest(http.MethodGet, "/users/getReview"+tt.query, bytes.NewBuffer(nil))
            rr := httptest.NewRecorder()

            http.HandlerFunc(h.GetUserReviews).ServeHTTP(rr, req)

            require.Equal(t, tt.wantStatusCode, rr.Code)

            if tt.wantResp != nil {
                var got entity.ReviewerPullRequests
                require.NoError(t, json.Unmarshal(rr.Body.Bytes(), &got))
                assert.Equal(t, tt.wantResp, &got)
            }
            if tt.wantBody != "" {
                assert.JSONEq(t, tt.wantBody, rr.Body.String())
            }
        })
    }
}