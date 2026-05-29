package handler_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/SkyShineTH/shipyard/todo-service/handler"
	"github.com/SkyShineTH/shipyard/todo-service/model"
	"github.com/gin-gonic/gin"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := db.AutoMigrate(&model.Todo{}); err != nil {
		t.Fatalf("automigrate: %v", err)
	}
	return db
}

// injectUser replaces RequireAuth for tests: sets user_id directly on the context.
func injectUser(id uint) gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("user_id", id)
		c.Next()
	}
}

func newRouter(db *gorm.DB, userID uint) *gin.Engine {
	r := gin.New()
	h := handler.NewTodoHandler(db)
	api := r.Group("/api/v1", injectUser(userID))
	api.GET("/todos", h.GetTodos)
	api.POST("/todos", h.CreateTodo)
	api.PUT("/todos/:id", h.UpdateTodo)
	api.DELETE("/todos/:id", h.DeleteTodo)
	return r
}

func doRequest(t *testing.T, router *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()
	var b []byte
	if body != nil {
		b, _ = json.Marshal(body)
	}
	req := httptest.NewRequest(method, path, bytes.NewReader(b))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestGetTodos_ReturnsEmptyListInitially(t *testing.T) {
	t.Parallel()
	rec := doRequest(t, newRouter(newTestDB(t), 1), http.MethodGet, "/api/v1/todos", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body)
	}
	var list []model.Todo
	json.NewDecoder(rec.Body).Decode(&list)
	if len(list) != 0 {
		t.Fatalf("expected empty list, got %d items", len(list))
	}
}

func TestCreateTodo_Success(t *testing.T) {
	t.Parallel()
	r := newRouter(newTestDB(t), 1)
	rec := doRequest(t, r, http.MethodPost, "/api/v1/todos", map[string]string{"title": "buy milk"})
	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body)
	}
	var todo model.Todo
	json.NewDecoder(rec.Body).Decode(&todo)
	if todo.Title != "buy milk" {
		t.Fatalf("unexpected title: %q", todo.Title)
	}
	if todo.Completed {
		t.Fatal("new todo should not be completed")
	}
}

func TestCreateTodo_MissingTitle_Returns400(t *testing.T) {
	t.Parallel()
	rec := doRequest(t, newRouter(newTestDB(t), 1), http.MethodPost, "/api/v1/todos", map[string]string{})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d: %s", rec.Code, rec.Body)
	}
}

func TestUpdateTodo_ToggleCompleted(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)
	r := newRouter(db, 1)

	create := doRequest(t, r, http.MethodPost, "/api/v1/todos", map[string]string{"title": "task"})
	var todo model.Todo
	json.NewDecoder(create.Body).Decode(&todo)

	completed := true
	rec := doRequest(t, r, http.MethodPut, fmt.Sprintf("/api/v1/todos/%d", todo.ID),
		map[string]*bool{"completed": &completed})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body)
	}
	var updated model.Todo
	json.NewDecoder(rec.Body).Decode(&updated)
	if !updated.Completed {
		t.Fatal("expected todo to be marked completed")
	}
}

func TestDeleteTodo_Success(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)
	r := newRouter(db, 1)

	create := doRequest(t, r, http.MethodPost, "/api/v1/todos", map[string]string{"title": "delete me"})
	var todo model.Todo
	json.NewDecoder(create.Body).Decode(&todo)

	rec := doRequest(t, r, http.MethodDelete, fmt.Sprintf("/api/v1/todos/%d", todo.ID), nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d: %s", rec.Code, rec.Body)
	}
}

// Cross-user isolation: a todo created by user 1 must not be visible to,
// updatable by, or deletable by user 2.
func TestTodos_CrossUserIsolation(t *testing.T) {
	t.Parallel()
	db := newTestDB(t)

	routerUser1 := newRouter(db, 1)
	routerUser2 := newRouter(db, 2)

	// User 1 creates a todo.
	create := doRequest(t, routerUser1, http.MethodPost, "/api/v1/todos", map[string]string{"title": "user1 private"})
	if create.Code != http.StatusCreated {
		t.Fatalf("create failed: %d %s", create.Code, create.Body)
	}
	var todo model.Todo
	json.NewDecoder(create.Body).Decode(&todo)

	// User 2 cannot see it in their list.
	list := doRequest(t, routerUser2, http.MethodGet, "/api/v1/todos", nil)
	var todos []model.Todo
	json.NewDecoder(list.Body).Decode(&todos)
	if len(todos) != 0 {
		t.Fatalf("user 2 should see 0 todos, got %d", len(todos))
	}

	// User 2 cannot update it.
	completed := true
	recUpdate := doRequest(t, routerUser2, http.MethodPut, fmt.Sprintf("/api/v1/todos/%d", todo.ID),
		map[string]*bool{"completed": &completed})
	if recUpdate.Code != http.StatusNotFound {
		t.Fatalf("expected 404 on cross-user update, got %d", recUpdate.Code)
	}

	// User 2 cannot delete it.
	recDelete := doRequest(t, routerUser2, http.MethodDelete, fmt.Sprintf("/api/v1/todos/%d", todo.ID), nil)
	if recDelete.Code != http.StatusNotFound {
		t.Fatalf("expected 404 on cross-user delete, got %d", recDelete.Code)
	}

	// User 1 can still read their own todo.
	listUser1 := doRequest(t, routerUser1, http.MethodGet, "/api/v1/todos", nil)
	var user1Todos []model.Todo
	json.NewDecoder(listUser1.Body).Decode(&user1Todos)
	if len(user1Todos) != 1 {
		t.Fatalf("user 1 should see 1 todo, got %d", len(user1Todos))
	}
}
