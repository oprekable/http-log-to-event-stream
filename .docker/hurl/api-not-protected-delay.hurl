GET http://host.docker.internal:8000/three/users

POST http://host.docker.internal:8000/three/users
{
    "action": "ADD_USER",
    "userID": "test-user",
    "fullName": "test-full-name"
}

PUT http://host.docker.internal:8000/three/users
{
    "action": "UPDATE_USER",
    "userID": "test-user",
    "fullName": "test-full-name"
}