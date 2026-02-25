package database

import (
	"backend/internal/models"
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/joho/godotenv/autoload"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// GetAuditLogsOptions configures listing audit logs (limit, skip, admin-only and action-type filters).
type GetAuditLogsOptions struct {
	Limit            int    // default 50, max 50
	Skip             int    // offset for pagination
	ActionType       string // optional: filter by action.action (e.g. create, update, delete, login, register)
	IncludeAdminOnly bool   // if true (admin only), return only entries with admin_info=true
}

// GetServiceLogsOptions configures listing service logs (pagination).
type GetServiceLogsOptions struct {
	Limit int // default 50, max 50
	Skip  int // offset for pagination
}

type Service interface {
	Health() map[string]string
	Register(user *models.User, ctx context.Context) error
	FindUserByEmail(ctx context.Context, email string) (*models.User, error)
	InsertAuditLog(ctx context.Context, log *models.AuditLog) error
	GetAuditLogs(ctx context.Context, userEmail string, isAdmin bool, opts GetAuditLogsOptions) ([]models.AuditLog, int64, error)

	InsertServiceLog(ctx context.Context, log *models.ServiceLog) error
	GetServiceLogs(ctx context.Context, isAdmin bool, allowedNamespaces []string, instanceName, namespace string, opts GetServiceLogsOptions) ([]models.ServiceLog, int64, error)
	GetInstanceStatusCache(ctx context.Context, instanceName, namespace string) (status string, err error)
	SetInstanceStatusCache(ctx context.Context, instanceName, namespace, status string) error
}

type service struct {
	db *mongo.Client
}

var (
	username = os.Getenv("MONGO_DB_ATLAS_USERNAME")
	password = os.Getenv("MONGO_DB_ATLAS_PASSWORD")
)

func New() Service {
	// Use the SetServerAPIOptions() method to set the version of the Stable API on the client
	serverAPI := options.ServerAPI(options.ServerAPIVersion1)

	// Construct the URI using environment variables
	uri := fmt.Sprintf("mongodb+srv://%s:%s@paas.cgyj2kh.mongodb.net/?appName=paas", username, password)

	opts := options.Client().ApplyURI(uri).SetServerAPIOptions(serverAPI)

	// Create a new client and connect to the server (v1 driver requires context)
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	client, err := mongo.Connect(ctx, opts)
	if err != nil {
		log.Fatal(err)
	}

	// Send a ping to confirm a successful connection
	if err := client.Ping(context.TODO(), readpref.Primary()); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Pinged your deployment. You successfully connected to MongoDB!")

	return &service{
		db: client,
	}
}

func (s *service) Health() map[string]string {
	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := s.db.Ping(ctx, readpref.Primary())
	if err != nil {
		log.Fatalf("db down: %v", err)
	}

	return map[string]string{
		"message": "It's healthy",
	}
}

//AUtH

func (s *service) Register(user *models.User, ctx context.Context) error {
	collection := s.db.Database("paas").Collection("users")

	//check if user already exists
	var existingUser models.User

	err := collection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&existingUser)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			//user doesnt exist so create user
			_, err = collection.InsertOne(context.TODO(), user)
			if err != nil {
				return err
			}

			return nil
		} else {
			//we hjave a db error
			return err
		}
	} else {
		//user already exists
		if existingUser != (models.User{}) {
			return errors.New("user already exists")
		}
	}

	return nil
}

// userLoginFields is used only for decoding; we project just these fields to avoid
// decode errors from _id or date fields stored in an unexpected format in the DB.
type userLoginFields struct {
	Email    string `bson:"email"`
	Password string `bson:"password"`
	IsAdmin  bool   `bson:"is_admin"`
}

