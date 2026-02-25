package server

import (
	"backend/internal/database"
	"backend/internal/kube"
	"backend/internal/models"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// emailToNamespace converts an email to a valid Kubernetes namespace name (DNS-1123 label).
func emailToNamespace(email string) string {
	s := strings.ToLower(strings.TrimSpace(email))
	s = strings.ReplaceAll(s, "@", "-at-")
	s = strings.ReplaceAll(s, ".", "-")
	if s == "" {
		return "default"
	}
	return s
}

// getUserNamespaceAndAdmin returns the current user's namespace (derived from email) and whether they are admin.
// Used to scope instance access: non-admins are restricted to their own namespace.
func (s *Server) getUserNamespaceAndAdmin(c *gin.Context) (userNS string, isAdmin bool) {
	email, _ := c.Get("user_email")
	admin, _ := c.Get("user_is_admin")
	emailStr, _ := email.(string)
	isAdmin, _ = admin.(bool)
	return emailToNamespace(emailStr), isAdmin
}

func (s *Server) RegisterRoutes() http.Handler {
	r := gin.Default()

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:5173", "http://192.214.178.2", "http://ryan-paas.stackit.gg", "https://ryan-paas.stackit.gg", "http://ryanpaas.stackit.gg", "https://ryanpaas.stackit.gg"}, // Add your frontend URL
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: true, // Enable cookies/auth
	}))

	r.GET("/", s.HelloWorldHandler)

	r.GET("/health", s.healthHandler)

	authGroup := r.Group("/auth")
	{
		authGroup.POST("/register", s.registerHandler)
		authGroup.POST("/login", s.loginHandler)
	}

	apiGroup := r.Group("/api", s.JWTMiddleware())
	{
		apiGroup.GET("/instances", s.getAllInstancesHandler)       //get all instances
		apiGroup.GET("/instances/:id", s.getInstanceHandler)       //get single instance
		apiGroup.POST("/instances", s.createInstanceHandler)       // create new instace
		apiGroup.PATCH("/instances/:id", s.updateInstanceHandler)  // update instance (partial)
		apiGroup.DELETE("/instances/:id", s.deleteInstanceHandler) //delete one
		apiGroup.GET("/audit-logs", s.getAuditLogsHandler)
		apiGroup.GET("/instances/:id/service-logs", s.getInstanceServiceLogsHandler)
		apiGroup.GET("/service-logs", s.getServiceLogsHandler)
	}
	//helo

	return r
}

func (s *Server) getInstanceHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "provide the instance name you would like to get",
		})
		return
	}

	userNS, isAdmin := s.getUserNamespaceAndAdmin(c)
	var namespace string
	if isAdmin {
		namespace = c.Query("namespace")
		if namespace == "" {
			namespace = "default"
		}
	} else {
		namespace = userNS
	}

	obj, err := s.kubeClient.Resource(kube.RedisFailOver).Namespace(namespace).Get(c.Request.Context(), id, v1.GetOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get redis failover",
			"details": err.Error(),
		})
		return
	}

	var instance models.RedisInstance
	instance.ConvertUnstructuredToRedisInstace(obj)
	if instance.Status == "Unknown" {
		instance.Status = kube.GetStatusFromStatefulSets(c.Request.Context(), s.kubeClient, instance.Namespace, instance.Name, instance.RedisReplicas, instance.SentinelReplicas)
	}

	port, _ := kube.GetRedisServicePort(c.Request.Context(), s.kubeClient, instance.Namespace, instance.Name)
	err = instance.GetConnectionInfo(port)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get connection info",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message":  "instance fetched succesfully",
		"instance": instance,
	})

}

