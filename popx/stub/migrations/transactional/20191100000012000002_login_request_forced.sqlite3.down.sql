INSERT INTO "_selfservice_login_requests_tmp" (id, request_url, issued_at, expires_at, active_method, csrf_token, created_at, updated_at) SELECT id, request_url, issued_at, expires_at, active_method, csrf_token, created_at, updated_at FROM "selfservice_login_requests"