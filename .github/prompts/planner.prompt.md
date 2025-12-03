---
agent: 'agent'
model: 'Claude Opus 4.5 (Preview)'
description: 'Generate production-ready plan to develop golang project with proper structure, Makefile, Dockerfile, CI/CD pipeline, and unit tests based on user requirements.'
---

# ComIO Community IO Storage

Generate a plan (not code) to develop a production-ready storage solution in Golang named "ComIO" (Community IO Storage).

the plan must be written in markdown format into apply.prompt.md in the .github/prompts directory.

DO NOT GENERATE ANY CODE. ONLY GENERATE A DETAILED PLAN. this plan would be readed by other ai agent that will implement the code based on this plan. format it to be easily readable by other ai agents in instructions format.

## Plan template

````
---
agent: 'agent'
model: 'Claude Opus 4.5 (Preview)'
description: 'Generate production-ready golang project with proper structure, Makefile, Dockerfile, CI/CD pipeline, and unit tests based on user requirements.'
---
<plan code here>
```


## Project Scope

The ComIO project aims to create a production-ready storage solution with the following features:
- RESTful API server for storage operations s3 compliant
- Command-line interface (CLI) using cobra for managing the storage system
- Storage Replication across multiple nodes
- Handle raw device for storage like full disk (e.g. /dev/sdb ) or partition (e.g. /dev/sdb1)
- User authentication and authorization

## Project Requirements

1. **Project Structure**: Follow best practices for Golang project structure, including proper package organization.
2. **Dependencies**: Use Go Modules for dependency management.
3. **Server Setup**: Implement a RESTful API server using a popular framework (e.g., Gin, Echo).
4. **Error Handling**: Implement robust error handling and logging throughout the application.
5. **Documentation**: Include comprehensive documentation for the codebase and API endpoints.
6. **CLI**: Implement a command-line interface for managing the storage system, including commands for start server, handle client and admin operation.
7. **Configuration Management**: Use environment variables or configuration files for managing application settings, you can use cobra that will help to create CLI applications.
8. **Makefile**: Create a Makefile to automate common tasks such as building,
9. **Dockerfile**: Provide a Dockerfile for containerizing the application.

## go.mod Template

```go
module github.com/danielino/comio

go 1.25
```

## Project Specifications

- this project relies on Bucket concept to organize objects
- each Bucket can have multiple objects stored within it
- objects are identified by unique keys within their respective Buckets
- implement versioning for objects to allow multiple versions of the same object to coexist
- support multipart uploads for large objects to enhance upload efficiency and reliability
- implement lifecycle policies to automate the management of objects over time (e.g., transition to different storage classes, expiration)
- ensure data integrity through checksums and validation mechanisms
- provide detailed logging and monitoring capabilities for storage operations
- implement backup and restore functionalities to safeguard against data loss
- ensure compliance with relevant data protection regulations and standards
- design the system to be scalable and capable of handling increasing amounts of data and requests

## Best Practices

- Keep tools focused and single-purpose
- Use descriptive names for types and functions
- Include JSON schema documentation in struct tags
- Always respect context cancellation
- Return descriptive errors
- Keep main.go minimal, logic in packages
- Write tests for tool handlers
- Document all exported functions
- Use interfaces for dependencies to facilitate testing
- Follow idiomatic Go conventions
- Ensure code is formatted with `gofmt`
- Use `go vet` to catch potential issues
- Implement logging with a structured logger like logrus or zap
- do not overengineer, keep it simple and pragmatic
- Write unit tests that cover at least 80% of the codebase