import os
import re

STRUCTS = """
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
"""

STRUCTS_VALIDATE = """
type NestedMetricsValidate struct {
	CPUUsage    float64 `json:"cpuUsage" validate:"gte=0"`
	MemoryBytes int64   `json:"memoryBytes" validate:"gte=0"`
	Active      bool    `json:"active"`
}

type EventRecordValidate struct {
	EventID   string                `json:"eventId" validate:"required"`
	Timestamp time.Time             `json:"timestamp" validate:"required"`
	Tags      []string              `json:"tags" validate:"omitempty"`
	Metrics   NestedMetricsValidate `json:"metrics" validate:"required"`
}

type LargePayloadValidate struct {
	ID          string                `json:"id" validate:"required"`
	UserID      int                   `json:"userId" validate:"required,gte=0"`
	AccountType string                `json:"accountType" validate:"required"`
	IsVerified  bool                  `json:"isVerified"`
	Balance     float64               `json:"balance" validate:"gte=0"`
	CreatedAt   time.Time             `json:"createdAt" validate:"required"`
	Events      []EventRecordValidate `json:"events" validate:"required,dive"`
	Metadata    map[string]string     `json:"metadata" validate:"omitempty"`
}
"""

GIN_STRUCTS_VALIDATE = STRUCTS_VALIDATE.replace('validate:', 'binding:')

GoCodeLargeInit = """
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

	largeDataValidate := make([]LargePayloadValidate, 50)
	for i := 0; i < 50; i++ {
		largeDataValidate[i] = LargePayloadValidate{
			ID:          fmt.Sprintf("uuid-%d", i),
			UserID:      i,
			AccountType: "premium",
			IsVerified:  i%2 == 0,
			Balance:     1500.50 + float64(i),
			CreatedAt:   now,
			Metadata:    map[string]string{"source": "web", "version": "1.0"},
			Events: []EventRecordValidate{
				{
					EventID:   fmt.Sprintf("evt-%d-1", i),
					Timestamp: now,
					Tags:      []string{"login", "success"},
					Metrics:   NestedMetricsValidate{CPUUsage: 45.2, MemoryBytes: 1024000, Active: true},
				},
				{
					EventID:   fmt.Sprintf("evt-%d-2", i),
					Timestamp: now,
					Tags:      []string{"view", "page"},
					Metrics:   NestedMetricsValidate{CPUUsage: 55.1, MemoryBytes: 2048000, Active: false},
				},
			},
		}
	}
"""

def update_file(p, is_gin=False, is_gofi=False):
    with open(p, 'r') as f:
        content = f.read()

    # Make sure time is imported
    if '"time"' not in content:
        content = content.replace('"os"\n', '"os"\n\t"time"\n')

    # Replace old LargePayloadSchema
    if is_gofi:
        content = re.sub(r'type LargePayloadSchema struct \{.*?\}\n}',
            'type LargePayloadSchema struct {\n\tRequest struct {\n\t\tBody []LargePayload\n\t}\n\tOk struct {\n\t\tBody string\n\t}\n\tErr struct {\n\t\tBody string\n\t}\n}',
            content, flags=re.DOTALL)
        content = re.sub(r'type LargeValidateSchema struct \{.*?\}\n}',
            'type LargeValidateSchema struct {\n\tRequest struct {\n\t\tBody []LargePayloadValidate\n\t}\n\tOk struct {\n\t\tBody string\n\t}\n\tErr struct {\n\t\tBody string\n\t}\n}',
            content, flags=re.DOTALL)

    # Inject the structs
    if 'type LargePayload struct' not in content:
        if is_gofi:
            insert_point = 'type SmallValidateSchema struct {'
            valid_structs = STRUCTS_VALIDATE
        elif is_gin:
            insert_point = 'func main() {'
            valid_structs = GIN_STRUCTS_VALIDATE
        else:
            insert_point = 'func main() {'
            valid_structs = STRUCTS_VALIDATE
            
        content = content.replace(insert_point, f"{STRUCTS}\n{valid_structs}\n{insert_point}")

    # Replace the initialization lines in main
    init_pattern = r'largeData := make\(\[\]SmallPayload, 1000\).*?largeDataValidate\[i\] = SmallPayloadValidate\{ID: i, Name: fmt\.Sprintf\("item %d", i\)\}\n\t\}'
    content = re.sub(init_pattern, GoCodeLargeInit.strip(), content, flags=re.DOTALL)
    
    content = content.replace('var p []SmallPayload\n', 'var p []LargePayload\n')
    content = content.replace('var p []SmallPayloadValidate\n', 'var p []LargePayloadValidate\n')

    with open(p, 'w') as f:
        f.write(content)

base = './cmd/httpbench'
update_file(base + '/gofi/main.go', is_gofi=True)
update_file(base + '/chi/main.go')
update_file(base + '/echo/main.go')
update_file(base + '/gin/main.go', is_gin=True)
update_file(base + '/fiber/main.go')
print("Done rewriting payloads in go main files.")
