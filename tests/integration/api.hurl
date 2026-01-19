# AgentBox API Integration Tests
# Run with: hurl --test tests/integration/api.hurl --variable host=http://localhost:18080

###############################################################################
# Health Check
###############################################################################

GET {{host}}/api/v1/health
HTTP 200
[Asserts]
jsonpath "$.code" == 0
jsonpath "$.message" == "success"
jsonpath "$.data.status" == "ok"
jsonpath "$.data.version" isString


###############################################################################
# Agents API
###############################################################################

# List all agents
GET {{host}}/api/v1/agents
HTTP 200
[Asserts]
jsonpath "$.code" == 0
jsonpath "$.data" isCollection
jsonpath "$.data" count >= 3


###############################################################################
# Providers API - List
###############################################################################

# List all providers
GET {{host}}/api/v1/providers
HTTP 200
[Asserts]
jsonpath "$.code" == 0
jsonpath "$.data" isCollection
jsonpath "$.data" count >= 5

# List providers filtered by agent
GET {{host}}/api/v1/providers?agent=claude-code
HTTP 200
[Asserts]
jsonpath "$.code" == 0
jsonpath "$.data" isCollection

# List providers filtered by category
GET {{host}}/api/v1/providers?category=official
HTTP 200
[Asserts]
jsonpath "$.code" == 0
jsonpath "$.data" isCollection


###############################################################################
# Providers API - Get
###############################################################################

# Get built-in provider
GET {{host}}/api/v1/providers/anthropic
HTTP 200
[Asserts]
jsonpath "$.code" == 0
jsonpath "$.data.id" == "anthropic"
jsonpath "$.data.name" == "Anthropic"
jsonpath "$.data.is_built_in" == true

# Get non-existent provider
GET {{host}}/api/v1/providers/nonexistent-provider-id
HTTP 404
[Asserts]
jsonpath "$.code" == 404


###############################################################################
# Providers API - CRUD
###############################################################################

# Create a custom provider
POST {{host}}/api/v1/providers
Content-Type: application/json
{
  "id": "test-integration-provider",
  "name": "Integration Test Provider",
  "description": "Created by Hurl integration test",
  "agent": "claude-code",
  "category": "third_party",
  "base_url": "https://api.test.example.com"
}
HTTP 201
[Asserts]
jsonpath "$.code" == 0
jsonpath "$.data.id" == "test-integration-provider"
jsonpath "$.data.name" == "Integration Test Provider"
jsonpath "$.data.is_built_in" == false

[Captures]
provider_id: jsonpath "$.data.id"


# Get the created provider
GET {{host}}/api/v1/providers/{{provider_id}}
HTTP 200
[Asserts]
jsonpath "$.code" == 0
jsonpath "$.data.id" == "test-integration-provider"


# Update the provider
PUT {{host}}/api/v1/providers/{{provider_id}}
Content-Type: application/json
{
  "name": "Updated Test Provider",
  "description": "Updated by Hurl"
}
HTTP 200
[Asserts]
jsonpath "$.code" == 0
jsonpath "$.data.name" == "Updated Test Provider"
jsonpath "$.data.description" == "Updated by Hurl"


# Delete the provider
DELETE {{host}}/api/v1/providers/{{provider_id}}
HTTP 200
[Asserts]
jsonpath "$.code" == 0


# Verify deletion
GET {{host}}/api/v1/providers/{{provider_id}}
HTTP 404


###############################################################################
# Profiles API
###############################################################################

# List profiles
GET {{host}}/api/v1/profiles
HTTP 200
[Asserts]
jsonpath "$.code" == 0
jsonpath "$.data" isCollection


###############################################################################
# Sessions API
###############################################################################

# List sessions
GET {{host}}/api/v1/sessions
HTTP 200
[Asserts]
jsonpath "$.code" == 0
jsonpath "$.data" isCollection


###############################################################################
# Tasks API
###############################################################################

# List tasks
GET {{host}}/api/v1/tasks
HTTP 200
[Asserts]
jsonpath "$.code" == 0
jsonpath "$.data.total" >= 0


###############################################################################
# Webhooks API
###############################################################################

# List webhooks
GET {{host}}/api/v1/webhooks
HTTP 200
[Asserts]
jsonpath "$.code" == 0

# Create a webhook
POST {{host}}/api/v1/webhooks
Content-Type: application/json
{
  "url": "https://httpbin.org/post",
  "secret": "test-secret-123",
  "events": ["task.created", "task.completed"]
}
HTTP 201
[Asserts]
jsonpath "$.code" == 0
jsonpath "$.data.url" == "https://httpbin.org/post"
jsonpath "$.data.is_active" == true

[Captures]
webhook_id: jsonpath "$.data.id"


# Get the created webhook
GET {{host}}/api/v1/webhooks/{{webhook_id}}
HTTP 200
[Asserts]
jsonpath "$.code" == 0
jsonpath "$.data.id" == {{webhook_id}}


# Update the webhook
PUT {{host}}/api/v1/webhooks/{{webhook_id}}
Content-Type: application/json
{
  "url": "https://httpbin.org/anything",
  "is_active": false
}
HTTP 200
[Asserts]
jsonpath "$.code" == 0
jsonpath "$.data.url" == "https://httpbin.org/anything"
jsonpath "$.data.is_active" == false


# Delete the webhook
DELETE {{host}}/api/v1/webhooks/{{webhook_id}}
HTTP 200
[Asserts]
jsonpath "$.code" == 0


# Verify deletion
GET {{host}}/api/v1/webhooks/{{webhook_id}}
HTTP 404
