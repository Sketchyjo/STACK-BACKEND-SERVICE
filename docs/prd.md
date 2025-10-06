
# Product Requirements Document (PRD)

**Product Name:** STACK  
**Prepared By:** Product Management (John)  
**Date:** September 27, 2025  
**Version:** Draft v0.1  

**Summary:**  
STACK is a Web3-native investment app designed for Gen Z users who are underserved by traditional banks and overwhelmed by complex crypto tools. It enables instant wealth-building through a hybrid model: fiat-to-stablecoin on-ramps, seamless investment in stocks/ETFs, and a protective AI CFO. The MVP focuses on onboarding, wallet management, stablecoin deposits, curated baskets, and a simplified investing flow. The long-term vision is to expand into a social, gamified financial hub.  

---

## Goals & Background

### Business Goals
- Drive rapid adoption with 10,000 Monthly Active Users (MAU) within the first 6 months of launch.  
- Establish a recurring revenue stream by converting at least 5% of free users into premium subscribers in year one.  
- Validate market viability by processing $1,000,000 in investments within the first year.  
- Position STACK as the first mover in the Gen Z-native hybrid Web3 + traditional finance investment space.  

### User Goals
- Create a safe, frictionless investment platform that demystifies Web3 while outperforming legacy banking in speed and fairness.  
- Deliver a product experience that matches the expectations of digital-native Gen Z: fast, social, intuitive, and aligned with values like sustainability and fairness.  
- Encourage consistent investing behavior through gamification and protective guidance from an AI CFO.  

### Background Context
The financial market for Gen Z is underserved. Traditional banking alienates younger users with fees, delays, and outdated UX, while Web3 remains a minefield of technical risks and hidden costs. Competitors like Cash App, Robinhood, and Coinbase capture fragments of the opportunity but fail to fully address the **trust + usability gap**. STACK bridges this gap with a hybrid model and an experience that feels designed for Gen Z culture.  

---

## User Personas

### Taylor – The Conscious & Connected Investor (Primary Persona)
- **Age:** 22  
- **Profile:** Digitally native, balances part-time work with side hustles (e.g., Etsy). Ambitious but cautious.  
- **Digital Habits:** Lives on TikTok, Instagram, Reddit, and Discord. Uses Notion/Pinterest for visual planning. Expects fast, engaging, intuitive experiences that feel like “TikTok-meets-Cash App.”  
- **Financial Behaviors:** Keeps most funds in savings + P2P apps (Cash App, Venmo). Dabbles on Robinhood but distrusts its business model. Avoids crypto due to complexity.  
- **Values/Motivations:** Wants financial independence, safety, and alignment with identity (e.g., sustainability, social impact). Goals: travel fund, apartment savings, safety net.  

### Jordan – The Banking-Frustrated Beginner
- **Age:** 21  
- **Pain Point:** Clunky traditional banking, delays (3–5 day ACH transfers), and punitive fees. Feels alienated by outdated systems.  

### Chris – The Crypto-Curious but Overwhelmed
- **Age:** 19  
- **Pain Point:** Intimidated by seed phrases, high gas fees, and irreversible mistakes. Tried but abandoned crypto apps after losing money.  

---

## Functional Requirements

### Core MVP Features
1. **User Onboarding & Managed Wallet**  
   - Simple sign-up with automatic creation of a secure, managed wallet.  
   - No seed phrase complexity; custody abstracted away.  

2. **Stablecoin Deposits**  
   - Support deposits from at least one EVM chain (e.g., Ethereum) and one non-EVM chain (e.g., Solana).  
   - Conversion into stablecoins for immediate use as buying power.  

3. **Investment Flow**  
   - Automatic conversion of stablecoins into fiat-equivalent buying power.  
   - Ability to invest in curated baskets of stocks/ETFs.  
   - Simple portfolio view with performance tracking.  

4. **Curated Investment Baskets**  
   - Launch with 5–10 “expert-curated” investment baskets (e.g., Tech Growth, Sustainability, ETFs).  
   - Designed to simplify decision-making for new investors.  

