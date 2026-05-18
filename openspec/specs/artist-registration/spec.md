## ADDED Requirements

### Requirement: Artist can register a professional account
The system SHALL allow a professional artist to self-register by providing identity and professional profile data. A successful registration SHALL atomically create a user identity, an artist profile, and an active artist role assignment in a single database transaction.

#### Scenario: Successful registration
- **WHEN** a POST request is made to `/api/artists/register` with all required fields valid
- **THEN** the system returns HTTP 201 with a response envelope containing the created artist's `id` (UUID), `fullname`, `email`, `phone`, `business_name`, `primary_service`, `city`, and `instagram`

#### Scenario: Password is never returned
- **WHEN** registration succeeds
- **THEN** the response body SHALL NOT contain the `password` field in any form

### Requirement: Registration input is validated
The system SHALL validate all required fields before any database write. The first failing field SHALL be returned immediately with a field-specific error code.

#### Scenario: Missing required field
- **WHEN** any of `fullname`, `email`, `phone`, `password`, `business_name`, `primary_service`, or `city` is absent or empty
- **THEN** the system returns HTTP 422 with `errors[].field` set to the offending field name and `errors[].code` set to `REQUIRED`

#### Scenario: Invalid email format
- **WHEN** `email` is present but not a valid email address
- **THEN** the system returns HTTP 422 with `errors[].field = "email"` and `errors[].code = "INVALID_EMAIL"`

#### Scenario: Password too short
- **WHEN** `password` is present but fewer than 8 characters
- **THEN** the system returns HTTP 422 with `errors[].field = "password"` and `errors[].code = "TOO_SHORT"`

#### Scenario: Invalid primary service
- **WHEN** `primary_service` is present but not one of `makeup`, `hair`, or `attire`
- **THEN** the system returns HTTP 422 with `errors[].field = "primary_service"` and `errors[].code = "INVALID_SERVICE"`

#### Scenario: Instagram is optional
- **WHEN** `instagram` is omitted from the request body
- **THEN** registration SHALL proceed normally and `instagram` in the response SHALL be an empty string

### Requirement: Email must be unique across all active accounts
The system SHALL reject registration if the provided email already belongs to an active (non-deleted) user account.

#### Scenario: Duplicate email
- **WHEN** a POST request is made with an `email` that already exists in `user.user_management` with `deleted_at IS NULL`
- **THEN** the system returns HTTP 409 with `errors[].code = "CONFLICT"` and no new rows are written to the database

#### Scenario: Soft-deleted email can be re-registered
- **WHEN** a POST request is made with an `email` that exists in `user.user_management` but with `deleted_at IS NOT NULL`
- **THEN** registration SHALL proceed normally as a new account

### Requirement: Password is stored securely
The system SHALL hash the password with Argon2id before any database write. The plaintext password SHALL never be logged or persisted.

#### Scenario: Password is hashed on storage
- **WHEN** registration succeeds
- **THEN** the value stored in `user.user_management.password` SHALL be an Argon2id PHC-format string, not the original plaintext

### Requirement: Registration response conforms to envelope format
The system SHALL wrap all responses in the standard envelope format.

#### Scenario: Success envelope
- **WHEN** registration succeeds
- **THEN** the response SHALL match `{ "success": true, "meta": { "requestId": "...", "timestamp": "..." }, "data": { ... }, "errors": [] }`

#### Scenario: Error envelope
- **WHEN** registration fails for any reason
- **THEN** the response SHALL match `{ "success": false, "meta": { ... }, "data": null, "errors": [{ "field": "...", "code": "...", "meta": {}, "message": "..." }] }`
