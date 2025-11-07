# Overview #

This project is a Payment Processing Proof of Concept (POC) built using Golang, Redis, Beanstalkd, and Docker Compose.

The goal of the POC is to demonstrate API Rate Limiting, Asynchronous Job Queueing, and Worker-based Background Processing inside a Microservices-inspired architecture — all using lightweight tools suitable for learning and interviews.

The application consists of two main services:

## API Service (HTTP Server) ##

- Accepts new payment requests

- Applies Redis-based rate limiting

- Stores each payment in Redis

- Enqueues payment jobs in Beanstalkd

- Returns the created payment with a PENDING status

## Worker Service ##

- Subscribes to a Beanstalkd queue

- Processes payment jobs asynchronously

- Simulates payment processing delay

- Updates payment status in Redis (PENDING → COMPLETED)

- Runs independently from the API

# Architecture #

``` mermaid
flowchart LR
    subgraph Client["Client"]
        A[POST /payment]
    end

    subgraph API["API Service"]
        B[Rate Limiter - Redis]
        C[Create Payment - Save to Redis]
        D[Enqueue Payment Job - Beanstalkd]
    end

    subgraph Queue["Beanstalkd Queue"]
        E[(payment_jobs)]
    end

    subgraph Worker["Worker Service"]
        F[Dequeue Job]
        G[Simulate Processing Delay]
        H[Update Payment Status in Redis]
    end

    subgraph Redis["Redis Database"]
        R[(payments)]
    end

    A --> B --> C --> D --> E
    F --> E
    F --> G --> H --> R
    C --> R
```

## System Architecture ##

This project implements a lightweight microservices-inspired payment processing system built around three core components:

### 1. API Service ###

The API is responsible for handling incoming payment requests from clients. It exposes an HTTP endpoint and performs the following tasks:

- Rate limiting using Redis (simple token-bucket approach)

- Validation and creation of a new payment object

- Persistence of the payment object in Redis

- Job queueing by pushing a serialized payment job to Beanstalkd

- Immediate response to the caller with status PENDING

The API does not perform any heavy processing — it only orchestrates the first step.

### 2. Worker Service ###

The worker is a long-running background service. It subscribes to a Beanstalkd tube (payment_jobs) and processes payment tasks asynchronously.

Its responsibilities:

- Retrieve jobs from Beanstalkd

- Deserialize the payment payload

- Simulate processing delay (to mimic real payment workflows)

- Update payment status in Redis

- Log processing progress

- Support scalable concurrency (you can run multiple worker instances)

This separation of concerns ensures the API remains fast and responsive even under load.

### 3. Redis (Shared Data Store) ###

Redis is used for:

- Shared persistence of payment objects

- Rate limiting counters

- Inter-service communication (API and Worker share the same payment storage)

Using Redis avoids the need for migrations or a full SQL database — ideal for this POC.

### 4. Beanstalkd (Message Queue) ###

Beanstalkd provides a simple, lightweight job queue for dispatching work to the worker:

- The API pushes serialized payment jobs

- Worker pulls jobs from the queue

- Supports retries, burying, delays, and multiple workers

### Request Flow (Summary) ###

1) Client → API
Sends a POST request with payment details.

2) API → Redis
Checks/updates rate limit.

3) API → Redis
Stores payment with status PENDING.

4) API → Beanstalkd
Pushes job to payment_jobs tube.

5) Worker → Beanstalkd
Retrieves job.

6) Worker → Redis
Updates payment status to COMPLETED.

7) Client → API → Redis
Client can query payment status anytime.


# Setup Instructions #

This project is fully containerized using Docker Compose, and requires only Docker installed on your machine. No additional dependencies (Go, Redis, Beanstalkd, etc.) are needed locally.

Follow the steps below to get the API, Worker, Redis, and Beanstalkd all running together.

1) ### Clone the Repository

```
git clone https://github.com/Ngang-muma-musa/payment_service_poc.git
cd payment_service_poc
```

2) ### Create the Environment File ###

    Create a new `.env` file by copying the contents of `.env.example`, then update the values according to your environment.

3) ### build and start project
    This builds the API and Worker binaries using the Dockerfile and run project:

    ```bash 
    make dev
    ```

    view logs

    ```
    make logs
    ```


# Simulation steps #

1) ### Start all services (API,Worker,Redis,Queue,Ui)
```
make dev
```
health check

```bash
curl http://localhost:8000/health
```
success response
```json
{
  "status": "OK"
}
```
2) ### Submit a payment Request

    Use `curl`, Postman or any RESR client

    Example request:
