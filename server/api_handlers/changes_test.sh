SITE_NAME="main"
API_URL="http://localhost:5000/main/check_mk/api/1.0"

USERNAME="cmkadmin"
PASSWORD="slatermmdev"

curl \
--request POST \
--header "Authorization: Bearer $USERNAME $PASSWORD" \
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