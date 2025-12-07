package suites

import (
	"time"

	"github.com/example/go-test-framework/framework/declarative"
	"github.com/example/go-test-framework/framework/env"
	"github.com/example/go-test-framework/framework/suite"
)

func init() {
	suite.RegisterSuite(&suite.TestSuite{
		ID:            "deposit-flow-health-e2e",
		Name:          "Complete Deposit Flow with Bonus Activation",
		Services:      []string{"payments-service"},
		Dependencies:  []string{"postgres", "mongodb", "redis", "rabbitmq"},
		ExecutionType: suite.ExecutionTypeSequential,
		Timeout:       10 * time.Minute,
		Retries:       2,
		Tests: []suite.TestDefinition{
			{Service: "payments-service", Type: "integration"},
		},
		DeclarativeTests: []declarative.DeclarativeTest{
			{
				Name:        "Create bonus and verify bonus_wallet creation",
				Description: "Создать бонус через API и проверить что создался bonus_wallet в PostgreSQL",
				Action: declarative.Action{
					Service:  "bonus-service",
					Endpoint: "/api/bonuses/create",
					Method:   "POST",
					Body: map[string]any{
						"userId":   1,
						"amount":   100,
						"type":     "promo",
						"currency": "USD",
					},
					Extract: map[string]string{
						"bonusId":     "id",
						"bonusAmount": "amount",
						"bonusType":   "type",
					},
				},
				ResponseAssertions: &declarative.ResponseAssertions{
					Status: 201,
					Body: &declarative.BodyAssertions{
						Contains: map[string]any{
							"type":   "promo",
							"amount": 100,
						},
					},
				},
				DelayAfter: 500 * time.Millisecond,
				Assertions: []declarative.Assertion{
					{
						Database:     "mongodb",
						DatabaseName: "payments",
						Collection:   "bonuses",
						Query: map[string]any{
							"_id":    "${bonusId}",
							"userId": 1,
						},
						Expected: declarative.ExpectedResult{
							Count: ptr(1),
							Contains: map[string]any{
								"amount": 100,
								"type":   "promo",
								"status": "active",
							},
						},
					},
					{
						Database: "postgres",
						Schema:   "admin_db",
						Table:    "bonus_wallets",
						Query: map[string]any{
							"user_id":  1,
							"bonus_id": "${bonusId}",
						},
						Expected: declarative.ExpectedResult{
							Count: ptr(1),
							Contains: map[string]any{
								"currency":       "USD",
								"status":         "active",
								"initial_amount": 100,
							},
						},
					},
				},
			},
		},
		Environment: env.EnvironmentConfig{
			Postgres: &env.PostgresConfig{
				Version: "14",
				Memory:  "1Gi",
				Databases: []env.PostgresDBEntry{
					{Name: "admin_db", Username: "admin_user", Password: "qwertziboi"},
				},
			},
			Mongo: &env.MongoConfig{
				Version: "6.0",
				Memory:  "512Mi",
				Databases: []env.MongoDBEntry{
					{Name: "payments", Username: "admin", Password: "password"},
				},
			},
			Redis:    &env.RedisConfig{Version: "7", Memory: "256Mi"},
			RabbitMQ: &env.RabbitConfig{Version: "3.11", Memory: "512Mi"},
		},
		Config: map[string]any{"targetEnvironment": "staging"},
	})
}

func ptr[T any](v T) *T {
	return &v
}
