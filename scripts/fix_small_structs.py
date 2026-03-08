import os

def fix_file(p):
    with open(p, 'r') as f:
        content = f.read()

    # Generic Schema Replacements
    content = content.replace('Body SmallPayloadValidate', 'Body utils.SmallPayloadValidate')
    content = content.replace('FormData SmallPayloadValidate', 'FormData utils.SmallPayloadValidate')
    content = content.replace('Multipart SmallPayloadValidate', 'Multipart utils.SmallPayloadValidate')
    content = content.replace('Body SmallPayload', 'Body utils.SmallPayload')
    
    # Slice mapping fallback if we missed it
    content = content.replace('make([]SmallPayload, 100)', 'make([]utils.SmallPayload, 100)')
    content = content.replace('data[i] = SmallPayload{ID: i, Name:', 'data[i] = utils.SmallPayload{ID: i, Name:')
    content = content.replace('SmallPayloadValidate{ID: 1, Name: "test"}', 'utils.SmallPayloadValidate{ID: 1, Name: "test"}')

    # Chi/Echo/Fiber/Gin specific parameter type replacements
    content = content.replace('p SmallPayloadValidate', 'p utils.SmallPayloadValidate')
    content = content.replace('p SmallPayload', 'p utils.SmallPayload')
    
    with open(p, 'w') as f:
        f.write(content)

base = './cmd/httpbench'
fix_file(base + '/gofi/main.go')
fix_file(base + '/chi/main.go')
fix_file(base + '/echo/main.go')
fix_file(base + '/gin/main.go')
fix_file(base + '/fiber/main.go')

print("Fixed SmallPayloadValidate mappings.")
