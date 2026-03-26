Stage 1 — RBAC/OAuth2/JWT skeleton

Summary
- Introduces in-memory RBAC and JWT-based authentication with token refresh.
- Adds token endpoints and RBAC gates for core API.

What changed
- novabackup/security.py: added in-memory user 'charlie' for edge-case RBAC tests
- novabackup/api.py: wired /token and /token/refresh; RBAC gating on endpoints
- tests/test_api_auth.py: tests for login/refresh and protected endpoint test for 401 when no token
- tests/test_api_rbac.py: tests RBAC admin/user permissions and edge-case with no roles
- README.md: basic Stage 1 notes

Why
- Establish a secure authentication/authorization foundation before cloud integration.

How to test
- Run unit tests: pytest -q
- Run API locally: uvicorn novabackup.api:get_app --reload
- Open http://localhost:8000/docs
- Use /token to obtain tokens, then call protected endpoints
- Cloud: in Stage 1, cloud is mocked; real providers can be wired later

Next steps
- Stage 2: cloud provider integration with Mock CI + real provider stubs
- Extend RBAC scopes and rotation for tokens

Files touched
- novabackup/security.py
- novabackup/api.py
- tests/test_api_auth.py
- tests/test_api_rbac.py
- README.md

Notes
- This is MVP; tokens stored in-memory; upgrade to DB-backed for production.
