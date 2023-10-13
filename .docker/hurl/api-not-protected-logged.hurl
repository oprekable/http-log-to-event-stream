GET http://host.docker.internal:8000/two/users

POST http://host.docker.internal:8000/two/users
{
    "action": "ADD_USER",
    "userID": "test-user",
    "fullName": "test-full-name"
}

PUT http://host.docker.internal:8000/two/users
{
    "action": "UPDATE_USER",
    "userID": "test-user",
    "fullName": "test-full-name"
}