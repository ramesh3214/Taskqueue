# TaskFlow

TaskFlow is a backend system built with Go to understand and implement production-oriented backend concepts such as authentication, caching, asynchronous task processing, message queues, background workers, retries, dead-letter queues, email processing, database management, and Docker-based deployment.

The project uses a layered architecture and separates the API server from background workers.

## Features

* User signup and login
* JWT-based authentication
* Password hashing using bcrypt
* Protected API routes
* PostgreSQL database
* GORM ORM
* Redis profile caching
* RabbitMQ message queue
* Asynchronous background task processing
* Dedicated worker service
* Welcome email processing
* Retry mechanism
* Dead Letter Queue
* Manual RabbitMQ acknowledgements
* Persistent RabbitMQ messages
* Dockerized API
* Dockerized worker
* Dockerized RabbitMQ
* External PostgreSQL database
* External Redis Cloud
* Environment-based configuration
* Graceful API shutdown
* Distributed-worker-ready architecture

## Architecture

```text
                         Client
                           |
                           v
                    +--------------+
                    |   Go API      |
                    |    Server     |
                    +--------------+
                           |
                           v
                     Controller
                           |
                           v
                       Service
                      /       \
                     /         \
                    v           v
             PostgreSQL       Redis
                    |
                    |
                    v
               Task Creation
                    |
                    v
              RabbitMQ Exchange
                    |
                    v
                Task Queue
                    |
                    v
              Background Worker
                    |
                    v
               Task Processing
                    |
                    v
               Email Service
```

## Project Structure

```text
taskflow/
|
├── cmd/
│   ├── api/
│   │   └── main.go
│   │
│   └── worker/
│       └── main.go
│
├── config/
│
├── route/
│
├── controller/
│
├── service/
│
├── repository/
│
├── dto/
│
├── model/
│
├── queue/
│   └── rabbitmq.go
│
├── worker/
│   ├── task_worker.go
│   └── email_worker.go
│
├── middleware/
│
├── util/
│
├── database/
│
├── Dockerfile
├── docker-compose.yml
├── .dockerignore
├── .env
├── .env.example
├── .gitignore
└── go.mod
```

## Layered Architecture

The API follows a layered architecture:

```text
HTTP Request
      |
      v
Controller
      |
      v
Service
      |
      v
Repository
      |
      v
PostgreSQL
```

### Controller

The controller handles HTTP requests and responses.

Responsibilities include:

* Reading request data
* Validating basic request input
* Calling service methods
* Returning HTTP responses

### Service

The service layer contains the main business logic.

Responsibilities include:

* Authentication logic
* Password hashing
* JWT generation
* Task creation
* Cache handling
* Calling repositories
* Publishing background tasks

### Repository

The repository layer communicates with PostgreSQL through GORM.

Responsibilities include:

* Creating records
* Reading records
* Updating records
* Deleting records
* Database-specific operations

### DTO

DTOs are used to define the structure of data exchanged between different parts of the application.

Example RabbitMQ task message:

```go
type TaskMessage struct {
    TaskID   uint   `json:"task_id"`
    TaskType string `json:"task_type"`
}
```

## Authentication

TaskFlow uses JWT-based authentication.

Authentication flow:

```text
Signup
  |
  v
Password
  |
  v
bcrypt Hash
  |
  v
PostgreSQL
```

Login flow:

```text
Email + Password
       |
       v
Find User
       |
       v
Compare bcrypt Hash
       |
       v
Generate JWT
       |
       v
Return Token
```

Protected request:

```text
Client
  |
  | Authorization: Bearer <token>
  v
JWT Middleware
  |
  v
Validate Token
  |
  v
Protected Controller
```

## Password Security

Passwords are never stored directly.

The password is hashed using bcrypt before being stored in PostgreSQL.

```text
Plain Password
      |
      v
bcrypt
      |
      v
Password Hash
      |
      v
PostgreSQL
```

During login, the provided password is compared with the stored bcrypt hash.

## PostgreSQL

PostgreSQL is the primary persistent database.

The project uses:

