package db_service

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DbService[DocType interface{}] interface {
	CreateDocument(ctx context.Context, id string, document *DocType) error
	FindAllDocuments(ctx context.Context) ([]DocType, error) 
	FindDocument(ctx context.Context, id string) (*DocType, error)
	FindDocumentsByCondition(ctx context.Context, filter bson.M) ([]DocType, error)
	UpdateDocument(ctx context.Context, id string, document *DocType) error
	DeleteDocument(ctx context.Context, id string) error
	Disconnect(ctx context.Context) error
}

var ErrNotFound = fmt.Errorf("document not found")
var ErrConflict = fmt.Errorf("conflict: document already exists")

type MongoServiceConfig struct {
	ServerHost string
	ServerPort int
	UserName   string
	Password   string
	DbName     string
	Collection string
	Timeout    time.Duration
}

type mongoSvc[DocType interface{}] struct {
	MongoServiceConfig
	client     atomic.Pointer[mongo.Client]
	clientLock sync.Mutex
}

func (m *mongoSvc[DocType]) FindDocumentsByCondition(ctx context.Context, filter bson.M) ([]DocType, error) {
	ctx, contextCancel := context.WithTimeout(ctx, m.Timeout)
	defer contextCancel()
	
	client, err := m.connect(ctx)
	if err != nil {
		return nil, err
	}
	
	db := client.Database(m.DbName)
	collection := db.Collection(m.Collection)
	
	// Find documents with filter
	cursor, err := collection.Find(ctx, filter)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	// Decode all documents
	var documents []DocType
	if err = cursor.All(ctx, &documents); err != nil {
		return nil, err
	}
	
	// Return empty slice if no documents found
	if documents == nil {
		documents = []DocType{}
	}
	
	return documents, nil
}

func NewMongoService[DocType interface{}](config MongoServiceConfig) DbService[DocType] {
	enviro := func(name string, defaultValue string) string {
		if value, ok := os.LookupEnv(name); ok {
			return value
		}
		return defaultValue
	}

	svc := &mongoSvc[DocType]{}
	svc.MongoServiceConfig = config

	if svc.ServerHost == "" {
		svc.ServerHost = enviro("MDM_API_MONGODB_HOST", "localhost")
	}

	if svc.ServerPort == 0 {
		port := enviro("MDM_API_MONGODB_PORT", "27017")
		if port, err := strconv.Atoi(port); err == nil {
			svc.ServerPort = port
		} else {
			log.Printf("Invalid port value: %v", port)
			svc.ServerPort = 27017
		}
	}

	if svc.UserName == "" {
		svc.UserName = enviro("MDM_API_MONGODB_USERNAME", "")
	}

	if svc.Password == "" {
		svc.Password = enviro("MDM_API_MONGODB_PASSWORD", "")
	}

	if svc.DbName == "" {
		svc.DbName = enviro("MDM_API_MONGODB_DATABASE", "mdm-patient-management")
	}

	if svc.Collection == "" {
		svc.Collection = enviro("MDM_API_MONGODB_COLLECTION", "patients")
	}

	if svc.Timeout == 0 {
		seconds := enviro("MDM_API_MONGODB_TIMEOUT_SECONDS", "10")
		if seconds, err := strconv.Atoi(seconds); err == nil {
			svc.Timeout = time.Duration(seconds) * time.Second
		} else {
			log.Printf("Invalid timeout value: %v", seconds)
			svc.Timeout = 10 * time.Second
		}
	}

	log.Printf(
		"MongoDB config: //%v@%v:%v/%v/%v",
		svc.UserName,
		svc.ServerHost,
		svc.ServerPort,
		svc.DbName,
		svc.Collection,
	)
	return svc
}

func (m *mongoSvc[DocType]) connect(ctx context.Context) (*mongo.Client, error) {
	// optimistic check
	client := m.client.Load()
	if client != nil {
		return client, nil
	}

	m.clientLock.Lock()
	defer m.clientLock.Unlock()
	// pesimistic check
	client = m.client.Load()
	if client != nil {
		return client, nil
	}

	ctx, contextCancel := context.WithTimeout(ctx, m.Timeout)
	defer contextCancel()

	var uri = fmt.Sprintf("mongodb://%v:%v", m.ServerHost, m.ServerPort)
	log.Printf("%s", "Using URI: "+uri)

	if len(m.UserName) != 0 {
		uri = fmt.Sprintf("mongodb://%v:%v@%v:%v", m.UserName, m.Password, m.ServerHost, m.ServerPort)
	}

	if client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri).SetConnectTimeout(10*time.Second)); err != nil {
		return nil, err
	} else {
		m.client.Store(client)
		return client, nil
	}
}

