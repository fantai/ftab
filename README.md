
# .http file and *REST Client*

A `.http` file is used to describe what will be send to a HTTP service, it's more convenient than `cURL` command, *vscode* has a great plugin named [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client), it extend `.http` file with addition features, such as

- variable repalce
- named request
- reference value from named request
- and more ...

with these features, user can express more complex business logic, for example

1. user login
2. refer the auth information after step 1, to check the user's profile

```http
# @name login
POST {{server}}/login
Content-Type: application/json

{
    "name": "",
    "password": "",
}


###

POST {{server}}/profile
Content-Type: application/json

{
    "auth": "{{login.response.body.$.auth}}",
    "kind": "all"
}
```

more information can be found in [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client)


# test cases

usually, in service development state, we write `.http` files to test our service, after development stage, these files can also be used as test cases, test cases should be executed by machine instead human round by round,  `ftab` is a tool to execute the `.http` file.

benchmark is also important for serivce, `ftab` can execute `.http` file with `requests` and `currency` like `ab` do with URL. for example

```
> ftab -i order.http -n 10000 -c 200
Total Requests      : 10,000
Currency            : 200
Successed           : 10,000
Failed              : 0

Time Used           : 19.0007ms
Reqeusts Per Second : 1,223/S
Send Speed          : 111.941K/S
Recv Speed          : 29.758K/S

Avg Time Used       : 19.0007ms
Min Time Used       : 19.0007ms
Max Time Used       : 19.0007ms

P50 Time Used       : 19.0007ms
P75 Time Used       : 19.0007ms
P90 Time Used       : 19.0007ms
P95 Time Used       : 19.0007ms
P99 Time Used       : 19.0007ms
```

# usage

1. get the prebuilt binary from [Release] or just comiple it by yourself
2. move the ftab executable file to a folder within `PATH`
3. type `ftab` see it's args

by default, `ftab` execute the `.http` named `test.http` in current folder once, output result in human readable format.


# *REST Client* compatible

`REST Client` have lots of features, supported features is list below

* [x] Send/Cancel/Rerun __HTTP request__ 
* [ ] Send __GraphQL query__ and author __GraphQL variables__ 
* [x] Organize _MULTIPLE_ requests in the same file (separated by `###` delimiter)
* [ ] Save raw response and response body only to local disk
* [ ] Authentication 
* [x] Environments and custom/system variables support
* [ ] Remember Cookies for subsequent requests
* [ ] Proxy support
* [ ] Send SOAP requests, as well as snippet support to build SOAP envelope easily
* [x] `HTTP` language support

## `ftab` enhance

- reality mock variable with localization (current only chinese supported )
    - human name
    - ID card NO
    - mobile number
    - email address