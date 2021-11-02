package badges

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/require"
)

func TestNewHandler(t *testing.T) {
	h := NewHandler(&service{}, &mockCache{})
	require.NotNil(t, h.(*handler).service)
	require.NotNil(t, h.(*handler).cache)
}

func TestHandlerServeHTTP(t *testing.T) {
	testRequest, err := http.NewRequest(
		http.MethodGet,
		"/v1/github/checks/krancour/foo/badge.svg",
		nil,
	)
	require.NoError(t, err)
	testBadge := CheckBadge{
		name:   "foo",
		status: CheckStatusQueued,
	}
	testCases := []struct {
		name       string
		handler    *handler
		assertions func(*http.Response)
	}{
		{
			name: "warm cache hit",
			handler: &handler{
				cache: &mockCache{
					GetWarmFn: func(string) (string, error) {
						return badgeURL(testBadge), nil // Hit
					},
				},
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusSeeOther, r.StatusCode)
				require.Equal(t, badgeURL(testBadge), r.Header.Get("Location"))
			},
		},
		{
			name: "warm cache error; service error; cold cache hit",
			handler: &handler{
				cache: &mockCache{
					GetWarmFn: func(string) (string, error) {
						return "", errors.New("something went wrong") // Error
					},
					GetColdFn: func(string) (string, error) {
						return badgeURL(testBadge), nil // Hit
					},
				},
				service: &mockService{
					CheckBadgeFn: func(
						context.Context,
						string,
						string,
						*CheckBadgeOptions,
					) (CheckBadge, error) {
						return CheckBadge{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusSeeOther, r.StatusCode)
				require.Equal(t, badgeURL(testBadge), r.Header.Get("Location"))
			},
		},
		{
			name: "warm cache error; service error; cold cache error",
			handler: &handler{
				cache: &mockCache{
					GetWarmFn: func(string) (string, error) {
						return "", errors.New("something went wrong") // Error
					},
					GetColdFn: func(string) (string, error) {
						return "", errors.New("something went wrong") // Error
					},
				},
				service: &mockService{
					CheckBadgeFn: func(
						context.Context,
						string,
						string,
						*CheckBadgeOptions,
					) (CheckBadge, error) {
						return CheckBadge{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusSeeOther, r.StatusCode)
				require.Equal(
					t,
					badgeURL(NewErrBadge(http.StatusInternalServerError)),
					r.Header.Get("Location"),
				)
			},
		},
		{
			name: "warm cache error; service error; cold cache miss",
			handler: &handler{
				cache: &mockCache{
					GetWarmFn: func(string) (string, error) {
						return "", errors.New("something went wrong") // Error
					},
					GetColdFn: func(string) (string, error) {
						return "", nil // Miss
					},
				},
				service: &mockService{
					CheckBadgeFn: func(
						context.Context,
						string,
						string,
						*CheckBadgeOptions,
					) (CheckBadge, error) {
						return CheckBadge{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusSeeOther, r.StatusCode)
				require.Equal(
					t,
					badgeURL(NewErrBadge(http.StatusInternalServerError)),
					r.Header.Get("Location"),
				)
			},
		},
		{
			name: "warm cache error; service success; cache set error",
			handler: &handler{
				cache: &mockCache{
					GetWarmFn: func(string) (string, error) {
						return "", errors.New("something went wrong") // Error
					},
					SetFn: func(string, string) error {
						return errors.New("something went wrong")
					},
				},
				service: &mockService{
					CheckBadgeFn: func(
						context.Context,
						string,
						string,
						*CheckBadgeOptions,
					) (CheckBadge, error) {
						return testBadge, nil
					},
				},
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusSeeOther, r.StatusCode)
				require.Equal(t, badgeURL(testBadge), r.Header.Get("Location"))
			},
		},
		{
			name: "warm cache error; service success; cache set success",
			handler: &handler{
				cache: &mockCache{
					GetWarmFn: func(string) (string, error) {
						return "", errors.New("something went wrong") // Error
					},
					SetFn: func(string, string) error {
						return nil
					},
				},
				service: &mockService{
					CheckBadgeFn: func(
						context.Context,
						string,
						string,
						*CheckBadgeOptions,
					) (CheckBadge, error) {
						return testBadge, nil
					},
				},
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusSeeOther, r.StatusCode)
				require.Equal(t, badgeURL(testBadge), r.Header.Get("Location"))
			},
		},
		{
			name: "warm cache miss; service error; cold cache hit",
			handler: &handler{
				cache: &mockCache{
					GetWarmFn: func(string) (string, error) {
						return "", nil // Miss
					},
					GetColdFn: func(string) (string, error) {
						return badgeURL(testBadge), nil // Hit
					},
				},
				service: &mockService{
					CheckBadgeFn: func(
						context.Context,
						string,
						string,
						*CheckBadgeOptions,
					) (CheckBadge, error) {
						return CheckBadge{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusSeeOther, r.StatusCode)
				require.Equal(t, badgeURL(testBadge), r.Header.Get("Location"))
			},
		},
		{
			name: "warm cache miss; service error; cold cache error",
			handler: &handler{
				cache: &mockCache{
					GetWarmFn: func(string) (string, error) {
						return "", nil // Miss
					},
					GetColdFn: func(string) (string, error) {
						return "", errors.New("something went wrong") // Error
					},
				},
				service: &mockService{
					CheckBadgeFn: func(
						context.Context,
						string,
						string,
						*CheckBadgeOptions,
					) (CheckBadge, error) {
						return CheckBadge{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusSeeOther, r.StatusCode)
				require.Equal(
					t,
					badgeURL(NewErrBadge(http.StatusInternalServerError)),
					r.Header.Get("Location"),
				)
			},
		},
		{
			name: "warm cache miss; service error; cold cache miss",
			handler: &handler{
				cache: &mockCache{
					GetWarmFn: func(string) (string, error) {
						return "", nil // Miss
					},
					GetColdFn: func(string) (string, error) {
						return "", nil // Miss
					},
				},
				service: &mockService{
					CheckBadgeFn: func(
						context.Context,
						string,
						string,
						*CheckBadgeOptions,
					) (CheckBadge, error) {
						return CheckBadge{}, errors.New("something went wrong")
					},
				},
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusSeeOther, r.StatusCode)
				require.Equal(
					t,
					badgeURL(NewErrBadge(http.StatusInternalServerError)),
					r.Header.Get("Location"),
				)
			},
		},
		{
			name: "warm cache miss; service success; cache set error",
			handler: &handler{
				cache: &mockCache{
					GetWarmFn: func(string) (string, error) {
						return "", nil // Miss
					},
					SetFn: func(key, value string) error {
						return errors.New("something went wrong")
					},
				},
				service: &mockService{
					CheckBadgeFn: func(
						context.Context,
						string,
						string,
						*CheckBadgeOptions,
					) (CheckBadge, error) {
						return testBadge, nil
					},
				},
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusSeeOther, r.StatusCode)
				require.Equal(t, badgeURL(testBadge), r.Header.Get("Location"))
			},
		},
		{
			name: "warm cache miss; service success; cache set success",
			handler: &handler{
				cache: &mockCache{
					GetWarmFn: func(string) (string, error) {
						return "", nil // Miss
					},
					SetFn: func(key, value string) error {
						return nil
					},
				},
				service: &mockService{
					CheckBadgeFn: func(
						context.Context,
						string,
						string,
						*CheckBadgeOptions,
					) (CheckBadge, error) {
						return testBadge, nil
					},
				},
			},
			assertions: func(r *http.Response) {
				require.Equal(t, http.StatusSeeOther, r.StatusCode)
				require.Equal(t, badgeURL(testBadge), r.Header.Get("Location"))
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			testRouter := mux.NewRouter()
			testRouter.HandleFunc(
				"/v1/github/checks/{owner}/{repo}/badge.svg",
				testCase.handler.ServeHTTP,
			).Methods(http.MethodGet)
			rr := httptest.NewRecorder()
			testRouter.ServeHTTP(rr, testRequest)
			res := rr.Result()
			defer res.Body.Close()
			testCase.assertions(res)
		})
	}
}

type mockService struct {
	CheckBadgeFn func(
		ctx context.Context,
		owner string,
		repo string,
		opts *CheckBadgeOptions,
	) (CheckBadge, error)
}

func (m *mockService) CheckBadge(
	ctx context.Context,
	owner string,
	repo string,
	opts *CheckBadgeOptions,
) (CheckBadge, error) {
	return m.CheckBadgeFn(ctx, owner, repo, opts)
}

type mockCache struct {
	SetFn     func(key string, value string) error
	GetWarmFn func(key string) (string, error)
	GetColdFn func(key string) (string, error)
}

func (m *mockCache) Set(key string, value string) error {
	return m.SetFn(key, value)
}

func (m *mockCache) GetWarm(key string) (string, error) {
	return m.GetWarmFn(key)
}

func (m *mockCache) GetCold(key string) (string, error) {
	return m.GetColdFn(key)
}
