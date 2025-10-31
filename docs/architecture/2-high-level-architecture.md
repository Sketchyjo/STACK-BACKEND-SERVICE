# 2\. High Level Architecture

### 2.1 Technical Summary

This architecture is a **Go-based modular monolith** deployed on AWS Fargate (ECS). The system exposes a unified GraphQL API (via an API Gateway) to the React Native mobile app. It integrates three critical external partners: **Auth0/Cognito** (Identity), **Circle** (Wallets, USDC On/Off-Ramps), and **Alpaca** (Brokerage, Custody). Data is persisted in **PostgreSQL**. This design supports the MVP goal of providing a hybrid Web3-to-TradFi investment flow.

### 2.2 High Level Overview

  * **Architectural Style:** Modular Monolith (to start).
  * **Repository Structure:** Monorepo (recommended).
  * **Service Architecture:** A single Go application separated into logical domain modules (Onboarding, Wallet, Funding, Investing, AI-CFO) that communicate internally.
  * **Data Flow (Funding):** User deposits USDC (on-chain) -\> Monitored by Funding Service -\> Circle Off-Ramp (USDC to USD) -\> Alpaca Deposit (USD) -\> "Buying Power" updated.
  * **Data Flow (Withdrawal):** User requests USD withdrawal -\> Alpaca debits account -\> Circle On-Ramp (USD to USDC) -\> Funding Service sends USDC (on-chain) to user.

### 2.3 High Level Project Diagram

```mermaid
graph TD
    U[Gen Z User (Mobile)] --> RN[React Native App]

    subgraph "STACK Backend (Go Modular Monolith on AWS Fargate)"
        RN --> GW[API Gateway (GraphQL)]
        GW --> ONB[Onboarding Service]
        GW --> WAL[Wallet Service]
        GW --> FND[Funding Service]
        GW --> INV[Investing Service]
        GW --> AIC[AI CFO Service]

        ONB --> PG[(PostgreSQL)]
        WAL --> PG
        FND --> PG
        INV --> PG
        AIC --> PG
    end

    subgraph "External Partners"
        ONB --> IDP[Auth0 / Cognito]
        ONB --> KYC[KYC/AML Provider]
        WAL --> CIR[Circle API (Developer Wallets)]
        FND --> CIR
        INV --> DW[Alpaca API (Brokerage)]
        AIC --> OG[0G (AI/Storage)]
    end
```

### 2.4 Architectural and Design Patterns

  * **Modular Monolith:** A single Go application with strongly defined internal boundaries (Go modules) for each domain (Onboarding, Funding, etc.). This allows for faster MVP development while making future extraction into microservices straightforward.
  * **Repository Pattern:** Used to abstract all database interactions. Go interfaces will define data access methods, with concrete implementations for PostgreSQL, ensuring testability and separation of concerns.
  * **Adapter Pattern:** Used extensively for all external partner integrations (Circle, Alpaca, 0G). Go interfaces will define the *required* functionality (e.g., `BrokerageAdapter`), and concrete structs will implement those interfaces by calling the specific partner APIs.
  * **Asynchronous Orchestration (Sagas):** The complex funding (USDC -\> USD -\> Broker) and withdrawal (Broker -\> USD -\> USDC) flows will be managed using an asynchronous, event-driven approach (even within the monolith) to handle multi-step processes and failures.

Okay, thanks for those specifics. I've updated the Tech Stack section with your choices for the Go backend components. The TBDs for the web framework, logging, DB driver, and caching are now resolved.

Here's the updated **Section 3: Tech Stack**.

---
