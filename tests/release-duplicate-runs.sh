#!/usr/bin/env bash
set -euo pipefail
# Simulate duplicate push runs for a single commit; the most recent attempt
# must govern the gate and cannot be bypassed by a previous successful run.
runs='[
{"id":101,"name":"Build and test Windows installer","status":"completed","conclusion":"success","created_at":"2026-10-26T10:00:00Z","run_number":20},
{"id":102,"name":"Build and test Windows installer","status":"in_progress","conclusion":null,"created_at":"2026-10-26T10:01:00Z","run_number":21}
]'
latest="$(printf '%s' "$runs" | jq -c 'sort_by(.created_at, .run_number, .id) | last // null')"
test "$(printf '%s' "$latest" | jq -r '.id')" = 102
test "$(printf '%s' "$latest" | jq -r '.conclusion // "pending"')" = pending
fail="$(printf '%s' "$runs" | jq -c '.[1].status="completed" | .[1].conclusion="failure" | sort_by(.created_at, .run_number, .id) | last')"
test "$(printf '%s' "$fail" | jq -r '.conclusion')" = failure
pass="$(printf '%s' "$runs" | jq -c '.[1].status="completed" | .[1].conclusion="success" | sort_by(.created_at, .run_number, .id) | last')"
test "$(printf '%s' "$pass" | jq -r '.id')" = 102
test "$(printf '%s' "$pass" | jq -r '.conclusion')" = success
echo 'PASS: duplicated runs use newest; latest pending/failure blocks, latest success allows.'
