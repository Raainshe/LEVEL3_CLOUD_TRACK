package server

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"backend/internal/database"
	"backend/internal/kube"
	"backend/internal/models"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
)

// mockDB implements database.Service for tests (no-op audit, no real DB).
type mockDB struct {
	loginUser *models.User // if set, FindUserByEmail returns this user for matching email
}

func (m *mockDB) Health() map[string]string { return map[string]string{"message": "ok"} }
func (m *mockDB) Register(*models.User, context.Context) error { return nil }
func (m *mockDB) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	if m.loginUser != nil && m.loginUser.Email == email {
		return m.loginUser, nil
	}
	return nil, nil
}
func (m *mockDB) InsertAuditLog(context.Context, *models.AuditLog) error { return nil }
func (m *mockDB) GetAuditLogs(context.Context, string, bool, database.GetAuditLogsOptions) ([]models.AuditLog, int64, error) {
	return nil, 0, nil
}
func (m *mockDB) InsertServiceLog(context.Context, *models.ServiceLog) error { return nil }
func (m *mockDB) GetServiceLogs(context.Context, bool, []string, string, string, database.GetServiceLogsOptions) ([]models.ServiceLog, int64, error) {
	return nil, 0, nil
}
func (m *mockDB) GetInstanceStatusCache(context.Context, string, string) (string, error) { return "", nil }
func (m *mockDB) SetInstanceStatusCache(context.Context, string, string, string) error { return nil }

// we can exercise the HTTP handlers without talking to a real cluster.
func newTestServerWithFakeKube(t *testing.T) *Server {
	t.Helper()

	scheme := runtime.NewScheme()
	listKinds := map[schema.GroupVersionResource]string{
		kube.RedisFailOver: "RedisFailoverList",
	}
	fakeClient := dynamicfake.NewSimpleDynamicClientWithCustomListKinds(scheme, listKinds)

	return &Server{
		kubeClient:    fakeClient,
		db:            &mockDB{},
		jwtSecret:     "test-secret",
		jwtTTLMinutes: 60,
	}
}

func mustHashPassword(t *testing.T, password string) []byte {
	t.Helper()
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		t.Fatalf("failed to hash password: %v", err)
	}
	return hash
}

func TestHelloWorldHandler(t *testing.T) {
	s := &Server{}
	r := gin.New()
	r.GET("/", s.HelloWorldHandler)

	req, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := `{"message":"Hello World"}`
	if strings.TrimSpace(rr.Body.String()) != expected {
		t.Errorf("Handler returned unexpected body: got %q want %q", rr.Body.String(), expected)
	}
}

func TestCreateInstanceHandler(t *testing.T) {
	s := newTestServerWithFakeKube(t)
	r := gin.New()
	r.POST("/instances", s.createInstanceHandler)

	body := `{"name":"test-instance","redisReplicas":3,"sentinelReplicas":3}`
	req, err := http.NewRequest(http.MethodPost, "/instances", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusCreated)
	}

	var resp models.RedisInstance
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.ID != "test-instance" {
		t.Errorf("unexpected id: got %q want %q", resp.ID, "test-instance")
	}
	if resp.Name != "test-instance" {
		t.Errorf("unexpected name: got %q want %q", resp.Name, "test-instance")
	}
	if resp.Namespace != "default" {
		t.Errorf("unexpected namespace: got %q want %q", resp.Namespace, "default")
	}
	if resp.RedisReplicas != 3 {
		t.Errorf("unexpected redisReplicas: got %d want %d", resp.RedisReplicas, 3)
	}
	if resp.SentinelReplicas != 3 {
		t.Errorf("unexpected sentinelReplicas: got %d want %d", resp.SentinelReplicas, 3)
	}
}

