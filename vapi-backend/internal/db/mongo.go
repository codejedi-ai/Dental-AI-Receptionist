package db

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

// ─── Email Verifications ───────────────────────────────────────

func (m *Mongo) InitiateVerification(ctx context.Context, email string) (string, error) {
	tokenBytes := make([]byte, 24)
	rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)

	// Invalidate previous tokens
	m.db.Collection("verifications").UpdateMany(ctx,
		bson.M{"email": email},
		bson.M{"$set": bson.M{"expired": true}})

	now := time.Now()
	expiresAt := now.Add(10 * time.Minute)

	_, err := m.db.Collection("verifications").InsertOne(ctx, bson.M{
		"email":     email,
		"token":     token,
		"verified":  false,
		"createdAt": now,
		"expiresAt": expiresAt,
	})
	return token, err
}

func (m *Mongo) CheckVerification(ctx context.Context, email string) string {
	// Check verified
	var doc bson.M
	err := m.db.Collection("verifications").FindOne(ctx,
		bson.M{"email": email, "verified": true}).Decode(&doc)
	if err == nil {
		return "verified"
	}

	// Check pending (not expired)
	err = m.db.Collection("verifications").FindOne(ctx,
		bson.M{
			"email":     email,
			"verified":  false,
			"expired":   bson.M{"$ne": true},
			"expiresAt": bson.M{"$gt": time.Now()},
		}).Decode(&doc)
	if err == nil {
		return "pending"
	}
	return "expired"
}

func (m *Mongo) MarkVerified(ctx context.Context, token string) bool {
	res, err := m.db.Collection("verifications").UpdateOne(ctx,
		bson.M{
			"token":     token,
			"expiresAt": bson.M{"$gt": time.Now()},
		},
		bson.M{"$set": bson.M{
			"verified":   true,
			"verifiedAt": time.Now(),
		}})
	return err == nil && res.MatchedCount > 0
}

// ─── Clinic Settings ──────────────────────────────────────────

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
	}

	topicLower := ""
	for _, r := range topic {
		topicLower += string(r)
	}

	for k, v := range keyMap {
		if topicLower == k || len(topicLower) >= 3 && len(k) >= 3 {
			if val, ok := doc[v].(string); ok {
				return val
			}
		}
	}
	return m.fallbackInfo(topic)
}

func (m *Mongo) fallbackInfo(topic string) string {
	fallbacks := map[string]string{
		"hours":        "Monday through Friday 8 AM to 6 PM, Saturday 9 AM to 2 PM. Closed on Sundays and statutory holidays.",
		"location":     "Please check our website for the full address.",
		"insurance":    "We accept most major dental insurance plans. Please bring your insurance card to your appointment.",
		"emergency":    "For after-hours dental emergencies, please go to your nearest emergency department.",
		"cancellation": "We ask for at least 24 hours notice for cancellations.",
		"newpatient":   "New patients are always welcome! Please arrive 15 minutes early to complete paperwork.",
		"payment":      "We accept cash, debit, Visa, Mastercard, and e-transfer.",
	}
	for k, v := range fallbacks {
		if len(topic) >= 3 && len(k) >= 3 {
			return v
		}
	}
	return "I don't have specific information about that topic. Please contact the clinic directly."
}

// ─── Call Logs ────────────────────────────────────────────────

func (m *Mongo) LogCall(ctx context.Context, phone *string, durationSecs *int64, status, summary string) error {
	_, err := m.db.Collection("call_logs").InsertOne(ctx, bson.M{
		"patientPhone":    phone,
		"durationSeconds": durationSecs,
		"status":          status,
		"summary":         summary,
		"startedAt":       time.Now(),
	})
	return err
}

// ─── Tool Call Audit ──────────────────────────────────────────

func (m *Mongo) LogToolCall(ctx context.Context, toolName string, args primitive.M, result, status string) error {
	_, err := m.db.Collection("tool_calls").InsertOne(ctx, bson.M{
		"toolName":  toolName,
		"arguments": args,
		"result":    result,
		"status":    status,
		"timestamp": time.Now(),
	})
	return err
}
