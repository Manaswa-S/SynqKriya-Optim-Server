# SynqKriya – Optimization Server
### Backend orchestration layer for an ML-based smart traffic management system.
---

This repository contains **only the Optimization Server**, responsible for data ingestion, coordination, metric computation, and persistence.
<br>
The system is designed as a **decoupled, event-driven pipeline** that connects live traffic cameras with ML inference and decision-making components.

---

> This is a systems and data-pipeline project, not an ML model repository.

---

### System Overview

The complete SynqKriya system consists of three major parts:

1. **Optimization Server** (this repository)
2. **Detection & Tracking ML Model** (YOLOv8-based, external)
3. **Decision-Making ML Model** (external)

This repo implements the **Optimization Server**, which acts as the backbone connecting:
- live RTSP camera feeds
- ML inference components
- downstream storage and post-processing

---

### Optimization Server Architecture

The Optimization Server is split into **three independent Go microservices**, communicating exclusively through **Redis Pub/Sub**.
Each stage is isolated, asynchronous, and independently scalable.

#### Architecture Diagram 
<img width="800px" height="500px" src="https://github.com/Manaswa-S/SynqKriya-Optim-Server/blob/main/arch.jpg" >

---

### Microservices Breakdown

### 1. Pre-Optim Service: Video Ingestion & Preprocessing

Responsible for handling raw traffic camera feeds and preparing data for ML consumption.

**Behavior**
- Accepts one or more RTSP camera stream URLs
- Spawns **one goroutine per camera**
- A central scheduler manages all camera routines
- All video capture, frame sampling, clip assembly, and compression are handled using **FFmpeg**.

For each camera:
- Captures video feed for **X seconds**
- Samples **1 frame out of every Y frames**
- Combines sampled frames into a short clip
- Compresses the clip
- Uploads the clip to **AWS S3**
- Generates a **signed S3 URL**
- Demo Output: <a href="https://github.com/Manaswa-S/SynqKriya-Optim-Server/blob/main/outputs/preoptim-detection.json"> Link </a>

This metadata is published to a **Redis Pub/Sub channel**.

**Design Intent**
- Decouple heavy video ingestion from ML inference
- Avoid inference backpressure caused by live streams

---

### 2. Detection & Tracking Model (External)
_Not part of this repository._

- Subscribes to the Pre-Optim Redis channel
- Downloads the video clip using the signed S3 URL
- Performs detection and tracking using:
  - YOLOv8
  - additional tracking logic
- Publishes structured inference results to another Redis channel

- Demo Output: <a href="https://github.com/Manaswa-S/SynqKriya-Optim-Server/blob/main/outputs/detection-midoptim.json"> Link </a>

---

### 3. Mid-Optim Service  
**Metric Computation Layer**

Acts as a semantic bridge between raw ML inference and decision-making.

**Behavior**
- Subscribes to detection model output channel
- Converts raw inference into structured traffic metrics

Computed metrics include:
- vehicle count
- traffic density
- congestion indicators
- flow-related metrics

- Demo Output: <a href="https://github.com/Manaswa-S/SynqKriya-Optim-Server/blob/main/outputs/midoptim-decision.json"> Link </a>

The computed metrics are published to a separate Redis channel.

**Design Intent**
- Keep perception (ML inference) and reasoning (decision-making) cleanly separated
- Normalize and structure data before decision logic

---

### 4. Decision-Making Model (External)
_Not part of this repository._

- Subscribes to the metrics channel
- Uses ML / optimization logic to decide traffic signal changes
- Outputs **only delta decisions**
- Does **not** emit full system state
- Only intersections / nodes that require changes are published

- Demo Output: <a href="https://github.com/Manaswa-S/SynqKriya-Optim-Server/blob/main/outputs/decision-postoptim.json"> Link </a>

Decision deltas are pushed to another Redis channel.

---

### 5. Post-Optim Service  
**Persistence & Post-Processing**

Handles final decisions, applies policies and makes them durable.

**Behavior**
- Subscribes to decision delta channel
- Stores decisions in **PostgreSQL**
- Maintains historical state and auditability
- Can trigger additional post-decision workflows

**Design Intent**
- Ensure durability, traceability, and correctness of traffic control decisions

---

## Tech Stack

- **Language:** Go
- **Concurrency:** Goroutines, centralized schedulers
- **Messaging:** Redis Pub/Sub
- **Storage:**
  - AWS S3 (video clips)
  - PostgreSQL (decision data)
- **Media Processing:** FFmpeg
- **Input Streams:** RTSP camera feeds
- **Architecture Style:**
  - Microservices
  - Event-driven
  - Loosely coupled pipeline

---

## Design Characteristics

- Strong stage-wise decoupling using Redis
- Fully asynchronous processing
- Independent scalability per service
- Delta-based decision propagation to reduce noise and recomputation
- Built as a production-oriented backend system, not an academic ML demo

---

## Scope & Positioning

This repository focuses on:
- systems engineering
- data pipelines
- service coordination
- real-time ingestion and persistence

It intentionally does **not** include:
- ML model training
- ML inference logic
- traffic optimization algorithms

Those components interact with this system externally.

---

## Disclaimer

This project was developed as part of A smart traffic management system.
It is not deployed in production and has not undergone real-world traffic validation.
