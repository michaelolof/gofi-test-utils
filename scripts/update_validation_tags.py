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
	CPUUsage    float64 `json:"cpuUsage" validate:"gte=0,lte=100"`
	MemoryBytes int64   `json:"memoryBytes" validate:"gte=0"`
	Active      bool    `json:"active"`
}

type EventRecordValidate struct {
	EventID   string                `json:"eventId" validate:"required,alphanum"`
	Timestamp time.Time             `json:"timestamp" validate:"required"`
	Tags      []string              `json:"tags" validate:"omitempty,max=5,dive,min=3,max=30"`
	Metrics   NestedMetricsValidate `json:"metrics" validate:"required"`
}

type LargePayloadValidate struct {
	ID          string                `json:"id" validate:"required,uuid4"`
	UserID      int                   `json:"userId" validate:"required,gte=1,lte=999999"`
	AccountType string                `json:"accountType" validate:"required,oneof=premium basic enterprise"`
	IsVerified  bool                  `json:"isVerified"`
	Balance     float64               `json:"balance" validate:"gte=0"`
	CreatedAt   time.Time             `json:"createdAt" validate:"required"`
	Events      []EventRecordValidate `json:"events" validate:"required,min=1,max=100,dive"`
	Metadata    map[string]string     `json:"metadata" validate:"omitempty,max=10,dive,keys,required,endkeys,required,max=255"`
}
"""

GIN_STRUCTS_VALIDATE = STRUCTS_VALIDATE.replace('validate:', 'binding:')

def update_file(p, is_gin=False):
    with open(p, 'r') as f:
        content = f.read()

    # Identify where the current old validation block is and replace it.
    # To do this safely, we will substring replace the specific blocks we injected earlier.
    
    # regex match to find the old type NestedMetricsValidate block up to Metadata string }
    old_valid_block_re = re.compile(r'type NestedMetricsValidate struct \{.*?type LargePayloadValidate struct \{.*?Metadata\s+map\[string\]string\s+`json:"metadata" validate:"omitempty"`\n\}', re.DOTALL)
    old_valid_block_gin_re = re.compile(r'type NestedMetricsValidate struct \{.*?type LargePayloadValidate struct \{.*?Metadata\s+map\[string\]string\s+`json:"metadata" binding:"omitempty"`\n\}', re.DOTALL)
    
    if is_gin:
        content = old_valid_block_gin_re.sub(GIN_STRUCTS_VALIDATE.strip(), content)
    else:
        content = old_valid_block_re.sub(STRUCTS_VALIDATE.strip(), content)

    with open(p, 'w') as f:
        f.write(content)

base = './cmd/httpbench'
update_file(base + '/gofi/main.go')
update_file(base + '/chi/main.go')
update_file(base + '/echo/main.go')
update_file(base + '/gin/main.go', is_gin=True)
update_file(base + '/fiber/main.go')
print("Done applying complex struct tags.")
