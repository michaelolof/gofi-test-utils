import re
import os

def update_file(p, is_gofi=False, is_gin=False):
    with open(p, 'r') as f:
        content = f.read()

    # Add the internal/utils import
    if "github.com/michaelolof/gofi-test-utils/internal/utils" not in content:
        content = content.replace('"os"\n', '"os"\n\t"github.com/michaelolof/gofi-test-utils/internal/utils"\n')

    # Re-map all the slice struct types to utils.*
    content = content.replace('var p []SmallPayload\n', 'var p []utils.SmallPayload\n')
    content = content.replace('var p []LargePayload\n', 'var p []utils.LargePayload\n')
    content = content.replace('var p []SmallPayloadValidate\n', 'var p []utils.SmallPayloadValidate\n')
    content = content.replace('var p []LargePayloadValidate\n', 'var p []utils.LargePayloadValidate\n')

    # Remove the struct blocks.
    # Target everything from type SmallPayload down to the end of LargePayloadValidate
    struct_block_re = re.compile(r'type SmallPayload struct \{.*?type LargePayloadValidate struct \{.*?\n\}\n', re.DOTALL)
    content = struct_block_re.sub('', content)

    # For Gofi, we must map the generic Schema structs to use utils.* 
    if is_gofi:
        content = content.replace('Body []LargePayload\n', 'Body []utils.LargePayload\n')
        content = content.replace('Body []LargePayloadValidate\n', 'Body []utils.LargePayloadValidate\n')
        content = content.replace('Body []SmallPayload\n', 'Body []utils.SmallPayload\n')
        content = content.replace('Body []SmallPayloadValidate\n', 'Body []utils.SmallPayloadValidate\n')

    # Replace the complex hard-coded array generation blocks with the utility functions
    # They usually start with `now := time.Now().UTC()` and go down to `}` closing the array loops
    init_pattern = re.compile(r'\s*now := time\.Now\(\)\.UTC\(\).*?\n\s+largeDataValidate\[i\] = LargePayloadValidate\{.*?\n\s+\}\n\s+\}', re.DOTALL)
    
    # In some older iterations, there might be small payload mock init blocks too (`smallData := make...`)
    # The new utility removes the need to mock arrays manually here, we just call GenerateLargeData()
    gen_block = '\n\tlargeData := utils.GenerateLargeData()\n\tlargeDataValidate := utils.GenerateLargeDataValidate()'
    content = init_pattern.sub(gen_block, content)

    with open(p, 'w') as f:
        f.write(content)

base = './cmd/httpbench'
update_file(base + '/gofi/main.go', is_gofi=True)
update_file(base + '/chi/main.go')
update_file(base + '/echo/main.go')
update_file(base + '/gin/main.go', is_gin=True)
update_file(base + '/fiber/main.go')

print("Successfully refactored main.go files to use the utils package.")
