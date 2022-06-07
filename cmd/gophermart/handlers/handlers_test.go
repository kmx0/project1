package handlers

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-playground/assert/v2"
	"github.com/kmx0/project1/internal/config"
	"github.com/kmx0/project1/internal/types"
	"github.com/stretchr/testify/require"
)

// var store repositories.Repository
// var cfg config.Config

// func SetRepository(s repositories.Repository) {
// 	store = s
// }

func TestHandleRegister(t *testing.T) {
	// s := storage.NewInMemory(config.Config{})
	// SetRepository(s)
	type wantStruct struct {
		statusCode int
		// counter     types.Counter
	}
	// var store repositories.Repository

	router := SetupRouter(config.Config{})
	tests := []struct {
		name string
		req  string
		body types.User
		want wantStruct
	}{
		{
			name: "success Register",
			req:  "/api/user/register",
			body: types.User{Login: "user1",Password: "PAss1"},
			want: wantStruct{
				statusCode: 200,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {

			// logrus.Info(tt.req)
			w := httptest.NewRecorder()
			// req, _ := http.NewRequest("GET", "/ping", nil)
			// bodyReader := bytes.NewReader(
			bodyBytes, err := json.Marshal(tt.body)
			require.NoError(t, err)
			bodyReader := bytes.NewReader(bodyBytes)
			request, _ := http.NewRequest(http.MethodPost, tt.req, bodyReader)

			router.ServeHTTP(w, request)
			res := w.Result()

			assert.Equal(t, tt.want.statusCode, res.StatusCode)
			err = res.Body.Close()
			require.NoError(t, err)
			// mapresult, err := ioutil.ReadAll(res.Body)
			// HandleCounter(tt.args.w, tt.args.r)
		})
	}
}
