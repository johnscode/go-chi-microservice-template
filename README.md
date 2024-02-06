# go-chi-microservice-template
A simple boilerplate microservice template using golang and chi

This app template makes it pretty easy to get a server and a set of REST
endpoints up and running. The template models a simple user service and 
is based off the rest example from the [chi repository](https://github.com/go-chi/chi).

This app template is already setup with logging and configuration using the server environment.
Environment is used for configuration rather than files to discourage the use of secrets files
(_which is a whole other conversation_)

Zerolog is used for logging due to its efficiency and versatile formatting rather 
than the builtin log module.

## To Do
- implement Put, Post and Delete on user
- implement user search
- dockerize it
- database (mongo, postgres, ???)
- gRPC and GraphQL endpoints (or maybe separate templates???)