func (s *Server) deleteInstanceHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "provide the instance name you would like to delete",
		})
		return
	}

	userNS, isAdmin := s.getUserNamespaceAndAdmin(c)
	var namespace string
	if isAdmin {
		namespace = c.Query("namespace")
		if namespace == "" {
			namespace = "default"
		}
	} else {
		namespace = userNS
	}

	// Fetch instance before delete so we can log what was deleted
	obj, getErr := s.kubeClient.Resource(kube.RedisFailOver).Namespace(namespace).Get(c.Request.Context(), id, v1.GetOptions{})
	var deletedDetails string
	if getErr == nil {
		var before models.RedisInstance
		before.ConvertUnstructuredToRedisInstace(obj)
		deletedDetails = fmt.Sprintf("redisReplicas: %d, sentinelReplicas: %d", before.RedisReplicas, before.SentinelReplicas)
	}

	err := s.kubeClient.Resource(kube.RedisFailOver).Namespace(namespace).Delete(c.Request.Context(), id, v1.DeleteOptions{})

	if err != nil {

		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "instance not found",
				"details": err.Error(),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to delete redis failover",
			"details": err.Error(),
		})
		return
	}

	email, _ := c.Get("user_email")
	if e, ok := email.(string); ok {
		s.logAudit(c, e, models.Action{
			Action:    "delete",
			Name:      id,
			Namespace: namespace,
			Details:   deletedDetails,
		}, false)
	}
	c.JSON(http.StatusOK, gin.H{
		"message": "instance deleted succesfully",
		"id":      id,
	})
}

func (s *Server) updateInstanceHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "provide the instance name you would like to update",
		})
		return
	}

	userNS, isAdmin := s.getUserNamespaceAndAdmin(c)
	var namespace string
	if isAdmin {
		namespace = c.Query("namespace")
		if namespace == "" {
			namespace = "default"
		}
	} else {
		namespace = userNS
	}

	var req models.UpdateInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}
	if isAdmin && req.Namespace != nil && *req.Namespace != "" {
		namespace = *req.Namespace
	}

	if req.RedisReplicas == nil && req.SentinelReplicas == nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "provide at least one of redisReplicas or sentinelReplicas to update",
		})
		return
	}
	if req.RedisReplicas != nil && *req.RedisReplicas <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "redisReplicas must be greater than 0",
		})
		return
	}
	if req.SentinelReplicas != nil && *req.SentinelReplicas <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "sentinelReplicas must be greater than 0",
		})
		return
	}

	obj, err := s.kubeClient.Resource(kube.RedisFailOver).Namespace(namespace).Get(c.Request.Context(), id, v1.GetOptions{})
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			c.JSON(http.StatusNotFound, gin.H{
				"error":   "instance not found",
				"details": err.Error(),
			})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get redis failover",
			"details": err.Error(),
		})
		return
	}

	// Capture current state for audit purposes before applying changes
	var before models.RedisInstance
	before.ConvertUnstructuredToRedisInstace(obj)

	if req.RedisReplicas != nil {
		if err := unstructured.SetNestedField(obj.Object, int64(*req.RedisReplicas), "spec", "redis", "replicas"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to set redis replicas",
				"details": err.Error(),
			})
			return
		}
	}
	if req.SentinelReplicas != nil {
		if err := unstructured.SetNestedField(obj.Object, int64(*req.SentinelReplicas), "spec", "sentinel", "replicas"); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to set sentinel replicas",
				"details": err.Error(),
			})
			return
		}
	}

	updated, err := s.kubeClient.Resource(kube.RedisFailOver).Namespace(namespace).Update(c.Request.Context(), obj, v1.UpdateOptions{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to update redis failover",
			"details": err.Error(),
		})
		return
	}

	var instance models.RedisInstance
	instance.ConvertUnstructuredToRedisInstace(updated)
	if instance.Status == "Unknown" {
		instance.Status = kube.GetStatusFromStatefulSets(c.Request.Context(), s.kubeClient, instance.Namespace, instance.Name, instance.RedisReplicas, instance.SentinelReplicas)
	}
	port, _ := kube.GetRedisServicePort(c.Request.Context(), s.kubeClient, instance.Namespace, instance.Name)
	if err := instance.GetConnectionInfo(port); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get connection info",
			"details": err.Error(),
		})
		return
	}

	email, _ := c.Get("user_email")
	if e, ok := email.(string); ok {
		changes := make([]string, 0, 2)
		if req.RedisReplicas != nil && before.RedisReplicas != *req.RedisReplicas {
			changes = append(changes, fmt.Sprintf("redisReplicas: %d -> %d", before.RedisReplicas, *req.RedisReplicas))
		}
		if req.SentinelReplicas != nil && before.SentinelReplicas != *req.SentinelReplicas {
			changes = append(changes, fmt.Sprintf("sentinelReplicas: %d -> %d", before.SentinelReplicas, *req.SentinelReplicas))
		}

		details := strings.Join(changes, ", ")
		s.logAudit(c, e, models.Action{
			Action:    "update",
			Name:      id,
			Namespace: namespace,
			Details:   details,
		}, false)
	}
	c.JSON(http.StatusOK, gin.H{
		"message":  "instance updated successfully",
		"instance": instance,
	})
}