func (s *service) FindUserByEmail(ctx context.Context, email string) (*models.User, error) {
	collection := s.db.Database("paas").Collection("users")
	projection := bson.M{"email": 1, "password": 1, "is_admin": 1}
	opts := options.FindOne().SetProjection(projection)
	var fields userLoginFields
	err := collection.FindOne(ctx, bson.M{"email": email}, opts).Decode(&fields)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil, nil
		}
		return nil, err
	}
	return &models.User{
		Email:    fields.Email,
		Password: fields.Password,
		IsAdmin:  fields.IsAdmin,
	}, nil
}

func (s *service) InsertAuditLog(ctx context.Context, log *models.AuditLog) error {
	collection := s.db.Database("paas").Collection("audit_logs")
	_, err := collection.InsertOne(ctx, log)
	return err
}

func (s *service) GetAuditLogs(ctx context.Context, userEmail string, isAdmin bool, opts GetAuditLogsOptions) ([]models.AuditLog, int64, error) {
	collection := s.db.Database("paas").Collection("audit_logs")

	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 50 {
		limit = 50
	}
	skip := opts.Skip
	if skip < 0 {
		skip = 0
	}

	filter := bson.M{}
	if isAdmin {
		// Admins see all actions; optionally restrict to admin-only entries (e.g. logins)
		if opts.IncludeAdminOnly {
			filter["admin_info"] = true
		}
	} else {
		// Non-admins see only their own actions, and only non-admin entries
		filter["user_email"] = userEmail
		filter["admin_info"] = false
	}
	if opts.ActionType != "" {
		filter["action.action"] = opts.ActionType
	}

	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))
	cursor, err := collection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var logs []models.AuditLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func (s *service) InsertServiceLog(ctx context.Context, log *models.ServiceLog) error {
	collection := s.db.Database("paas").Collection("service_logs")
	_, err := collection.InsertOne(ctx, log)
	return err
}

func (s *service) GetServiceLogs(ctx context.Context, isAdmin bool, allowedNamespaces []string, instanceName, namespace string, opts GetServiceLogsOptions) ([]models.ServiceLog, int64, error) {
	collection := s.db.Database("paas").Collection("service_logs")

	limit := opts.Limit
	if limit <= 0 {
		limit = 50
	}
	if limit > 50 {
		limit = 50
	}
	skip := opts.Skip
	if skip < 0 {
		skip = 0
	}

	filter := bson.M{}
	if instanceName != "" {
		filter["instance_name"] = instanceName
	}
	if namespace != "" {
		filter["namespace"] = namespace
	} else if !isAdmin {
		if len(allowedNamespaces) == 0 {
			return nil, 0, nil
		}
		filter["namespace"] = bson.M{"$in": allowedNamespaces}
	}

	total, err := collection.CountDocuments(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	findOpts := options.Find().
		SetSort(bson.D{{Key: "timestamp", Value: -1}}).
		SetSkip(int64(skip)).
		SetLimit(int64(limit))
	cursor, err := collection.Find(ctx, filter, findOpts)
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(ctx)

	var logs []models.ServiceLog
	if err := cursor.All(ctx, &logs); err != nil {
		return nil, 0, err
	}
	return logs, total, nil
}

func (s *service) GetInstanceStatusCache(ctx context.Context, instanceName, namespace string) (string, error) {
	collection := s.db.Database("paas").Collection("instance_status_cache")
	var doc models.InstanceStatusCache
	err := collection.FindOne(ctx, bson.M{"instance_name": instanceName, "namespace": namespace}).Decode(&doc)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return "", nil
		}
		return "", err
	}
	return doc.Status, nil
}

func (s *service) SetInstanceStatusCache(ctx context.Context, instanceName, namespace, status string) error {
	collection := s.db.Database("paas").Collection("instance_status_cache")
	doc := models.InstanceStatusCache{
		InstanceName: instanceName,
		Namespace:    namespace,
		Status:       status,
		UpdatedAt:    time.Now(),
	}
	opts := options.Update().SetUpsert(true)
	_, err := collection.UpdateOne(ctx,
		bson.M{"instance_name": instanceName, "namespace": namespace},
		bson.M{"$set": doc},
		opts,
	)
	return err
}