5. **AI CFO (MVP Version)**  
   - Provides automated weekly performance summaries.  
   - On-demand portfolio analysis to highlight diversification, risk, and potential mistakes.  

6. **Brokerage Integration**  
   - Secure backend integration for trade execution and custody of traditional assets.  

---

### Out of Scope for MVP
- Advanced AI CFO with conversational nudges.  
- Full social/gamified features (profiles, following, leaderboards, copy investing).  
- User-curated baskets, debit card, P2P payments, time-lock investments.  

---

### Post-MVP Roadmap
- **Phase 2:** Full AI CFO, advanced social suite, user-curated baskets.  
- **1–2 Years:** Expansion into debit card, P2P payments, business accounts, startup launchpad.  

---

## Success Metrics

### Business Objectives
- **User Acquisition:** 10,000 Monthly Active Users (MAU) within 6 months post-launch.  
- **Monetization:** 5% conversion from free users to premium tier in Year 1.  
- **Validation:** $1,000,000 in processed investment volume in Year 1.  

### User Success Metrics
- **Empowerment:** Users feel more in control of their financial future (via surveys).  
- **Confidence:** Users feel safe and protected (via NPS and retention).  
- **Habit Formation:** % of users with recurring investments increases steadily.  

### Key Performance Indicators (KPIs)
- **Engagement:** Daily Active Users (DAU), Monthly Active Users (MAU).  
- **Retention:** Week 1, Month 1, Month 3 retention rates.  
- **Conversion:** Sign-up → Funded Account rate; Free → Premium rate.  
- **Financial:** Total Assets Under Management (AUM).  

---

## Technical Considerations

### Target Platforms
- Native mobile applications for **iOS and Android**.  

### Technology Stack
- **Frontend:** React Native (cross-platform mobile framework).  
- **Backend:** Node.js (NestJS framework).  
- **Database:** PostgreSQL.  

### Infrastructure & Integrations
- **0G** for storage and AI capabilities.  
- **Circle** for stablecoins and wallet infrastructure.  
- Wallet management tool to abstract custody and simplify UX.  
- Brokerage partner integration for execution and custody of traditional assets.  

### Architecture Strategy
- Initial approach: **Modular Monolith** within a monorepo for faster MVP delivery.  
- Long-term: Designed to evolve into a more distributed architecture as user base scales.  

### Constraints
- **Timeline:** 3-week deadline for MVP cycle.  
- **Dependencies:** Reliance on 0G and Circle APIs.  

### Assumptions
- A suitable brokerage partner can be secured.  
- Regulatory compliance model is viable.  
- Third-party APIs (0G, Circle, brokerage) are stable and cost-effective at scale.  

---

## Risks & Open Questions

### Key Risks
1. **Partner Risk (Brokerage):**  
   - Dependency on securing a reliable brokerage partner for trade execution and custody.  
   - Risk of delays or unfavorable terms impacting launch timeline.  

2. **Regulatory Risk:**  
   - Hybrid Web3 + traditional model may face compliance challenges in multiple jurisdictions.  
   - Potential impact on stablecoin usage, custody models, and cross-border flows.  

3. **Execution Risk:**  
   - Tight 3-week MVP timeline creates potential for scope creep or missed deadlines.  
   - Reliance on multiple third-party APIs increases fragility.  

4. **Technical Risk (APIs):**  
   - Performance and reliability of 0G, Circle, and brokerage APIs are critical to user experience.  
   - Unknown at-scale costs of API usage.  

5. **Market Risk (User Trust):**  
   - Gen Z skepticism toward both banks and crypto means adoption may be slower than expected.  
   - Building trust will require careful onboarding and consistent reliability.  

---

### Open Questions
- Which brokerage partner will be selected and how will integration work?  
- Has a full legal/regulatory review been conducted?  
- What are the true at-scale costs of Circle, 0G, and brokerage APIs?  
- How will fraud prevention and compliance (KYC/AML) be handled in MVP vs post-MVP?  
