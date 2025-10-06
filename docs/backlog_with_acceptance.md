
# 📋 Story Backlog for STACK MVP (with Acceptance Criteria)

---

## Epic 1: Onboarding & Wallet Management

1. **User Registration**  
_As a new user, I want to sign up with email/phone, so that I can quickly create an account._  
**Acceptance Criteria:**  
- User can sign up with valid email or phone number.  
- Invalid input shows error message.  
- Successful sign-up redirects to onboarding flow.  

2. **Managed Wallet Creation**  
_As a new user, I want a secure wallet to be auto-generated, so that I don’t have to manage seed phrases._  
**Acceptance Criteria:**  
- Wallet is automatically created upon sign-up.  
- User is not exposed to seed phrase or private key.  
- Wallet is linked to user account securely.  

3. **Security Abstraction**  
_As a cautious investor, I want custody handled for me, so that I feel safe without technical complexity._  
**Acceptance Criteria:**  
- Custody managed by trusted provider.  
- No manual wallet management required.  
- Security audit log exists for each wallet.  

4. **Onboarding Completion Tracking**  
_As a product team, I want to measure drop-off during onboarding, so that I can optimize the funnel._  
**Acceptance Criteria:**  
- Each onboarding step logs completion metrics.  
- Analytics can show funnel conversion rates.  

---

## Epic 2: Stablecoin Deposits

1. **Deposit via Ethereum (EVM)**  
_As a user, I want to deposit stablecoins from Ethereum, so that I can fund my account easily._  
**Acceptance Criteria:**  
- User can input deposit address.  
- Deposits are detected and credited within 1 min.  
- Errors are logged and displayed if transaction fails.  

2. **Deposit via Solana (non-EVM)**  
_As a user, I want to deposit stablecoins from Solana, so that I have multiple funding options._  
**Acceptance Criteria:**  
- Solana address generated for each user.  
- Deposits confirm within 1 min.  
- Balance updates automatically.  

3. **Automatic Conversion to Buying Power**  
_As a user, I want my deposits converted into buying power, so that I can invest immediately._  
**Acceptance Criteria:**  
- Deposited stablecoins convert automatically into fiat-equivalent balance.  
- Conversion rate matches published exchange rates.  
- User sees updated balance instantly.  

4. **Deposit Confirmation**  
_As a user, I want to see instant confirmation of my deposit, so that I know my funds are safe._  
**Acceptance Criteria:**  
- Notification appears after successful deposit.  
- Confirmation includes amount + timestamp.  

---

## Epic 3: Investment Flow

1. **Buy into Basket**  
_As a user, I want to invest in a curated basket, so that I can diversify without deep research._  
**Acceptance Criteria:**  
- User can choose basket and amount to invest.  
- Investment executes successfully.  
- Portfolio reflects new holdings.  

2. **Sell Basket Holdings**  
_As a user, I want to sell out of a basket, so that I can take profits or cut losses._  
**Acceptance Criteria:**  
- User can select basket to sell.  
- Sale amount updates buying power.  
- Transaction receipt provided.  

3. **Portfolio Overview**  
_As a user, I want to see my portfolio balance, so that I know how my money is allocated._  
**Acceptance Criteria:**  
- Dashboard shows holdings by basket.  
- Values update in near real time.  

4. **Performance Tracking**  
_As a user, I want to see basket performance over time, so that I can understand my results._  
**Acceptance Criteria:**  
- Each basket shows % gain/loss.  
- Historical performance charts available.  

---

## Epic 4: Curated Baskets

1. **Expert-Defined Basket List**  
_As a new investor, I want to choose from 5–10 baskets, so that I have a safe starting point._  
**Acceptance Criteria:**  
- App lists at least 5 baskets.  
- Each basket has name + description.  

2. **Basket Detail Page**  
_As a user, I want to view details of each basket (composition, sectors), so that I understand what I’m buying._  
**Acceptance Criteria:**  
- Each basket shows stock/ETF breakdown.  
- Sector distribution displayed.  

3. **Basket Recommendation**  
_As a cautious investor, I want the app to suggest a basket based on my preferences, so that I feel supported._  
**Acceptance Criteria:**  
- User can answer short preference survey.  
- Recommendation algorithm suggests 1–2 baskets.  

---

## Epic 5: AI CFO (MVP Version)

1. **Weekly Summary Report**  
_As a user, I want to receive a weekly summary, so that I understand my portfolio’s performance._  
**Acceptance Criteria:**  
- Email/app notification with weekly summary.  
- Report includes gains/losses, diversification status.  

2. **On-Demand Portfolio Analysis**  
_As a user, I want to request an instant analysis, so that I can spot risks or diversification gaps._  
**Acceptance Criteria:**  
- User can click “Analyze Portfolio.”  
- AI CFO responds with diversification insights.  

3. **AI CFO Dashboard Widget**  
_As a user, I want a simple AI CFO insights widget, so that I get value without reading long reports._  
**Acceptance Criteria:**  
- Widget displays 2–3 insights.  
- Updates weekly.  

---

## Epic 6: Brokerage Integration

1. **Trade Execution API Call**  
_As a system, I want to execute trades through the brokerage partner, so that users’ investments are processed securely._  
**Acceptance Criteria:**  
- API call executes trade successfully.  
- Error handling covers retries + failures.  

2. **Custody Confirmation**  
_As a system, I want to confirm asset custody, so that user holdings are verifiable._  
**Acceptance Criteria:**  
- Brokerage returns custody confirmation.  
- Portfolio updates reflect custody data.  

3. **Error Handling for Failed Trades**  
_As a user, I want to be notified if a trade fails, so that I can take corrective action._  
**Acceptance Criteria:**  
- Failed trades trigger push notification.  
- Error message provides guidance.  
