# Topics
* [Auth configuration](#auth-configuration)
  * [OIDC Authentication](#oidc-Authentication)
  * [Basic authentication](#basic-authentication)

## Auth configuration

FastTrackML supports 2 types of authentication: OIDC based authentication and Basic authentication.

### OIDC Authentication

To enable OIDC authentication, FastTrackML should be run with next command line parameters:
```
--auth-oidc-client-id client_id
--auth-oidc-client-secret client_secret 
--auth-oidc-provider-endpoint http://127.0.0.1 
--auth-oidc-admin-role admin 
--auth-oidc-claim-roles groups 
--auth-oidc-scopes email openid 
```
where:
- `auth-oidc-client-id` - is IDP client id.
- `auth-oidc-client-secret` - is IDP client secret.
- `auth-oidc-provider-endpoint` - is IDP discover endpoint.
- `auth-oidc-admin-role` - `admin` role name(optional). If omitted, then `admin` resources will be disabled. 
- `auth-oidc-claim-roles` - property in `claims` which identify array of `roles` to inspect. claims example:
  ```
  {
     "roles": ["role1", "role2"]
  }
  ```
  or
  ```
  {
     "groups":  ["role1", "role2"]
  }
  ```
  so in that case `auth-oidc-claim-roles` could be `roles` or `groups`. 
Relation between roles and namespaces has to be configured inside the database.
- `auth-oidc-scopes` - list of `scopes` which will be requested from IDP and be present in `claims`.

### Basic authentication

Basic authentication supports 2 different ways:
```
--auth-username username --auth-password password
```
in that case FastTrackML will be restricted with `username` and `password` and user will have access to all the existing resources: `mlflow`, `aim`, `admin`, `chooser`.
```
--auth-username username --auth-password password --auth-users-config /path/to/config.file
```
where:
- `auth-users-config` is a `users` configuration file which should have the following format:
```
users:
  - name: user1
    password: password1
    roles:
      - admin
  - name: user2
    password: password2
    roles:
      - ns:default
      - ns:first
      - ns:second
      - ns:third
  - name: user3
    password: password3
    roles:
      - ns:default
      - ns:third
```
so in that case FastTrackML will use `auth-username` and `auth-password` to check that this user exists in 
`auth-users-config` file and user has all the necessary permissions to access to the requested resource. 
Access will be restricted based on provided `roles` in `auth-users-config` file. 
Special role `admin` gives user access to all the available resources and namespaces: `aim`, `mlflow`, `admin`, `chooser`.