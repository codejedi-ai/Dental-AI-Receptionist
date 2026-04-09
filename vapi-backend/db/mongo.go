package db

import (
	"context"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Mongo struct {
	client *mongo.Client
	db     *mongo.Database
}

func NewMongo(ctx context.Context, uri, dbName string) (*Mongo, error) {
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(uri))
	if err != nil {
		return nil, err
	}
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}
	return &Mongo{client: client, db: client.Database(dbName)}, nil
}

func (m *Mongo) Close() {
	m.client.Disconnect(context.Background())
}

func (m *Mongo) GetClinicInfo(ctx context.Context, topic string) string {
	var doc bson.M
	err := m.db.Collection("settings").FindOne(ctx, bson.M{"_id": "clinic"}).Decode(&doc)
	if err != nil {
		return m.fallbackInfo(topic)
	}

	keyMap := map[string]string{
		"hours":              "hours",
		"location":           "address",
		"address":            "address",
		"insurance":          "insurance",
		"emergency":          "emergency",
		"cancellation":       "cancellationPolicy",
		"cancellationpolicy": "cancellationPolicy",
		"newpatient":         "newPatientInfo",
		"newpatientinfo":     "newPatientInfo",
		"payment":            "payment",
		"services":           "services",
		"service":            "services",
		"dentists":           "dentists",
		"dentist":            "dentists",
		"general":            "name",
	}

	topicLower := strings.ToLower(strings.TrimSpace(topic))

	if field, ok := keyMap[topicLower]; ok {
		if val, ok := doc[field].(string); ok && val != "" {
			return val
		}
	}
	for k, v := range keyMap {
		if strings.HasPrefix(topicLower, k) {
			if val, ok := doc[v].(string); ok && val != "" {
				return val
			}
		}
	}
	return m.fallbackInfo(topic)
}

func (m *Mongo) fallbackInfo(topic string) string {
	fallbacks := map[string]string{
		"hours":        "Monday through Friday 8 AM to 6 PM, Saturday 9 AM to 2 PM. Closed on Sundays.",
		"address":      "123 Smile Street, Suite 100, Toronto, ON M5H 2N2, Canada.",
		"insurance":    "We accept most major dental insurance plans including Sun Life, Manulife, Canada Life, and Blue Cross. Please call to verify your specific plan.",
		"emergency":    "For dental emergencies outside clinic hours, please visit the nearest emergency room or call 911. For severe pain, knocked-out teeth, or heavy bleeding, seek immediate medical attention.",
		"services":     "We offer general dentistry including cleanings, fillings, crowns, bridges, root canals, extractions, whitening, and cosmetic dentistry.",
		"dentists":     "Our dentists include Dr. Michael Park, Dr. Priya Sharma, and Dr. Sarah Chen.",
	}
	if val, ok := fallbacks[strings.ToLower(topic)]; ok {
		return val
	}
	return "I don't have specific information on that. Please call the clinic directly for more details."
}

func (m *Mongo) LogCall(ctx context.Context, summary *string, duration *int64, status string) error {
	doc := bson.M{
		"status":    status,
		"createdAt": time.Now(),
	}
	if summary != nil {
		doc["summary"] = *summary
	}
	if duration != nil {
		doc["durationSeconds"] = *duration
	}
	_, err := m.db.Collection("call_logs").InsertOne(ctx, doc)
	return err
}
