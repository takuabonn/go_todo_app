package handler

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/takuabonn/go_todo_app/entity"
	"github.com/takuabonn/go_todo_app/store"
	"github.com/takuabonn/go_todo_app/testutil"
)

func TestListTask(t *testing.T) {
	t.Parallel()
	type want struct {
		status int
		rspFile string
	}
	tests := map[string]struct {
		reqFile string
		want    want
	}{
		"ok": {
			want: want{
				status:  http.StatusOK,
				rspFile: "testdata/list_task/ok_rsp.json.golden",
			},
		},
		
	}

	for n, tt := range tests {
		tt := tt
		t.Run(n, func(t *testing.T) {
			t.Parallel()

			w := httptest.NewRecorder()
			r := httptest.NewRequest(
				http.MethodGet,
				"/tasks",
				nil,
			)

			sut := ListTask{
				Store: &store.TaskStore{
					Tasks: map[entity.TaskID]*entity.Task{
						1: {
							ID: 1,
							Title: "test1",
							Status: "todo",
							Created: time.Now(),
						},
						2: {
							ID: 2,
							Title: "test2",
							Status: "done",
							Created: time.Now(),
						},
						
					},
				},
			}
			sut.ServeHTTP(w, r)

			resp := w.Result()
			testutil.AssertResponse(t, 
				resp, tt.want.status, testutil.LoadFile(t, tt.want.rspFile),
			)

		})

	}

}