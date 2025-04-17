# Organisation API

This module provides API endpoints for managing organisations in the go-saas platform.

## API Endpoints

All endpoints are under the `/auth/organisations` path and require authentication.

### List Organisations

```
GET /auth/organisations
```

Returns a list of organisations that the authenticated user is a member of.

### Create Organisation

```
POST /auth/organisations
```

Creates a new organisation with the authenticated user as the owner.

**Request Body:**
```json
{
  "name": "My Organisation",
  "description": "Description of my organisation"
}
```

### Get Organisation

```
GET /auth/organisations/{organisationID}
```

Returns details of a specific organisation. The user must be a member of the organisation.

### Update Organisation

```
PUT /auth/organisations/{organisationID}
```

Updates an organisation's details. The user must be the owner of the organisation.

**Request Body:**
```json
{
  "name": "Updated Organisation Name",
  "description": "Updated description"
}
```

### Delete Organisation

```
DELETE /auth/organisations/{organisationID}
```

Deletes an organisation. The user must be the owner of the organisation.

### Add Organisation Member

```
POST /auth/organisations/{organisationID}/members
```

Adds a member to an organisation. The user must be the owner of the organisation.

**Request Body:**
```json
{
  "account_id": "user-account-id",
  "role": "member"
}
```

### List Organisation Members

```
GET /auth/organisations/{organisationID}/members
```

Returns a list of members of an organisation. The user must be a member of the organisation.

### Remove Organisation Member

```
DELETE /auth/organisations/{organisationID}/members/{accountID}
```

Removes a member from an organisation. The user must be the owner of the organisation.

### Update Organisation Member Role

```
PUT /auth/organisations/{organisationID}/members/{accountID}/role
```

Updates a member's role in an organisation. The user must be the owner of the organisation.

**Request Body:**
```json
{
  "role": "admin"
}
```

## Roles

The following roles are supported:

- `owner`: The owner of the organisation. Has full access to the organisation.
- `admin`: An administrator of the organisation. Has access to most organisation features.
- `member`: A regular member of the organisation. Has limited access to organisation features. 