* PostgreSQL
* GORM
* Aiven PostgreSQL

The application connects using the `DB_URL` environment variable.

Example:

```env
DB_URL=your_postgresql_connection_string
```

Database migrations are performed using GORM:

```go
db.AutoMigrate(
    &model.Auth{},
    &model.Task{},
)
```

## Database Models

### Auth

```go
type Auth struct {
    ID       uint   `gorm:"primaryKey" json:"id"`
    Name     string `json:"name"`
    Email    string `json:"email"`
    Password string `json:"password"`
    Age      int    `json:"age"`
}
```

### Task

```go
type Task struct {
    ID         uint   `gorm:"primaryKey"`
    UserID     uint
    Email      string
    TaskType   string
    RetryCount int
    Status     string
}
```

Task status values:

```text
PENDING
COMPLETED
FAILED
```

## Redis Caching

Redis is used for profile caching.

Redis is not the primary database.

PostgreSQL remains the source of truth.

The application uses a cache-aside pattern.

```text
Get Profile
     |
     v
   Redis
     |
     +---- HIT ----> Return Cached Profile
     |
     |
     +---- MISS
           |
           v
      PostgreSQL
           |
           v
       Redis SET
           |
           v
      Return Profile
```

Profile cache key:

```text
user:profile:<userID>
```

Cache expiration:

```text
10 minutes
```

Redis configuration:

```env
REDIS_ADDR=your_redis_address
REDIS_USERNAME=your_redis_username
REDIS_PASSWORD=your_redis_password
```

When a profile is updated or deleted, the related Redis cache is invalidated.

## RabbitMQ

RabbitMQ is used as the message broker for asynchronous background tasks.

The API does not directly process the background task.

Instead:

```text
API
 |
 v
Create Task in PostgreSQL
 |
 v
Publish Message
 |
 v
RabbitMQ
 |
 v
Worker
 |
 v
Process Task
```

This keeps the HTTP request independent from background processing.

## RabbitMQ Exchanges

The application uses three exchanges:

```text
task_exchange
retry_exchange
dlq_exchange
```

### Task Exchange

Used for normal task processing.

```text
task_exchange
      |
      | task.created
      v
task_queue
```

### Retry Exchange

Used when task processing fails but the retry limit has not been reached.

```text
retry_exchange
      |
      | task.retry
      v
task_retry_queue
```

### Dead Letter Exchange

Used when a task has permanently failed.

```text
dlq_exchange
      |
      | task.failed
      v
task_dlq
```

## RabbitMQ Queues

The application uses:

```text
task_queue
task_retry_queue
task_dlq
```

### task_queue

Contains normal tasks waiting to be processed.

### task_retry_queue

Contains failed tasks waiting before another processing attempt.

The retry queue uses a message TTL.

Current TTL:

```text
5000 milliseconds
```

After the TTL expires, RabbitMQ routes the message back to the main task exchange.

### task_dlq

Contains tasks that have permanently failed after the retry limit.

## Background Worker

The worker runs separately from the API server.

The worker:

1. Connects to PostgreSQL
2. Connects to RabbitMQ
3. Creates required queues and exchanges
4. Consumes messages
5. Reads task information from PostgreSQL
6. Processes the task
7. Updates task status
8. Acknowledges the RabbitMQ message

Architecture:

```text
RabbitMQ
    |
    v
Task Worker
    |
    v
Get Task From PostgreSQL
    |
    v
Process Task
    |
    +----------------------+
    |                      |
    v                      v
 SUCCESS                 FAILURE
    |                      |
    v                      v
COMPLETED              Retry Count++
    |                      |
    v                      v
   ACK                 Retry Queue
```

## Email Worker

Welcome emails are processed asynchronously.

When a user signs up:

```text
Signup
  |
  v
Create User
  |
  v
Create WELCOME_EMAIL Task
  |
  v
Store Task in PostgreSQL
  |
  v
Publish RabbitMQ Message
  |
  v
Worker
  |
  v
EmailWorker
  |
  v
SMTP
  |
  v
Welcome Email
```