func TestDeleteInstanceHandler(t *testing.T) {
	s := newTestServerWithFakeKube(t)

	// Seed the fake Kubernetes client with a RedisFailover so delete succeeds.
	const (
		namespace = "default"
		name      = "test-instance"
	)
	rf := kube.BuildRedisFailover(name, namespace, 3, 3)
	if _, err := s.kubeClient.
		Resource(kube.RedisFailOver).
		Namespace(namespace).
		Create(context.Background(), rf, v1.CreateOptions{}); err != nil {
		t.Fatalf("failed to seed fake kube client: %v", err)
	}

	r := gin.New()
	r.DELETE("/instances/:id", s.deleteInstanceHandler)

	req, err := http.NewRequest(http.MethodDelete, "/instances/test-instance", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resp map[string]string
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp["message"] != "instance deleted succesfully" {
		t.Errorf("unexpected message: got %q want %q", resp["message"], "instance deleted succesfully")
	}
	if resp["id"] != "test-instance" {
		t.Errorf("unexpected id: got %q want %q", resp["id"], "test-instance")
	}
}

func TestGetAllInstancesHandler(t *testing.T) {
	s := newTestServerWithFakeKube(t)

	// Seed one RedisFailover so list has data.
	const (
		namespace = "default"
		name      = "test-instance"
	)
	rf := kube.BuildRedisFailover(name, namespace, 2, 2)
	if _, err := s.kubeClient.
		Resource(kube.RedisFailOver).
		Namespace(namespace).
		Create(context.Background(), rf, v1.CreateOptions{}); err != nil {
		t.Fatalf("failed to seed fake kube client: %v", err)
	}

	r := gin.New()
	r.GET("/instances", s.getAllInstancesHandler)

	req, err := http.NewRequest(http.MethodGet, "/instances", nil)
	if err != nil {
		t.Fatal(err)
	}

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resp struct {
		Instances []models.RedisInstance `json:"instances"`
		Count     int                    `json:"count"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if resp.Count != 1 {
		t.Errorf("unexpected count: got %d want %d", resp.Count, 1)
	}
	if len(resp.Instances) != 1 {
		t.Fatalf("expected 1 instance, got %d", len(resp.Instances))
	}

	inst := resp.Instances[0]
	if inst.ID != name {
		t.Errorf("unexpected id: got %q want %q", inst.ID, name)
	}
	if inst.Name != name {
		t.Errorf("unexpected name: got %q want %q", inst.Name, name)
	}
	if inst.Namespace != namespace {
		t.Errorf("unexpected namespace: got %q want %q", inst.Namespace, namespace)
	}
}

func TestUpdateInstanceHandler(t *testing.T) {
	t.Setenv("REDIS_GATEWAY_HOST", "localhost")
	t.Setenv("REDIS_GATEWAY_PORT", "6379")

	s := newTestServerWithFakeKube(t)

	const (
		namespace = "default"
		name      = "test-instance"
	)
	rf := kube.BuildRedisFailover(name, namespace, 3, 3)
	if _, err := s.kubeClient.
		Resource(kube.RedisFailOver).
		Namespace(namespace).
		Create(context.Background(), rf, v1.CreateOptions{}); err != nil {
		t.Fatalf("failed to seed fake kube client: %v", err)
	}

	r := gin.New()
	r.PATCH("/instances/:id", s.updateInstanceHandler)

	body := `{"redisReplicas":5}`
	req, err := http.NewRequest(http.MethodPatch, "/instances/test-instance", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var resp struct {
		Message  string               `json:"message"`
		Instance models.RedisInstance `json:"instance"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}
	if resp.Message != "instance updated successfully" {
		t.Errorf("unexpected message: got %q", resp.Message)
	}
	if resp.Instance.RedisReplicas != 5 {
		t.Errorf("unexpected redisReplicas in response: got %d want 5", resp.Instance.RedisReplicas)
	}
	if resp.Instance.SentinelReplicas != 3 {
		t.Errorf("sentinelReplicas should be unchanged: got %d want 3", resp.Instance.SentinelReplicas)
	}

	// Assert the resource in the cluster was actually updated
	updated, err := s.kubeClient.Resource(kube.RedisFailOver).Namespace(namespace).Get(context.Background(), name, v1.GetOptions{})
	if err != nil {
		t.Fatalf("failed to get updated resource: %v", err)
	}
	redisReplicas, found, _ := unstructured.NestedInt64(updated.Object, "spec", "redis", "replicas")
	if !found || redisReplicas != 5 {
		t.Errorf("spec.redis.replicas: got %v (found=%v) want 5", redisReplicas, found)
	}
	sentinelReplicas, found, _ := unstructured.NestedInt64(updated.Object, "spec", "sentinel", "replicas")
	if !found || sentinelReplicas != 3 {
		t.Errorf("spec.sentinel.replicas should be unchanged: got %v (found=%v) want 3", sentinelReplicas, found)
	}
}

func TestLoginAndProtectedRoute(t *testing.T) {
	gin.SetMode(gin.TestMode)

	loginUser := &models.User{
		Email:    "admin@example.com",
		Password: string(mustHashPassword(t, "password123")),
		IsAdmin:  true,
	}
	s := &Server{
		db:            &mockDB{loginUser: loginUser},
		jwtSecret:     "test-secret",
		jwtTTLMinutes: 60,
	}

	r := gin.New()
	r.POST("/auth/login", s.loginHandler)
	r.GET("/api/protected", s.JWTMiddleware(), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "ok"})
	})

	// login with correct credentials
	body := `{"email":"admin@example.com","password":"password123"}`
	req, err := http.NewRequest(http.MethodPost, "/auth/login", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	r.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Fatalf("login returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	var loginResp struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &loginResp); err != nil {
		t.Fatalf("failed to unmarshal login response: %v", err)
	}
	if loginResp.Token == "" {
		t.Fatalf("expected non-empty token")
	}

	// protected route without token -> 401
	reqNoToken, err := http.NewRequest(http.MethodGet, "/api/protected", nil)
	if err != nil {
		t.Fatal(err)
	}
	rrNoToken := httptest.NewRecorder()
	r.ServeHTTP(rrNoToken, reqNoToken)
	if status := rrNoToken.Code; status != http.StatusUnauthorized {
		t.Errorf("protected route without token: got %v want %v", status, http.StatusUnauthorized)
	}

	// protected route with token -> 200
	reqWithToken, err := http.NewRequest(http.MethodGet, "/api/protected", nil)
	if err != nil {
		t.Fatal(err)
	}
	reqWithToken.Header.Set("Authorization", "Bearer "+loginResp.Token)
	rrWithToken := httptest.NewRecorder()
	r.ServeHTTP(rrWithToken, reqWithToken)
	if status := rrWithToken.Code; status != http.StatusOK {
		t.Errorf("protected route with token: got %v want %v", status, http.StatusOK)
	}
}