func (m *mongoSvc[DocType]) Disconnect(ctx context.Context) error {
	client := m.client.Load()

	if client != nil {
		m.clientLock.Lock()
		defer m.clientLock.Unlock()

		client = m.client.Load()
		defer m.client.Store(nil)
		if client != nil {
			if err := client.Disconnect(ctx); err != nil {
				return err
			}
		}
	}
	return nil
}

func (m *mongoSvc[DocType]) CreateDocument(ctx context.Context, id string, document *DocType) error {
	ctx, contextCancel := context.WithTimeout(ctx, m.Timeout)
	defer contextCancel()
	client, err := m.connect(ctx)
	if err != nil {
		return err
	}
	db := client.Database(m.DbName)
	collection := db.Collection(m.Collection)
	result := collection.FindOne(ctx, bson.D{{Key: "id", Value: id}})
	switch result.Err() {
	case nil: // no error means there is conflicting document
		return ErrConflict
	case mongo.ErrNoDocuments:
		// do nothing, this is expected
	default: // other errors - return them
		return result.Err()
	}

	_, err = collection.InsertOne(ctx, document)
	return err
}

func (m *mongoSvc[DocType]) FindAllDocuments(ctx context.Context) ([]DocType, error) {
	ctx, contextCancel := context.WithTimeout(ctx, m.Timeout)
	defer contextCancel()
	
	client, err := m.connect(ctx)
	if err != nil {
		return nil, err
	}
	
	db := client.Database(m.DbName)
	collection := db.Collection(m.Collection)
	
	// Find all documents
	cursor, err := collection.Find(ctx, bson.D{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)
	
	// Decode all documents
	var documents []DocType
	if err = cursor.All(ctx, &documents); err != nil {
		return nil, err
	}
	
	// Return empty slice if no documents found
	if documents == nil {
		documents = []DocType{}
	}
	
	return documents, nil
}

func (m *mongoSvc[DocType]) FindDocument(ctx context.Context, id string) (*DocType, error) {
	ctx, contextCancel := context.WithTimeout(ctx, m.Timeout)
	defer contextCancel()
	client, err := m.connect(ctx)
	if err != nil {
		return nil, err
	}
	db := client.Database(m.DbName)
	collection := db.Collection(m.Collection)
	result := collection.FindOne(ctx, bson.D{{Key: "id", Value: id}})
	switch result.Err() {
	case nil:
	case mongo.ErrNoDocuments:
		return nil, ErrNotFound
	default: // other errors - return them
		return nil, result.Err()
	}
	var document *DocType
	if err := result.Decode(&document); err != nil {
		return nil, err
	}
	return document, nil
}

func (m *mongoSvc[DocType]) UpdateDocument(ctx context.Context, id string, document *DocType) error {
	ctx, contextCancel := context.WithTimeout(ctx, m.Timeout)
	defer contextCancel()
	client, err := m.connect(ctx)
	if err != nil {
		return err
	}
	db := client.Database(m.DbName)
	collection := db.Collection(m.Collection)
	result := collection.FindOne(ctx, bson.D{{Key: "id", Value: id}})
	switch result.Err() {
	case nil:
	case mongo.ErrNoDocuments:
		return ErrNotFound
	default: // other errors - return them
		return result.Err()
	}
	_, err = collection.ReplaceOne(ctx, bson.D{{Key: "id", Value: id}}, document)
	return err
}

func (m *mongoSvc[DocType]) DeleteDocument(ctx context.Context, id string) error {
	ctx, contextCancel := context.WithTimeout(ctx, m.Timeout)
	defer contextCancel()
	client, err := m.connect(ctx)
	if err != nil {
		return err
	}
	db := client.Database(m.DbName)
	collection := db.Collection(m.Collection)
	result := collection.FindOne(ctx, bson.D{{Key: "id", Value: id}})
	switch result.Err() {
	case nil:
	case mongo.ErrNoDocuments:
		return ErrNotFound
	default: // other errors - return them
		return result.Err()
	}
	_, err = collection.DeleteOne(ctx, bson.D{{Key: "id", Value: id}})
	return err
}