The email worker uses the `EmailService`.

The SMTP configuration is provided through environment variables.

```env
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email
SMTP_PASSWORD=your_app_password
```

For Gmail SMTP, an App Password should be used instead of the normal Gmail account password.

## Task Processing

A task is initially created with:

```text
Status = PENDING
RetryCount = 0
```

Example:

```go
task := &model.Task{
    UserID:     result.ID,
    Email:      result.Email,
    TaskType:   "WELCOME_EMAIL",
    Status:     "PENDING",
    RetryCount: 0,
}
```

The RabbitMQ message contains the task ID and task type.

The worker then retrieves the complete task from PostgreSQL.

This keeps PostgreSQL as the source of truth for task information.

## Successful Task Flow

```text
RabbitMQ Message
      |
      v
Worker Receives Message
      |
      v
Get Task From PostgreSQL
      |
      v
Process Task
      |
      v
Processing Successful
      |
      v
Update Status = COMPLETED
      |
      v
ACK RabbitMQ Message
```

The task is only marked `COMPLETED` after the actual processing succeeds.

## Retry Mechanism

If task processing fails, the worker increments the retry count.

Example:

```text
Initial RetryCount = 0

Failure
  |
  v
RetryCount = 1
  |
  v
Retry Queue
```

The retry queue waits for the configured TTL before sending the message back for another processing attempt.

Current retry condition:

```text
RetryCount <= 3
```

After the retry limit is exceeded, the task is marked as failed.

## Dead Letter Queue

The Dead Letter Queue is used for tasks that cannot be processed successfully after the configured retry attempts.

Flow:

```text
Task
 |
 v
Processing Failure
 |
 v
Retry
 |
 v
Processing Failure
 |
 v
Retry
 |
 v
Processing Failure
 |
 v
Retry Limit Exceeded
 |
 v
Status = FAILED
 |
 v
DLQ
```

The DLQ allows failed messages to be preserved for debugging and future investigation instead of being lost.

## RabbitMQ Acknowledgement

The worker uses manual acknowledgements.

A successful task follows:

```text
Process Task
    |
    v
Update DB = COMPLETED
    |
    v
ACK Message
```

If processing fails:

```text
Process Task
    |
    v
Increment Retry Count
    |
    v
Publish To Retry Queue
    |
    v
ACK Original Message
```

When the retry limit is exceeded:

```text
Process Task
    |
    v
Update DB = FAILED
    |
    v
Publish To DLQ
    |
    v
ACK Original Message
```

`msg.Ack(false)` acknowledges the current RabbitMQ message.

The `false` parameter means that the acknowledgement applies only to the current message and not multiple previously delivered messages.

## Database State vs RabbitMQ State

These are two different concepts.

Database status:

```text
PENDING
COMPLETED
FAILED
```

RabbitMQ acknowledgement:

```text
ACK
```

The database status represents application state.

RabbitMQ ACK represents message acknowledgement to the broker.

They should not be treated as the same thing.

## Docker

TaskFlow is containerized using Docker.

The Docker setup contains:

```text
API Container
Worker Container
RabbitMQ Container
```

PostgreSQL and Redis remain external services.

```text
                    Docker Network
                         |
          +--------------+--------------+
          |              |              |
          v              v              v
       API            Worker        RabbitMQ
          |              |              |
          |              |              |
          +--------------+--------------+
                         |
              External Services
                  /          \
                 /            \
                v              v
          PostgreSQL         Redis
            Aiven          Redis Cloud
```

## Dockerfile

The project uses a multi-stage Docker build.

Build stage:

```text
Go Alpine Image
      |
      v
Download Dependencies
      |
      v
Copy Source Code
      |
      v
Build API Binary
      |
      v
Build Worker Binary
```

Runtime stage:

```text
Alpine
  |
  +---- api
  |
  +---- worker
```

The same Docker image contains both binaries.

Docker Compose starts them as separate containers.

## Docker Compose Services

The Compose setup contains three services:

```text
rabbitmq
api
worker
```

### RabbitMQ

RabbitMQ exposes:

```text
5672
```

for AMQP communication.

Management UI:

```text
15672
```

The RabbitMQ container includes a health check so that API and worker containers wait until RabbitMQ is ready.

### API

The API container runs the API binary.

The API receives HTTP requests and publishes background tasks to RabbitMQ.

### Worker

The worker container runs the worker binary.

The worker consumes messages from RabbitMQ and processes background tasks.

## Docker Networking

Inside Docker Compose, services communicate using service names.

For example, the API and worker connect to RabbitMQ using:

```text
amqp://guest:guest@rabbitmq:5672/
```

The hostname is:

```text
rabbitmq
```

not:

```text
localhost
```

because `localhost` inside a container refers to that same container.

## Environment Variables

The project uses environment variables for configuration and secrets.

Example `.env`:

```env
DB_URL=your_postgresql_connection_string

REDIS_ADDR=your_redis_address
REDIS_USERNAME=your_redis_username
REDIS_PASSWORD=your_redis_password

JWT_TOKEN=your_jwt_secret

RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/

SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email
SMTP_PASSWORD=your_app_password
```

Do not commit `.env` to GitHub.

Use `.env.example` to document the required environment variables without exposing secrets.

Example:

```env
DB_URL=

REDIS_ADDR=
REDIS_USERNAME=
REDIS_PASSWORD=

JWT_TOKEN=

RABBITMQ_URL=

SMTP_HOST=
SMTP_PORT=
SMTP_USERNAME=
SMTP_PASSWORD=
```


## Getting Started

Follow the steps below to run TaskFlow locally after cloning the repository.

### 1. Clone the Repository

```bash
git clone https://github.com/your-username/taskflow.git
```

Move into the project directory:

```bash
cd taskflow
```

### 2. Configure Environment Variables

Create a `.env` file in the root directory of the project.

```text
taskflow/
├── .env
├── docker-compose.yml
├── Dockerfile
├── go.mod
└── ...
```

Add the required environment variables:

```env
DB_URL=your_postgresql_connection_string

REDIS_ADDR=your_redis_address
REDIS_USERNAME=your_redis_username
REDIS_PASSWORD=your_redis_password

JWT_TOKEN=your_jwt_secret

RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/

SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USERNAME=your_email
SMTP_PASSWORD=your_gmail_app_password
```

Do not use your normal Gmail password for SMTP. Use a Gmail App Password.

The `.env` file contains sensitive credentials and should never be committed to GitHub.

### 3. External Services

TaskFlow requires PostgreSQL and Redis.

PostgreSQL is used as the primary database.

Redis is used for profile caching.

The project is configured to use:

* PostgreSQL through Aiven
* Redis through Redis Cloud

You can use other PostgreSQL and Redis providers as long as the required connection details are provided through the environment variables.

RabbitMQ does not need to be installed separately because it is started automatically through Docker Compose.

### 4. Start the Application

Make sure Docker Desktop is installed and running.

Build the Docker images and start all services:

```bash
docker compose up --build
```

To run the application in the background:

```bash
docker compose up --build -d
```

Docker Compose starts:

```text
RabbitMQ
API
Worker
```

Check the running containers:

```bash
docker compose ps
```

Expected services:

```text
taskflow-rabbitmq
taskflow-api
taskflow-worker
```

### 5. Check API Logs

To view API logs:

```bash
docker compose logs -f api
```

The API should start successfully after connecting to PostgreSQL, Redis, and RabbitMQ.

### 6. Check Worker Logs

To view worker logs:

```bash
docker compose logs -f worker
```

The worker should show messages similar to:

```text
Database connected successfully!
RabbitMQ connected successfully!
RabbitMQ setup successful!
Task worker is listening for messages...
```

### 7. Check RabbitMQ

Open the RabbitMQ management dashboard:

```text
http://localhost:15672
```

Default credentials:

```text
Username: guest
Password: guest
```

From the dashboard you can inspect:

* Exchanges
* Queues
* Messages
* Consumers
* Connections
* Bindings

### 8. Check API Health

Once the API container is running, open:

