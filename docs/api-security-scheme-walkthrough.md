# API Security Scheme Walkthrough

This page explains how request security works in godec, why the local validator package exists, and what to do when you add or change a security scheme in [internal/api/spec.yaml](../internal/api/spec.yaml).

## 1. Purpose

The API spec is the source of truth for routes, schemas, and security requirements. At runtime, the server loads that spec, validates requests against it, and then applies API key authentication where the spec requires it.

The current concrete example is `ApiKeyAuth`, which expects an `X-API-Key` header on protected operations such as `POST /v1/media/upload-url`.

## 2. Why the local validator package exists

The repository uses a local wrapper at [internal/middleware/echovalidator](../internal/middleware/echovalidator) instead of calling `github.com/oapi-codegen/echo-middleware` directly.

That wrapper exists so the application can:

- Keep request validation and security behavior in one place.
- Plug in godec-specific authentication logic for `ApiKeyAuth`.
- Convert OpenAPI validation errors into the HTTP responses this project already uses.
- Keep the rest of the app focused on services and handlers, not OpenAPI plumbing.

In practice, the wrapper sits between Echo and the generated API handlers and calls `openapi3filter.ValidateRequest(...)` with the project’s authentication function.

## 3. Request flow

The runtime path is:

1. [internal/api/spec.yaml](../internal/api/spec.yaml) defines the OpenAPI contract, including `components.securitySchemes.ApiKeyAuth`.
2. [internal/api/documentation_handlers.go](../internal/api/documentation_handlers.go) embeds that same spec and serves it through `GetSwagger()`.
3. [cmd/api/main.go](../cmd/api/main.go) loads the embedded spec, clears `swagger.Servers` for runtime validation, and registers `echovalidator.OapiRequestValidatorWithOptions(...)`.
4. The validator middleware finds the route in the OpenAPI document and runs request validation.
5. If the operation requires `ApiKeyAuth`, the validator calls `middleware.APIKeyAuthenticator(...)`.
6. `middleware.APIKeyAuthenticator(...)` reads `X-API-Key`, validates it through the API key service adapter, and stores the validated key in the request context.
7. If validation succeeds, Echo continues to the generated strict handler and the feature handler.

For the current setup, `ApiKeyAuth` is the only security scheme wired into runtime auth.

## 4. Adding a new security scheme

When you add a new scheme to `components.securitySchemes` in `spec.yaml`, treat it as a contract change plus a runtime change.

Do this:

1. Add the new scheme definition under `components.securitySchemes`.
2. Attach that scheme to the operations that should require it.
3. Update the validator/auth glue so the runtime knows how to recognize the new scheme name.
4. Add the request extraction and validation logic for the new credential source.
5. Make sure the protected endpoints still return the intended auth error responses.
6. Regenerate API code if the spec change affects generated types or operation signatures.

If you only edit the spec and do not update runtime validation, the documentation changes but the server does not actually enforce the new scheme.

## 5. Checklist after editing `spec.yaml`

Use this checklist after changing a security scheme or a protected operation:

- Update `components.securitySchemes` with the exact scheme name used by the operation.
- Verify the operation’s `security` block points at that scheme.
- Check whether the scheme name is handled by [internal/middleware/auth.go](../internal/middleware/auth.go).
- Confirm the request header or credential source matches what the middleware reads.
- Regenerate API code if the spec change affects generated handlers or models.
- Run the API tests or start the server and confirm the protected route still validates as expected.

## 6. Troubleshooting

Common mistakes:

- **Scheme added only in the spec**: the docs update, but runtime auth does not.
- **Scheme name mismatch**: the operation says one name, but the middleware only recognizes another.
- **Header mismatch**: the spec and validator disagree on the credential header, such as `X-API-Key`.
- **Missing protected route wiring**: the operation has a `security` block, but no matching runtime auth function exists.
- **Forgetting to reload generated code**: the handler surface may lag behind the spec if code generation is needed.

If you are working on the current API key flow, the quickest sanity check is: `ApiKeyAuth` in the spec, `X-API-Key` in the request, and `middleware.APIKeyAuthenticator(...)` in the validator options.
