#!/usr/bin/env bash

set -euo pipefail

next_monday=$(date -d next-monday +%Y-%m-%dT17:00:00Z)
echo $next_monday
curl -X POST \
    -H "X-TrackerToken: ${TRACKER_API_KEY}" \
    -H "Content-Type: application/json" \
    -d "{ \
    \"name\": \"$(<bump-v7-version/version)\", \
    \"story_type\": \"release\", \
    \"deadline\": \"${next_monday}\" \
}" \
"https://www.pivotaltracker.com/services/v5/projects/$TRACKER_PROJECT_ID/stories"

exit 0