```http
GET /health
```

Example:

```bash
curl http://localhost:50000/health
```

Expected response:

```json
{
    "status": "ok",
    "message": "server is working"
}
```

### 9. Test the Background Task Flow

After starting the API and worker, create a new user using the signup API.

The application will perform the following flow:

```text
Signup Request
      |
      v
API
      |
      v
Create User in PostgreSQL
      |
      v
Create WELCOME_EMAIL Task
      |
      v
Publish Task to RabbitMQ
      |
      v
Worker Receives Task
      |
      v
Worker Fetches Task From PostgreSQL
      |
      v
Email Worker
      |
      v
SMTP
      |
      v
Welcome Email
      |
      v
Task Status = COMPLETED
```

You can monitor the complete flow using:

```bash
docker compose logs -f api worker
```

### 10. Stop the Application

To stop all containers:

```bash
docker compose down
```

To stop containers and remove the Docker network:

```bash
docker compose down
```

To rebuild the application after making code changes:

```bash
docker compose up --build
```

### 11. Development Without Docker

If you want to run the Go application directly on your local machine instead of Docker, make sure Go is installed.

Check the Go version:

```bash
go version
```

Install dependencies:

```bash
go mod download
```

Run the API:

```bash
go run ./cmd/api
```

Run the worker in another terminal:

```bash
go run ./cmd/worker
```

For this setup, RabbitMQ must be running locally and the `RABBITMQ_URL` should point to the local RabbitMQ instance.

Example:

```env
RABBITMQ_URL=amqp://guest:guest@localhost:5672/
```

When running with Docker Compose, use the Docker service name instead:

```env
RABBITMQ_URL=amqp://guest:guest@rabbitmq:5672/
```

### 12. Complete Setup Summary

For a fresh setup, the process is:

```text
Clone Repository
      |
      v
Create .env
      |
      v
Add PostgreSQL Credentials
      |
      v
Add Redis Credentials
      |
      v
Add JWT Secret
      |
      v
Add SMTP Credentials
      |
      v
Run Docker Compose
      |
      v
RabbitMQ Starts
      |
      v
API Starts
      |
      v
Worker Starts
      |
      v
Open /health
      |
      v
Test Signup
      |
      v
Background Task Processing
```

After these steps, TaskFlow should be ready to run locally.


## Running With Docker

Make sure Docker is installed and running.

Build and start the complete application:

```bash
docker compose up --build
```

Run in detached mode:

```bash
docker compose up --build -d
```

Check running containers:

```bash
docker compose ps
```

View API logs:

```bash
docker compose logs -f api
```

View worker logs:

```bash
docker compose logs -f worker
```

View RabbitMQ logs:

```bash
docker compose logs -f rabbitmq
```

Stop the application:

```bash
docker compose down
```

## RabbitMQ Management UI

RabbitMQ provides a web-based management interface.

Open:

```text
http://localhost:15672
```

Default local credentials:

```text
Username: guest
Password: guest
```

From the management UI, you can inspect:

* Exchanges
* Queues
* Messages
* Consumers
* Connections
* Bindings
* Queue depth

## API

The API runs inside the Docker network and is exposed to the host through the configured port.

Health endpoint:

```http
GET /health
```

Example response:

```json
{
    "status": "ok",
    "message": "server is working"
}
```

## Testing Background Task Flow

A typical test flow is:

```text
1. Start Docker Compose
2. Start API
3. Start Worker
4. Call Signup API
5. User is created in PostgreSQL
6. Task is created with PENDING status
7. Task message is published to RabbitMQ
8. Worker receives the message
9. Worker fetches task from PostgreSQL
10. Worker processes WELCOME_EMAIL
11. Email is sent
12. Task status becomes COMPLETED
13. RabbitMQ message is ACKed
```

## Testing Retry

To test retry behaviour, the email processing can temporarily be configured to return an error.

Example:

```go
case "WELCOME_EMAIL":
    return fmt.Errorf("simulated email sending failure")
```

The worker will then:

