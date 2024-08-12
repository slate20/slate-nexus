#!/bin/bash

HOST_NAME=$1

curl \
--request POST \
--header "Authorization: Bearer $API_USER $AUTOMATION_SECRET" \
--header "Accept: application/json" \
--header "Content-Type: application/json" \
--data '{
  "host_name": "'$HOST_NAME'",
  "mode": "refresh"
}' \
"$API_URL/domain-types/service_discovery_run/actions/start/invoke"

# Wait for 10 seconds
sleep 30

curl \
--request POST \
--header "Authorization: Bearer $API_USER $AUTOMATION_SECRET" \
--header "Accept: application/json" \
--header "If-Match: *" \
--header "Content-Type: application/json" \
--data '{
    "force_foreign_changes": false,
    "redirect": false,
    "sites": [
        "main"
    ]
}' \
"$API_URL/domain-types/activation_run/actions/activate-changes/invoke"