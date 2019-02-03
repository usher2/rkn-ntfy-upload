Upload files by HTTP REST API
=============================

### Client API

* **URL**

https://api.vkostre.org/api-01

## HTTP methods

| HTTP METHOD   | POST         | GET                 |
|---------------|--------------|---------------------|
| /upload       | Upload files | -                   |
| /task/<task>  | -            | Check verify status |


# Upload two files

* Request

  * **URL:** `https://api.vkostre.org/api-01/upload` <br />
    **Method:** `POST` <br />
    **EXAMPLE:** `curl -X POST -F "file1=@/somepath/somefile1" -F "file2=@/somepath/somefile2" https://api.vkostre.org/api-01/upload`

* Success Response

  * **Code:** 201 <br />
    **Content:** `{ task: "<task>" }`

* Error Response

  * **Code:** 500 <br />
    **Content:** `{ error: "server_error" }`

  * **Code:** 400 <br />
    **Content:** `{ error: "file_too_big" }`

  * **Code:** 400 <br />
    **Content:** `{ error: "invalid_request" }`

# Check status

* Request

  * **URL:** `https://api.vkostre.org/api-01/task/<task>` <br />
    **Method:** `GET` <br />
    **EXAMPLE:** `curl -X GET https://api.vkostre.org/api-01/task/<task>`

* Success Response

  * **Code:** 202 <br />
    **Content:** `{ status: "wait" }` <br />
    **Description:** Waiting for stage1 verification

  * **Code:** 200  <br />
    **Content:** `{ status: "ok" }` <br />
    **Description:** Verification passed

  * **Code:** 200  <br />
    **Content:** `{ status: "failed" }` <br />
    **Description:** Verification failed

* Error Response

  * **Code:** 500  <br />
    **Content:** `{ error: "server_error" }` <br />
    **Description:** Something error

  * **Code:** 400  <br />
    **Content:** `{ error: "invalid_request" }` <br />
    **Description:** Bad request

  * **Code:** 400  <br />
    **Content:** `{ error: "invalid_task" }` <br />
    **Description:** Task number not found?