```bash
curl -X POST http://localhost:8000/payments \
  -H "Content-Type: application/json" \
  -d '{
    "user_Id": "user_123",
    "amount": 1500,
    "currency": "XAF"
  }'
```

Expected Response:
```json
{
    "data": {
        "ID": "716e7646-c573-4595-8666-9b58f908364c",
        "user_id": "user:101",
        "amount": 1000,
        "currency": "XAF",
        "status": "pending",
        "initiated_at": "2025-11-05T19:50:00.84307459Z",
        "UpdatedAt": "0001-01-01T00:00:00Z"
    },
    "message": "Payment queued successfully",
    "status": "success"
}
```

- ✅ Payment created
- ✅ Stored in Redis
- ✅ Job sent to Beanstalkd
- ✅ Worker will soon process it

3)  ### View the Job in Beanstalkd UI
```
http://localhost:8902
```
You can now see:

- Ready jobs

- Reserved jobs

- Delayed jobs

- Buried jobs

4) ### Watch the Worker Process the Job

```nginx
worker-1  | 2025/11/05 19:48:40 Worker #1 started
worker-1  | 2025/11/05 19:48:40 Worker started...
worker-1  | 2025/11/05 19:49:50 Processing payment ID: edd4147d-66ad-40a2-82e9-a8ea890c6af1 for user user:101 amount: 1000.00 XAF
worker-1  | 2025/11/05 19:49:53 Payment edd4147d-66ad-40a2-82e9-a8ea890c6af1 processed successfully
worker-1  | 2025/11/05 19:49:55 Processing payment ID: 5b1834b8-f404-46c6-911c-b7ed7eeb2668 for user user:101 amount: 1000.00 XAF
worker-1  | 2025/11/05 19:49:58 Payment 5b1834b8-f404-46c6-911c-b7ed7eeb2668 processed successfully
worker-1  | 2025/11/05 19:50:00 Processing payment ID: 716e7646-c573-4595-8666-9b58f908364c for user user:101 amount: 1000.00 XAF
worker-1  | 2025/11/05 19:50:03 Payment 716e7646-c573-4595-8666-9b58f908364c processed successfully
```

Worker behavior:

- ✅ Retrieves job
- ✅ Simulates processing using time.Sleep
- ✅ Updates status in Redis
- ✅ Deletes the job from the queue

5) ### Verify Payment Status After Processing

```
curl http://localhost:8080/payments/{paymentID}
```
Expected Output

``` json
{
    "ID": "936f27aa-1584-4e35-8508-ac059e7111c2",
    "user_id": "user:100",
    "amount": 1000,
    "currency": "XAF",
    "status": "PROCESSED",
    "initiated_at": "2025-11-04T23:59:00.283625839Z",
    "UpdatedAt": "0001-01-01T00:00:00Z"
}
```

6) ### Test Rate Limiting

Submit multiple payment requests quickly:

``` bash
for i in {1..10}; do
  curl -X POST http://localhost:8000/payments \
    -H "Content-Type: application/json" \
    -d '{"userId":"test","amount":100,"currency":"XAF"}'
done
```

if the limit is exceeded, you wil receive
```json
{
  "error": "Rate limit exceeded. Try again later."
}
```

- ✅ Redis-based rate limiter works
- ✅ The system protects itself under load

7) ### Scale Worker Concurrency (Optional)
``` ini
MAX_WORKERS=5
```

then 

```
make dev
```

You will now see multiple workers running concurrently

```nginx
Worker #1 started
Worker #2 started
Worker #3 started
```

- ✅ Faster processing
- API✅ Distributed background workload

# Project structure #

```bash
payment_service_poc/
│
├── cmd/
│   ├── api/                    # API entrypoint
│   │   └── main.go
│   └── worker/                 # Worker entrypoint
│       └── main.go
│
├── internal/
│   ├── domain/                 # Entities and domain interfaces
│   │   └── payment.go
│   │
│   ├── app/                    # Application layer (use-cases)
│   │   └── payment_service.go
│   │
│   ├── infrastructure/         # Adapters
│   │   ├── redis/              # Redis rate limiter + Redis payment repo
│   │   ├── beanstalk/          # Queue adapter
│   │   ├── orm/                # Repository implementations
│   │   └── worker/             # Worker logic
│   │
│   └── presentation/
│       ├── restapi/handler/    # HTTP handlers
│       └── router/             # Echo routes + HTTP server setup
│
├── Dockerfile                  # Multi-stage build
├── docker-compose.yml          # Entire system definition
├── Makefile                    # Build, dev, run automation
├── .env.example                # Environment variable template
└── README.md
```
