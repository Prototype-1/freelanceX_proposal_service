
# Proposal Service — FreelanceX

## Overview
freelanceX_proposal_service carries proposal related logic where 80% of the operation can only be done by a user with freelancer as his role while signup through the freelanceX_user_service.

## Tech Stack
- Go (Golang)
- gRPC
- MongoDB
- Protocol Buffers
- Kafka

## Setup

### 1. Clone & Navigate
```bash
git clone https://github.com/Prototype-1/freelancex_proposal_service.git
cd freelancex_proposal_service
```

## Install Dependencies

go mod tidy

### Create .env File

PORT=50052
MONGO_URI=mongodb://localhost:27017
MONGO_DB=freelancex_proposals

## Start the Service

go run main.go

### Proto Definitions

    proto/proposal/proposal.proto

Regenerate:

protoc --go_out=. --go-grpc_out=. proto/proposal/proposal.proto

#### Notes

    Templates allow reusable proposal sections.

    Proposals embed content directly for versioning.

## Maintainers

aswin100396@gmail.com