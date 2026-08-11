#!/usr/bin/env bash
#
# Integration tests for the user domain. Runs on its own or as part of the suite
# driven by api_test.sh.
#
# Usage:
#   zarf/integration/userapp_test.sh [base-url]

source "$(dirname "${BASH_SOURCE[0]}")/lib.sh"
preflight

# -----------------------------------------------------------------------------
# GET /users

describe 'GET /users returns the user collection'
if request GET /users; then
	assert_eq 'status code' "${RESP_CODE}" '200'
	assert_eq 'content type' "${RESP_TYPE}" 'application/json'
	assert_eq 'body is a json array' "$(jqv 'type')" 'array'

	count="$(jqv 'length')"
	if [[ -n "${count}" ]]; then
		pass "returned ${count} user(s)"
	fi

	if [[ "${count:-0}" -gt 0 ]]; then
		# Every element must carry the full app layer representation with the
		# documented primitive types.
		shape="$(jqv '
			map(
				(.id          | type == "number") and
				(.name        | type == "string") and
				(.email       | type == "string") and
				(.enabled     | type == "boolean") and
				(.dateCreated | type == "string") and
				(.dateUpdated | type == "string")
			) | all')"
		assert_eq 'every user has the expected fields and types' "${shape}" 'true'

		rfc3339='^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}([.][0-9]+)?(Z|[+-][0-9]{2}:[0-9]{2})$'
		assert_eq 'dates are RFC3339' \
			"$(jqv "map((.dateCreated | test(\"${rfc3339}\")) and (.dateUpdated | test(\"${rfc3339}\"))) | all")" \
			'true'

		assert_eq 'emails are non-empty' "$(jqv 'map(.email | length > 0) | all')" 'true'
	else
		skip 'no users in the database, field assertions skipped' \
			'seed some users to exercise the full payload'
	fi
fi

describe '/users rejects an unsupported method'
if request PATCH /users; then
	assert_eq 'status code' "${RESP_CODE}" '405'
fi

# -----------------------------------------------------------------------------
# User lifecycle: POST -> GET -> PUT -> DELETE
#
# Every created user carries a run specific suffix so the suite can be run
# repeatedly against the same database, and anything it creates is removed on
# exit even when an assertion fails part way through.

SUFFIX="$(suite_id)"
readonly SUFFIX

EMAIL="dave.${SUFFIX}@example.com"
readonly EMAIL

# created_ids collects every user id the suite creates so cleanup can remove
# them regardless of which test created them.
created_ids=()

cleanup_users() {
	local id
	for id in ${created_ids[@]+"${created_ids[@]}"}; do
		curl --silent --show-error --max-time 10 --output /dev/null \
			--request DELETE "${BASE_URL}/users/${id}" 2>/dev/null || true
	done
}
on_exit cleanup_users

USER_ID=''

describe 'POST /users creates a user'
if request POST /users \
	--header 'Content-Type: application/json' \
	--data "{\"name\":\"Dave Tester\",\"email\":\"${EMAIL}\"}"; then

	assert_eq 'status code' "${RESP_CODE}" '201'
	assert_eq 'content type' "${RESP_TYPE}" 'application/json'
	assert_eq 'name' "$(jqv '.name')" 'Dave Tester'
	assert_eq 'email' "$(jqv '.email')" "${EMAIL}"
	assert_eq 'enabled defaults to true' "$(jqv '.enabled')" 'true'

	assert_eq 'response has the expected fields and types' "$(jqv '
		(.id          | type == "number") and
		(.name        | type == "string") and
		(.email       | type == "string") and
		(.enabled     | type == "boolean") and
		(.dateCreated | type == "string") and
		(.dateUpdated | type == "string")')" 'true'

	USER_ID="$(jqv '.id')"
	if [[ -n "${USER_ID}" ]]; then
		created_ids+=("${USER_ID}")
		pass "created user id ${USER_ID}"
	else
		fail 'response carries an id'
	fi
fi

describe 'POST /users rejects a malformed json body'
if request POST /users \
	--header 'Content-Type: application/json' \
	--data '{"name":"Dave Tester",'; then

	assert_eq 'status code' "${RESP_CODE}" '400'
fi

describe 'POST /users rejects an invalid name'
if request POST /users \
	--header 'Content-Type: application/json' \
	--data "{\"name\":\"a\",\"email\":\"valid.${SUFFIX}@example.com\"}"; then

	assert_eq 'status code' "${RESP_CODE}" '400'
	assert_eq 'body is a json array' "$(jqv 'type')" 'array'
	assert_eq 'field error reported for name' "$(jqv 'map(.field) | index("name") != null')" 'true'
	assert_eq 'every entry has field and error' \
		"$(jqv 'map((.field | type == "string") and (.error | type == "string")) | all')" 'true'
fi

describe 'POST /users rejects an invalid email'
if request POST /users \
	--header 'Content-Type: application/json' \
	--data '{"name":"Dave Tester","email":"not-an-email"}'; then

	assert_eq 'status code' "${RESP_CODE}" '400'
	assert_eq 'body is a json array' "$(jqv 'type')" 'array'
	assert_eq 'field error reported for email' "$(jqv 'map(.field) | index("email") != null')" 'true'
fi

describe 'POST /users reports every invalid field at once'
if request POST /users \
	--header 'Content-Type: application/json' \
	--data '{"name":"a","email":"not-an-email"}'; then

	assert_eq 'status code' "${RESP_CODE}" '400'
	assert_eq 'body is a json array' "$(jqv 'type')" 'array'
	assert_eq 'both fields reported' \
		"$(jqv 'map(.field) | (index("name") != null) and (index("email") != null)')" 'true'
fi

describe 'POST /users rejects a duplicate email'
if request POST /users \
	--header 'Content-Type: application/json' \
	--data "{\"name\":\"Other Tester\",\"email\":\"${EMAIL}\"}"; then

	assert_eq 'status code' "${RESP_CODE}" '409'
fi

describe 'GET /users/{user_id} returns the created user'
if [[ -z "${USER_ID}" ]]; then
	skip 'no user was created, lookup skipped'
elif request GET "/users/${USER_ID}"; then
	assert_eq 'status code' "${RESP_CODE}" '200'
	assert_eq 'content type' "${RESP_TYPE}" 'application/json'
	assert_eq 'body is a json object' "$(jqv 'type')" 'object'
	assert_eq 'id' "$(jqv '.id')" "${USER_ID}"
	assert_eq 'name' "$(jqv '.name')" 'Dave Tester'
	assert_eq 'email' "$(jqv '.email')" "${EMAIL}"
fi

describe 'PUT /users/{user_id} applies a partial update'
if [[ -z "${USER_ID}" ]]; then
	skip 'no user was created, update skipped'
elif request PUT "/users/${USER_ID}" \
	--header 'Content-Type: application/json' \
	--data '{"name":"Dave Updated"}'; then

	assert_eq 'status code' "${RESP_CODE}" '200'
	assert_eq 'content type' "${RESP_TYPE}" 'application/json'
	assert_eq 'id is unchanged' "$(jqv '.id')" "${USER_ID}"
	assert_eq 'name is updated' "$(jqv '.name')" 'Dave Updated'
	assert_eq 'email is preserved' "$(jqv '.email')" "${EMAIL}"
	assert_eq 'enabled is preserved' "$(jqv '.enabled')" 'true'
fi

describe 'GET /users/{user_id} reflects the update'
if [[ -z "${USER_ID}" ]]; then
	skip 'no user was created, re-read skipped'
elif request GET "/users/${USER_ID}"; then
	assert_eq 'status code' "${RESP_CODE}" '200'
	assert_eq 'name' "$(jqv '.name')" 'Dave Updated'
	assert_eq 'email' "$(jqv '.email')" "${EMAIL}"
fi

describe 'PUT /users/{user_id} rejects an invalid email'
if [[ -z "${USER_ID}" ]]; then
	skip 'no user was created, validation skipped'
elif request PUT "/users/${USER_ID}" \
	--header 'Content-Type: application/json' \
	--data '{"email":"not-an-email"}'; then

	assert_eq 'status code' "${RESP_CODE}" '400'
	assert_eq 'body is a json array' "$(jqv 'type')" 'array'
	assert_eq 'field error reported for email' "$(jqv 'map(.field) | index("email") != null')" 'true'
fi

describe 'PUT /users/{user_id} rejects a malformed json body'
if [[ -z "${USER_ID}" ]]; then
	skip 'no user was created, validation skipped'
elif request PUT "/users/${USER_ID}" \
	--header 'Content-Type: application/json' \
	--data '{"name":'; then

	assert_eq 'status code' "${RESP_CODE}" '400'
fi

describe 'DELETE /users/{user_id} removes the user'
if [[ -z "${USER_ID}" ]]; then
	skip 'no user was created, delete skipped'
elif request DELETE "/users/${USER_ID}"; then
	assert_eq 'status code' "${RESP_CODE}" '204'
	assert_eq 'body is empty' "${RESP_BODY}" ''

	if request GET "/users/${USER_ID}"; then
		assert_eq 'status code after delete' "${RESP_CODE}" '404'
	fi
fi

# -----------------------------------------------------------------------------
# User id handling

describe 'GET /users/{user_id} returns not found for an unknown id'
if request GET /users/999999999; then
	assert_eq 'status code' "${RESP_CODE}" '404'
fi

describe 'PUT /users/{user_id} returns not found for an unknown id'
if request PUT /users/999999999 \
	--header 'Content-Type: application/json' \
	--data '{"name":"Ghost User"}'; then

	assert_eq 'status code' "${RESP_CODE}" '404'
fi

describe 'DELETE /users/{user_id} returns not found for an unknown id'
if request DELETE /users/999999999; then
	assert_eq 'status code' "${RESP_CODE}" '404'
fi

describe 'GET /users/{user_id} rejects a non numeric id'
if request GET /users/abc; then
	assert_eq 'status code' "${RESP_CODE}" '400'
fi

describe 'PUT /users/{user_id} rejects a non numeric id'
if request PUT /users/abc \
	--header 'Content-Type: application/json' \
	--data '{"name":"Dave Tester"}'; then

	assert_eq 'status code' "${RESP_CODE}" '400'
fi

describe 'DELETE /users/{user_id} rejects a non numeric id'
if request DELETE /users/abc; then
	assert_eq 'status code' "${RESP_CODE}" '400'
fi
