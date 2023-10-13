GET http://host.docker.internal:8000/one/users
Authorization: Bearer asdsadsadsasdsadsadsa

POST http://host.docker.internal:8000/one/users
Authorization: Bearer asdsadsadsasdsadsadsa
{
    "action": "ADD_USER",
    "userID": "test-user",
    "fullName": "test-full-name"
}

PUT http://host.docker.internal:8000/one/users
Authorization: Bearer asdsadsadsasdsadsadsa
{
    "action": "UPDATE_USER",
    "userID": "test-user",
    "fullName": "test-full-name"
}