```text
Process Task
     |
     v
Failure
     |
     v
Increment RetryCount
     |
     v
Retry Queue
     |
     v
Wait for TTL
     |
     v
Main Queue
     |
     v
Process Again
```

After the retry limit is exceeded:

```text
Status = FAILED
      |
      v
Dead Letter Queue
```

## Distributed Workers

The architecture supports running multiple worker instances.

Example:

```text
                 RabbitMQ
                    |
          +---------+---------+
          |         |         |
          v         v         v
       Worker 1  Worker 2  Worker 3
```

RabbitMQ distributes messages between consumers.

This allows the application to scale background task processing horizontally.

For example, additional worker instances can be started when task volume increases.

## Concurrency

Multiple workers can process different messages concurrently.

For example:

```text
Task 1 -> Worker 1
Task 2 -> Worker 2
Task 3 -> Worker 3
Task 4 -> Worker 1
```

Concurrency and atomicity are different concepts.

RabbitMQ can distribute messages between workers, but database operations may still require additional protections depending on the application's requirements.

## Idempotency

Background workers should consider idempotency.

For example, if an email is successfully sent but the worker crashes before acknowledging the RabbitMQ message, RabbitMQ may deliver the message again.

This can potentially result in the same email being sent more than once.

A production system can use an idempotency strategy to prevent duplicate processing.

Possible future approaches include:

* Idempotency keys
* Task execution records
* Database constraints
* Processed-message tracking
* Transactional outbox patterns

## Security

The project follows basic security practices:

* Passwords are hashed using bcrypt
* JWT is used for authentication
* Secrets are stored in environment variables
* `.env` is excluded from Git
* Redis credentials are not hardcoded
* Database credentials are not hardcoded
* SMTP credentials are not hardcoded

Production deployments should also use:

* HTTPS
* Strong JWT secrets
* Proper secret management
* Network restrictions
* Secure database credentials
* Secure Redis credentials
* Secure RabbitMQ credentials

## Graceful Shutdown

The API listens for operating system termination signals.

When a shutdown signal is received:

```text
Shutdown Signal
      |
      v
Stop Accepting Requests
      |
      v
Graceful HTTP Server Shutdown
      |
      v
Release Resources
```

The application uses a timeout context for graceful shutdown.

## Current Technology Stack

### Backend

* Go
* Gorilla Mux
* GORM

### Database

* PostgreSQL
* Aiven PostgreSQL

### Cache

* Redis
* Redis Cloud

### Message Queue

* RabbitMQ

### Email

* SMTP
* Gmail SMTP

### Authentication

* JWT
* bcrypt

### Infrastructure

* Docker
* Docker Compose

## Why This Project Was Built

The purpose of TaskFlow is to move beyond basic CRUD APIs and understand how backend systems work when asynchronous processing, caching, queues, workers, retries, and external services are involved.

The project focuses on understanding the flow between different backend components rather than only implementing individual APIs.

## Learning Goals

The project was built to understand:

* Go backend architecture
* Layered architecture
* REST APIs
* Authentication
* JWT
* Password hashing
* PostgreSQL
* GORM
* Redis caching
* Cache invalidation
* RabbitMQ
* Exchanges
* Queues
* Routing keys
* Message acknowledgements
* Background workers
* Asynchronous processing
* Retry mechanisms
* Dead Letter Queues
* SMTP email processing
* Docker
* Docker Compose
* Container networking
* Distributed workers
* Idempotency concepts
* Production-oriented backend design

## Future Improvements

Possible future improvements include:

* Worker concurrency configuration
* Better retry tracking
* Exponential backoff
* Improved idempotency handling
* Transactional outbox pattern
* Prometheus metrics
* Grafana dashboards
* Structured logging
* Distributed tracing
* Better error monitoring
* RabbitMQ authentication and permissions
* Production secret management
* Kubernetes deployment
* Automated CI/CD
* Automated tests
* Integration tests
* Load testing

## Author

Ramesh Sah

2026 B.Tech Computer Science Graduate

Interested in Backend Development, Full-Stack Development, Software Development, Distributed Systems, and System Design.