func (s *Server) getAllInstancesHandler(c *gin.Context) {
	userNS, isAdmin := s.getUserNamespaceAndAdmin(c)

	var list *unstructured.UnstructuredList
	var err error
	if isAdmin {
		list, err = s.kubeClient.Resource(kube.RedisFailOver).List(c.Request.Context(), v1.ListOptions{})
	} else {
		list, err = s.kubeClient.Resource(kube.RedisFailOver).Namespace(userNS).List(c.Request.Context(), v1.ListOptions{})
	}
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to list redis failovers",
			"details": err.Error(),
		})
		return
	}

	//convert each cr intop a redis instance
	instances := make([]models.RedisInstance, 0, len(list.Items))

	for _, item := range list.Items {
		var instance models.RedisInstance
		instance.ConvertUnstructuredToRedisInstace(&item)
		if instance.Status == "Unknown" {
			instance.Status = kube.GetStatusFromStatefulSets(c.Request.Context(), s.kubeClient, instance.Namespace, instance.Name, instance.RedisReplicas, instance.SentinelReplicas)
		}
		port, _ := kube.GetRedisServicePort(c.Request.Context(), s.kubeClient, instance.Namespace, instance.Name)
		err = instance.GetConnectionInfo(port)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":   "failed to get connection info",
				"details": err.Error(),
			})
			return
		}
		instances = append(instances, instance)
	}

	c.JSON(http.StatusOK, gin.H{
		"instances": instances,
		"count":     len(instances),
	})
}

func (s *Server) createInstanceHandler(c *gin.Context) {
	userNS, isAdmin := s.getUserNamespaceAndAdmin(c)

	var req models.CreateInstanceRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}
	if req.RedisReplicas <= 0 {
		req.RedisReplicas = 3
	}

	if req.SentinelReplicas <= 0 {
		req.SentinelReplicas = 3
	}

	if isAdmin {
		if req.Namespace == "" {
			req.Namespace = "default"
		}
	} else {
		req.Namespace = userNS
	}

	name := req.Name
	if name == "" {
		name = "redis-" + time.Now().Format("20060102150405")
	}

	if err := kube.EnsureNamespace(c.Request.Context(), s.kubeClient, req.Namespace); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to ensure namespace",
			"details": err.Error(),
		})
		return
	}

	//build the failover
	rf := kube.BuildRedisFailover(name, req.Namespace, req.RedisReplicas, req.SentinelReplicas)

	created, err := s.kubeClient.
		Resource(kube.RedisFailOver).
		Namespace(req.Namespace).
		Create(c.Request.Context(), rf, v1.CreateOptions{})

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":      "failed to create redis failover",
			"details":    err.Error(),
			"full-error": err,
		})
		return
	}

	now := time.Now()
	resp := models.RedisInstance{
		ID:               name,
		Name:             name,
		Namespace:        req.Namespace,
		RedisReplicas:    req.RedisReplicas,
		SentinelReplicas: req.SentinelReplicas,
		Status:           "PROVISIONING",
		CreatedAt:        now,
		UpdatedAt:        now,
	}

	err = resp.GetConnectionInfo(0)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get connection info",
			"details": err.Error(),
		})
		return
	}

	_ = created

	email, _ := c.Get("user_email")
	if e, ok := email.(string); ok {
		details := fmt.Sprintf("redisReplicas: %d, sentinelReplicas: %d", req.RedisReplicas, req.SentinelReplicas)
		s.logAudit(c, e, models.Action{
			Action:    "create",
			Name:      name,
			Namespace: req.Namespace,
			Details:   details,
		}, false)
	}
	c.JSON(http.StatusCreated, resp)

}

func (s *Server) registerHandler(c *gin.Context) {

	var req models.RegisterRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	if req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "email or password are required",
		})
		return
	}

	user := &models.User{
		ID:        primitive.NewObjectID(),
		Email:     req.Email,
		Password:  req.Password,
		IsAdmin:   false,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	err := user.HashPassword()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"erorr": "failed to hash password",
		})
	}

	if err := s.db.Register(user, c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to register user",
			"message": err.Error(),
		})
		return
	}

	s.logAudit(c, req.Email, models.Action{Action: "register", Name: "", Namespace: ""}, false)
	c.JSON(http.StatusOK, gin.H{
		"message": "user registered successfully",
		"user":    *user,
	})

}

