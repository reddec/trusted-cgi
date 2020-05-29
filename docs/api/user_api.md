# UserAPI

User/admin profile API


* [UserAPI.Login](#userapilogin) - Login user by username and password. Returns signed JWT
* [UserAPI.ChangePassword](#userapichangepassword) - Change password for the user



## UserAPI.Login

Login user by username and password. Returns signed JWT

* Method: `UserAPI.Login`
* Returns: `*Token`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | login | `string` |
| 1 | password | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "UserAPI.Login",
    "params" : []
}
EOF
```
### Token

```go
type Token struct {
}
```

## UserAPI.ChangePassword

Change password for the user

* Method: `UserAPI.ChangePassword`
* Returns: `bool`

* Arguments:

| Position | Name | Type |
|----------|------|------|
| 0 | token | `*Token` |
| 1 | password | `string` |

```bash
curl -H 'Content-Type: application/json' --data-binary @- "https://127.0.0.1:3434/u/" <<EOF
{
    "jsonrpc" : "2.0",
    "id" : 1,
    "method" : "UserAPI.ChangePassword",
    "params" : []
}
EOF
```
### Token

```go
type Token struct {
}
```