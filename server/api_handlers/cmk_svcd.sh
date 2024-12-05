#!/bin/bash

HOST_NAME=$1
# API_USER=cmkadmin
# AUTOMATION_SECRET=slatermmdev
# SITE_NAME=main
# API_URL=http://localhost:5000/main/check_mk/api/1.0

curl \
--request POST \
--header "Authorization: Bearer $API_USER $AUTOMATION_SECRET" \
--header "Accept: application/json" \
--header "Content-Type: application/json" \
--data '{
  "host_name": "'$HOST_NAME'",
  "mode": "tabula_rasa"
}' \
"$API_URL/domain-types/service_discovery_run/actions/start/invoke"

# Wait for 10 seconds
sleep 10

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