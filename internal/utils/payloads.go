package utils

import (
	"fmt"
	"time"

	"github.com/google/uuid"
)

type SmallPayload struct {
	ID   int    `json:"id" form:"id"`
	Name string `json:"name" form:"name"`
}

type SmallPayloadValidate struct {
	ID   int    `json:"id" form:"id" validate:"required" binding:"required"`
	Name string `json:"name" form:"name" validate:"required" binding:"required"`
}

type NestedMetrics struct {
	CPUUsage    float64 `json:"cpuUsage"`
	MemoryBytes int64   `json:"memoryBytes"`
	Active      bool    `json:"active"`
}

type EventRecord struct {
	EventID   string        `json:"eventId"`
	Timestamp time.Time     `json:"timestamp"`
	Tags      []string      `json:"tags"`
	Metrics   NestedMetrics `json:"metrics"`
}

type LargePayload struct {
	ID          string            `json:"id"`
	UserID      int               `json:"userId"`
	AccountType string            `json:"accountType"`
	IsVerified  bool              `json:"isVerified"`
	Balance     float64           `json:"balance"`
	CreatedAt   time.Time         `json:"createdAt"`
	Events      []EventRecord     `json:"events"`
	Metadata    map[string]string `json:"metadata"`
}

// Validation counterparts
type NestedMetricsValidate struct {
	CPUUsage    float64 `json:"cpuUsage" validate:"gte=0,lte=100" binding:"gte=0,lte=100"`
	MemoryBytes int64   `json:"memoryBytes" validate:"gte=0" binding:"gte=0"`
	Active      bool    `json:"active"`
}

type EventRecordValidate struct {
	EventID   string                `json:"eventId" validate:"required,alphanum" binding:"required,alphanum"`
	Timestamp time.Time             `json:"timestamp" validate:"required" binding:"required"`
	Tags      []string              `json:"tags" validate:"omitempty,max=5,dive,min=2,max=30" binding:"omitempty,max=5,dive,min=2,max=30"`
	Metrics   NestedMetricsValidate `json:"metrics" validate:"required" binding:"required"`
}

type LargePayloadValidate struct {
	ID          string                `json:"id" validate:"required,uuid4" binding:"required,uuid4"`
	UserID      int                   `json:"userId" validate:"required,gte=1,lte=999999" binding:"required,gte=1,lte=999999"`
	AccountType string                `json:"accountType" validate:"required,oneof=premium basic enterprise" binding:"required,oneof=premium basic enterprise"`
	IsVerified  bool                  `json:"isVerified"`
	Balance     float64               `json:"balance" validate:"gte=0" binding:"gte=0"`
	CreatedAt   time.Time             `json:"createdAt" validate:"required" binding:"required"`
	Events      []EventRecordValidate `json:"events" validate:"required,min=1,max=100,dive" binding:"required,min=1,max=100,dive"`
	Metadata    map[string]string     `json:"metadata" validate:"omitempty,max=10,dive,keys,required,endkeys,required,max=255" binding:"omitempty,max=10,dive,keys,required,endkeys,required,max=255"`
}

// EventRecordGofi and LargePayloadGofi carry the constraints of
// EventRecordValidate and LargePayloadValidate in Gofi's rule syntax. Gofi
// validates nested structs through their own tags, so it has no dive or
// keys rules, and min/max on a slice or map limit its length. Per-element
// string lengths (dive,min=2,max=30 on Tags) have no Gofi equivalent.
type EventRecordGofi struct {
	EventID   string                `json:"eventId" validate:"required,alphanum"`
	Timestamp time.Time             `json:"timestamp" validate:"required"`
	Tags      []string              `json:"tags" validate:"omitempty,max=5"`
	Metrics   NestedMetricsValidate `json:"metrics" validate:"required"`
}

type LargePayloadGofi struct {
	ID          string            `json:"id" validate:"required,uuid4"`
	UserID      int               `json:"userId" validate:"required,gte=1,lte=999999"`
	AccountType string            `json:"accountType" validate:"required,oneof=premium basic enterprise"`
	IsVerified  bool              `json:"isVerified"`
	Balance     float64           `json:"balance" validate:"gte=0"`
	CreatedAt   time.Time         `json:"createdAt" validate:"required"`
	Events      []EventRecordGofi `json:"events" validate:"required,min=1,max=100"`
	Metadata    map[string]string `json:"metadata" validate:"omitempty,max=10"`
}

// ----- Generators -----

func GenerateLargeData() []LargePayload {
	now := time.Now().UTC()
	largeData := make([]LargePayload, 50)
	for i := 0; i < 50; i++ {
		largeData[i] = LargePayload{
			ID:          fmt.Sprintf("uuid-%d", i),
			UserID:      i,
			AccountType: "premium",
			IsVerified:  i%2 == 0,
			Balance:     1500.50 + float64(i),
			CreatedAt:   now,
			Metadata:    map[string]string{"source": "web", "version": "1.0"},
			Events: []EventRecord{
				{
					EventID:   fmt.Sprintf("evt-%d-1", i),
					Timestamp: now,
					Tags:      []string{"login", "success"},
					Metrics:   NestedMetrics{CPUUsage: 45.2, MemoryBytes: 1024000, Active: true},
				},
				{
					EventID:   fmt.Sprintf("evt-%d-2", i),
					Timestamp: now,
					Tags:      []string{"view", "page"},
					Metrics:   NestedMetrics{CPUUsage: 55.1, MemoryBytes: 2048000, Active: false},
				},
			},
		}
	}
	return largeData
}

func GenerateLargeDataValidate() []LargePayloadValidate {
	now := time.Now().UTC()
	largeDataValidate := make([]LargePayloadValidate, 50)
	for i := 0; i < 50; i++ {
		largeDataValidate[i] = LargePayloadValidate{
			ID:          uuid.NewString(),
			UserID:      i + 1,
			AccountType: "premium",
			IsVerified:  i%2 == 0,
			Balance:     1500.50 + float64(i),
			CreatedAt:   now,
			Metadata:    map[string]string{"source": "web", "version": "1.0"},
			Events: []EventRecordValidate{
				{
					EventID:   fmt.Sprintf("evt%d1", i),
					Timestamp: now,
					Tags:      []string{"login", "success", "choir"},
					Metrics:   NestedMetricsValidate{CPUUsage: 45.2, MemoryBytes: 1024000, Active: true},
				},
				{
					EventID:   fmt.Sprintf("evt%d2", i),
					Timestamp: now,
					Tags:      []string{"view", "page", "stuff"},
					Metrics:   NestedMetricsValidate{CPUUsage: 55.1, MemoryBytes: 2048000, Active: false},
				},
			},
		}
	}
	return largeDataValidate
}

// GenerateLargeDataGofi returns GenerateLargeDataValidate's data as Gofi types.
func GenerateLargeDataGofi() []LargePayloadGofi {
	src := GenerateLargeDataValidate()
	out := make([]LargePayloadGofi, len(src))
	for i, p := range src {
		events := make([]EventRecordGofi, len(p.Events))
		for j, e := range p.Events {
			events[j] = EventRecordGofi(e)
		}
		out[i] = LargePayloadGofi{
			ID:          p.ID,
			UserID:      p.UserID,
			AccountType: p.AccountType,
			IsVerified:  p.IsVerified,
			Balance:     p.Balance,
			CreatedAt:   p.CreatedAt,
			Events:      events,
			Metadata:    p.Metadata,
		}
	}
	return out
}