func (s *Server) loginHandler(c *gin.Context) {
	var req models.RegisterRequest
	if err := c.ShouldBindBodyWithJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid request body",
		})
		return
	}

	if req.Email == "" || req.Password == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "email or password are required",
		})
		return
	}

	user, err := s.db.FindUserByEmail(c.Request.Context(), req.Email)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to look up user",
			"details": err.Error(),
		})
		return
	}
	if user == nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid credentials",
		})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "invalid credentials",
		})
		return
	}

	token, err := s.generateToken(req.Email, user.IsAdmin)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "failed to generate token",
		})
		return
	}

	s.logAudit(c, req.Email, models.Action{Action: "login", Details: "User logged in Successfully"}, true)
	c.JSON(http.StatusOK, gin.H{
		"token": token,
		"user": gin.H{
			"email":    req.Email,
			"is_admin": user.IsAdmin,
		},
	})
}

func (s *Server) HelloWorldHandler(c *gin.Context) {
	resp := make(map[string]string)
	resp["message"] = "Hello World"

	c.JSON(http.StatusOK, resp)
}

func (s *Server) healthHandler(c *gin.Context) {
	c.JSON(http.StatusOK, s.db.Health())
}

func (s *Server) getAuditLogsHandler(c *gin.Context) {
	email := c.GetString("user_email")
	if email == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "user email not found"})
		return
	}
	admin, _ := c.Get("user_is_admin")
	isAdmin := false
	if b, ok := admin.(bool); ok {
		isAdmin = b
	}

	page := 1
	if p := c.Query("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	limit := 50
	skip := (page - 1) * limit
	actionType := strings.TrimSpace(c.Query("type"))
	includeAdminOnly := c.Query("admin_only") == "true" && isAdmin

	opts := database.GetAuditLogsOptions{
		Limit:            limit,
		Skip:             skip,
		ActionType:       actionType,
		IncludeAdminOnly: includeAdminOnly,
	}
	logs, total, err := s.db.GetAuditLogs(c.Request.Context(), email, isAdmin, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get audit logs",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"audit_logs": logs,
		"count":      len(logs),
		"total":      total,
		"page":       page,
	})
}

func (s *Server) getInstanceServiceLogsHandler(c *gin.Context) {
	id := c.Param("id")
	if id == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "instance id required"})
		return
	}

	userNS, isAdmin := s.getUserNamespaceAndAdmin(c)
	namespace := c.Query("namespace")
	if namespace == "" {
		namespace = userNS
	}
	if !isAdmin && namespace != userNS {
		c.JSON(http.StatusForbidden, gin.H{"error": "access denied to this instance's service logs"})
		return
	}

	page := 1
	if p := c.Query("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	limit := 50
	skip := (page - 1) * limit

	allowedNamespaces := []string{userNS}
	if isAdmin {
		allowedNamespaces = nil
	}
	opts := database.GetServiceLogsOptions{Limit: limit, Skip: skip}
	logs, total, err := s.db.GetServiceLogs(c.Request.Context(), isAdmin, allowedNamespaces, id, namespace, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get service logs",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service_logs": logs,
		"count":        len(logs),
		"total":        total,
		"page":         page,
	})
}

func (s *Server) getServiceLogsHandler(c *gin.Context) {
	userNS, isAdmin := s.getUserNamespaceAndAdmin(c)

	page := 1
	if p := c.Query("page"); p != "" {
		if n, err := strconv.Atoi(p); err == nil && n > 0 {
			page = n
		}
	}
	limit := 50
	skip := (page - 1) * limit
	instanceFilter := strings.TrimSpace(c.Query("instance"))
	namespaceFilter := strings.TrimSpace(c.Query("namespace"))

	allowedNamespaces := []string{userNS}
	if isAdmin {
		allowedNamespaces = nil
	}
	opts := database.GetServiceLogsOptions{Limit: limit, Skip: skip}
	logs, total, err := s.db.GetServiceLogs(c.Request.Context(), isAdmin, allowedNamespaces, instanceFilter, namespaceFilter, opts)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":   "failed to get service logs",
			"details": err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"service_logs": logs,
		"count":        len(logs),
		"total":        total,
		"page":         page,
	})
}
