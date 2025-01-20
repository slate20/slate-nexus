#!/bin/bash

source .env
AUTH_URL=https://auth.${NEXUS_FQDN}/api/v3

# Get the default-authentication-flow and default-invalidation-flow pks from the response
response=$(curl -X GET "$AUTH_URL/flows/instances/" \
    -H "Authorization: Bearer ${AK_BT_TOKEN}" \
    -H "Content-Type: application/json")

# extract auth_flow_pk where slug is default-authentication-flow
auth_flow_pk=$(echo "$response" | jq -r '.results[] | select(.slug == "default-authentication-flow") | .pk')

# extract invalidation_flow_pk where slug is default-invalidation-flow
invalidation_flow_pk=$(echo "$response" | jq -r '.results[] | select(.slug == "default-invalidation-flow") | .pk')

echo "Auth flow pk: $auth_flow_pk"
echo "Invalidation flow pk: $invalidation_flow_pk"

echo "Setting up Nexus provider"
response=$(curl -X POST "$AUTH_URL/providers/proxy/" \
    -H "Authorization: Bearer ${AK_BT_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "Provider for Nexus",
        "authorization_flow": "'${auth_flow_pk}'",
        "invalidation_flow": "'${invalidation_flow_pk}'",
        "external_host": "https://auth.'${NEXUS_FQDN}'",
        "internal_host_ssl_validation": true,
        "mode": "forward_domain",
        "intercept_header_auth": true,
        "cookie_domain": "'${NEXUS_FQDN}'",
        "access_token_validity": "hours=8"
    }')

# DEBUG: Print the response
echo "$response"

# Extract the provider pk from the response
provider_pk=$(echo "$response" | jq -r '.pk')

echo "Provider pk: $provider_pk"

echo "Setting up the Nexus Application"
curl -X POST "$AUTH_URL/core/applications/" \
    -H "Authorization: Bearer ${AK_BT_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "Nexus",
        "slug": "nexus",
        "provider": "'${provider_pk}'",
        "open_in_new_tab": true,
        "meta_launch_url": "https://'${NEXUS_FQDN}'",
        "policy_engine_mode": "any"
    }'

echo "Getting the pk of the authentik Embedded Outpost"
response=$(curl -X GET "$AUTH_URL/outposts/instances/" \
    -H "Authorization: Bearer ${AK_BT_TOKEN}" \
    -H "Content-Type: application/json")

# Extract the first result from the results array
result=$(echo "$response" | jq -r '.results[0]')

echo "Result: $result"

# Extract the pk from the result
outpost_pk=$(echo "$result" | jq -r '.pk')

echo "Outpost pk: $outpost_pk"

# Construct the PUT URL
put_url="$AUTH_URL/outposts/instances/$outpost_pk/"

echo "PUT url: $put_url"

# Parse the response and modify the providers array
updated_body=$(echo "$result" | jq --argjson new_providers "$provider_pk" '
    .providers = [$new_providers] | 
    {name: .name, type: .type, providers: .providers, config: .config}')

echo "Updated outpost: $updated_body"

# Make the PUT request
curl -X PUT "$put_url" \
    -H "Authorization: Bearer ${AK_BT_TOKEN}" \
    -H "Content-Type: application/json" \
    -d "$updated_body"

echo "Updated outpost with the Nexus provider"

# Set the default flow title
curl -X PUT "$AUTH_URL/flows/instances/default-authentication-flow/" \
    -H "Authorization: Bearer ${AK_BT_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{
        "name": "Welcome to Slate Nexus!",
        "slug": "default-authentication-flow",
        "title": "Welcome to Slate Nexus!",
        "designation": "authentication"
    }'

echo "Set default flow title"

# Set Branding
curl -X POST "$AUTH_URL/core/brands/" \
    -H "Authorization: Bearer ${AK_BT_TOKEN}" \
    -H "Content-Type: application/json" \
    -d '{
        "domain": "auth.'${NEXUS_FQDN}'",
        "default": false,
        "branding_title": "Slate Nexus",
        "branding_logo": "/media/logo.png",
        "branding_favicon": "/media/logo.png"
    }'

echo "Set branding"