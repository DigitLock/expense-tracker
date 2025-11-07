# Expense Tracker - Business Requirements Document

## 1. Document Overview
- **Document Owner:** Igor Kudinov - Business & System Analyst
- **Date / Version:** November 1, 2025 / v1.0
- **Related Initiatives / Projects:** Personal Finance Management System, BSA Portfolio Project
- **Main Stakeholders:** Project Owner (BSA/Developer), Family Members (Primary Users)

---
<br />


## 2. Executive Summary
**Expense Tracker** is a personal finance management system designed to provide comprehensive tracking and analysis of family financial flows. The system addresses the critical need for unified financial visibility that enables informed decision-making regarding household budget management.

The primary business problem being solved is the lack of a centralized system for monitoring family income, expenses, and savings, which prevents effective financial planning and goal achievement. Currently, the family relies on mental estimates and basic understanding of cash flows without systematic tracking or analysis capabilities.

The solution will deliver immediate value through improved financial transparency, enabling better spending decisions, systematic savings accumulation, and long-term financial planning. Additionally, this project serves as a comprehensive portfolio piece demonstrating full-cycle business and system analysis capabilities, from requirements gathering through technical implementation.

---
<br />


## 3. Business Context
### General Business Goal
Establish systematic financial management practices for family budget planning and achieve sustainable savings accumulation through data-driven decision making.
<br /><br /> 

### Problem or Opportunity Being Addressed
#### Core Problem:
There is no unified system for tracking family financial flows that meets specific user experience requirements, which prevents making informed financial decisions and leads to inability to plan major expenses and form systematic savings.

#### Existing Solutions Gap:
While market offers various personal finance applications, none provide the desired level of interface personalization and user experience tailored to specific family workflow needs. Available solutions either lack required functionality or present overly complex interfaces that don't align with preferred interaction patterns.

#### Specific Pain Points:
- Lack of visibility into spending patterns and money allocation
- No systematic approach to savings and financial goal tracking
- Inability to analyze historical financial data for future planning
- Manual processes leading to incomplete or inconsistent financial records
- No coordination mechanism for family members' financial activities
<br /> 

### Market or Regulatory Factors
- Growing trend towards personal financial management and budgeting apps
- Increased focus on financial literacy and planning post-economic uncertainties
- Multi-currency environment (RSD/EUR) requiring currency conversion capabilities
- No specific regulatory requirements for personal finance tracking
<br /> 

### Impact on Existing Processes, Teams, or Systems
- **Current Process:** Ad-hoc mental tracking with approximate understanding of cash flows
- **Process Change:** Transition to systematic digital recording and analysis
- **Team Impact:** Family members will need to adopt new data entry habits
- **System Integration:** Initially standalone system with potential future integrations (SMS reading, bank APIs)  

---
<br />


## 4. Objectives and Goals
| Goal | KPI / Metric | Expected Benefit | Priority |
|------|---------------|------------------|-----------|
| Launch MVP with core functionality | MVP deployed and operational within 1 month | Immediate start of systematic financial tracking | High |
| Achieve daily family usage | System used by family members daily for 90% of financial transactions | Complete visibility into family financial flows | High |
| Implement administrative capabilities | Reference data management and system administration interfaces operational | Self-service system customization and maintenance | Medium |
| Establish baseline financial data | 3 months of complete financial data collected | Foundation for informed financial decision-making | Medium |
| Develop comprehensive feature set | Multi-currency support, account management, and categorization functionality | Complete personal finance management solution | Low |

---
<br />


## 5. Stakeholders
| Name | Department / Role | Responsibility / Interest | Influence |
|------|--------------------|---------------------------|------------|
| Igor Kudinov | Project Owner / BSA / Developer | Project leadership, requirements definition, system design and implementation | High |
| Wife | Primary User / Family Financial Manager | Daily system usage, data entry, financial decision making | High |
| Igor Kudinov (User) | Primary User / Family Financial Manager | Daily system usage, data entry, financial decision making | High |
| Extended Family Members | Potential Future Users | Possible system adoption after successful family implementation | Low |
| Financial Analyst (Future) | External User / Consultant | Limited access for financial analysis and reporting (post-MVP) | Medium |

---
<br />


## 6. Current State (As-Is)
### Current Financial Management Process
#### Primary Tools Used:
- Banking applications for transaction history review
- SMS notifications for expense alerts and balance updates
- Occasional Excel/Google Sheets attempts (discontinued due to inconvenience)
- Sporadic notes in phone (often lost among other information)
- Paper-based expense planning for major financial decisions

#### Current Workflow:
1. **Income Tracking:** Basic awareness through bank account deposits
2. **Expense Monitoring:** Reactive review through banking apps and SMS alerts
3. **Planning:** Ad-hoc paper calculations for major expenses and savings allocation
4. **Record Keeping:** No systematic record maintenance
5. **Analysis:** Mental estimates based on approximate understanding of cash flows
<br />

### Key Pain Points and Inefficiencies
#### Data Fragmentation:
- Financial information scattered across multiple sources (bank apps, SMS, notes)
- No centralized view of complete financial picture
- Important financial notes lost among other unrelated information

#### Process Inefficiencies:
- Reactive rather than proactive financial management
- Manual paper calculations for budgeting decisions
- Multiple failed attempts to use spreadsheet solutions due to poor user experience
- No systematic categorization of expenses

#### Analysis Limitations:
- Inability to track spending patterns over time
- No visibility into category-based expense analysis
- Lack of structured approach to savings and financial goal tracking
- Insufficient data for informed financial decision making

#### Coordination Challenges:
- No shared system for family financial coordination
- Individual awareness without consolidated family budget view

---
<br />


## 7. Future State (To-Be)
### Key Business Changes Expected
#### Centralized Financial Management:
- Single system consolidating all family financial data (income, expenses, savings)
- Real-time visibility into financial status across multiple accounts and currencies
- Systematic categorization and tracking of all financial transactions

#### Proactive Financial Planning:
- Data-driven decision making based on historical spending patterns
- Structured approach to savings allocation and financial goal tracking
- Automated calculations replacing manual paper-based budgeting

#### Enhanced Family Coordination:
- Shared access to family financial data for both spouses
- Synchronized financial tracking eliminating information gaps
- Collaborative budget management and financial decision making
<br />

### Improvements or Efficiencies Gained
#### Data Consolidation:
- Replace fragmented data sources (bank apps, SMS, notes) with unified system
- Eliminate lost financial information through systematic digital storage
- Provide comprehensive financial reporting and analysis capabilities

#### Process Automation:
- Automated transaction categorization reducing manual data entry
- Multi-currency support with automatic conversion for accurate tracking
- Digital record keeping eliminating paper-based calculations

#### User Experience:
- Personalized interface designed for family workflow preferences
- Intuitive data entry and reporting interfaces
- Mobile and web accessibility for convenient usage
<br />

### Dependencies or Integrations Required
#### MVP Phase:
- Manual data entry processes (no external integrations)
- Backend API with PostgreSQL database for centralized data storage
- Web and mobile clients connecting to backend services
- Basic reporting and analytics functionality
- Currency exchange rate API integration for daily RSD/EUR rate updates

#### Future Enhancements:
- Local caching in browser and mobile app with server synchronization (post-MVP offline mode)
- Standalone mobile app with local database and cloud backup (Google Drive)
- SMS parsing integration for automated transaction capture
- Advanced currency exchange rate API integration for real-time and historical conversions
- Potential banking API integrations for automatic transaction import  

---
<br />


## 8. Business Requirements
| ID | Requirement | Description | Priority | Acceptance Criteria |
|----|--------------|-------------|-----------|----------------------|
| BR-1 | Income Management | The system must allow the business to record, categorize, and track all family income sources with multi-currency support | High | Users can add income entries with categories, amounts, currencies, and dates |
| BR-2 | Expense Tracking | The system must allow the business to record, categorize, and track all family expenses across different accounts and categories | High | Users can add expense entries with customizable categories, accounts, and transaction details |
| BR-3 | Savings Management | The system must allow the business to track and categorize savings across cash and non-cash accounts with goal-oriented allocation | High | Users can allocate funds to different savings categories and track progress |
| BR-4 | Account Management | The system must allow the business to maintain multiple accounts (cash, non-cash, savings) with balance tracking | High | System maintains accurate balances across different account types |
| BR-5 | Multi-Currency Support | The system must allow the business to operate with multiple currencies (RSD, EUR) with conversion capabilities | Medium | Users can select base currency and view converted amounts for all transactions using up-to-date exchange rates retrieved from an external provider (at least daily) |
| BR-6 | User Management | The system must allow the business to support multiple family members with shared access to financial data | Low | Multiple users can access and modify family financial data |
| BR-7 | Basic Reporting | The system must allow the business to generate basic financial reports by categories and time periods | Low | System generates income, expense, and savings reports with filtering options |
| BR-8 | Data Categorization | The system must allow the business to customize basic category CRUD for income and expense categories according to family preferences | Low | Users can create, modify, and delete custom categories and subcategories |
| BR-9 | Reference Data Management | The system must allow the business to manage system reference data (currencies, accounts, categories) | Low | Administrators can maintain system reference data through dedicated interfaces |
| BR-10 | Audit Trail | The system must allow the business to track all financial data modifications for accountability | Low | System logs all data changes with user and timestamp information |
| BR-11 | Data Import/Export | [POST-MVP] The system must allow the business to import financial data from external sources and export data for backup or analysis | Low | Users can import/export transactions in standard formats (CSV, JSON) with data validation |
| BR-12 | Goals and Savings Targets Management | [POST-MVP] The system must allow the business to set, track, and visualize progress towards financial goals and savings targets | Low | Users can create savings goals with target amounts, deadlines, and visual progress tracking |

---
<br />


## 9. Functional and Non-Functional Needs
### 9.1 Functional Needs
#### Core Financial Operations:
- Transaction entry interfaces (income, expense operations)
- Account balance calculations and real-time updates
- Category and subcategory management with user customization
- Multi-currency transaction processing with base currency conversion

#### Data Management:
- User authentication and session management for family members
- CRUD operations for all financial entities (transactions, accounts, categories)
- Data validation and business rule enforcement
- Import/export capabilities for financial data

#### Reporting and Analytics:
- Financial summary dashboards with key metrics
- Category-based expense analysis and trends
- Time-period filtering and comparison reports
- Savings progress tracking and goal monitoring

#### Administrative Functions:
- Reference data management interfaces (currencies, account types)
- User management and access control (post-MVP)
- System configuration and customization options

#### Notifications and Alerts (post-MVP):
- Data entry reminders and user engagement notifications
- Goal achievement and milestone alerts
- Budget threshold warnings and financial insights
- System maintenance and update notifications
<br />

### 9.2 Non-Functional Needs
#### Performance Expectations:
- Response time: <2 seconds for standard operations
- Database queries: <500ms for reporting functions
- Support for concurrent family users (2-3 simultaneous sessions)
- Mobile app responsiveness optimized for financial data entry

#### Security and Compliance:
> Note: This section includes both MVP and post-MVP expectations.  
> Requirements explicitly marked as [POST-MVP] are not mandatory for the initial release.
- Secure user authentication and session management
- Data encryption for sensitive financial information in transit and at rest on the server
- Basic handling of network errors during data entry (no silent data loss, clear user feedback)
- Input validation and SQL injection protection
- Regular data backup and recovery procedures
- [POST-MVP] Local data encryption for offline mode and cached data on client devices

#### Scalability and Operational Constraints:
- PostgreSQL database optimization for financial transaction volumes
- API rate limiting and resource management
- Logging and monitoring for system health tracking
- Graceful error handling and user feedback mechanisms

#### Usability and Accessibility:
- Intuitive interface design optimized for family workflow
- Mobile-first responsive design for cross-device usage
- Multi-language support preparation (English/Russian)
- Accessibility standards compliance for financial applications  

---
<br />


## 10. Risks and Dependencies
| Risk | Impact | Probability | Mitigation |
|------|---------|-------------|-------------|
| Low user adoption by family members | High - system won't deliver expected value without daily usage | Medium | Involve family in requirements gathering; focus on intuitive UX design; provide training and support |
| Technical complexity exceeding timeline | Medium - delayed MVP delivery | Medium | Start with simplified MVP; iterative development approach; focus on core features first |
| Data loss or corruption | High - loss of valuable financial history | Low | Implement regular backups; database transaction integrity; thorough testing procedures |
| Currency conversion accuracy issues | Medium - incorrect financial reporting | Low | Use reliable exchange rate APIs; implement validation checks; manual override capabilities |
| Scope creep during development | Medium - extended timeline and complexity | High | Strict MVP definition; change control process; regular stakeholder alignment |
| Mobile app development challenges | Medium - limited platform coverage | Medium | Start with web version; use proven mobile frameworks (Flutter); cross-platform approach |
| External API dependencies | Medium - disruption of currency conversion and future integrations | Medium | Use multiple exchange rate providers; implement fallback mechanisms; monitor API availability |

---
<br />


## 11. Expected Benefits and ROI
### Qualitative Benefits:
- **Improved Financial Transparency:** Complete visibility into family financial flows and spending patterns
- **Enhanced Decision Making:** Data-driven approach to financial planning and major purchase decisions  
- **Reduced Financial Stress:** Systematic approach eliminates uncertainty about budget status and savings progress
- **Better Family Coordination:** Shared financial system improves communication and alignment on financial goals
- **Professional Portfolio Value:** Comprehensive BSA documentation demonstrating full project lifecycle capabilities

### Quantitative Benefits:
- **Time Savings:** Estimated 2-3 hours per month saved on manual financial calculations and record keeping
- **Improved Savings Rate:** Target 10-15% improvement in savings allocation through better expense visibility
- **Reduced Financial Errors:** Elimination of manual calculation errors and missed expense tracking
- **ROI Estimation:** ROI = (Time Saved × $25/hour) / Development Effort ≈ 300% over 12 months

### Before vs After Comparison:
| Metric | Current State (Before) | Target State (After) | Improvement |
|--------|------------------------|---------------------|-------------|
| Time spent on financial tracking | 3-4 hours/month | 1 hour/month | 67% reduction |
| Financial data accuracy | ~70% (estimates) | 95%+ (systematic) | 25+ percentage points |
| Savings allocation visibility | Limited/unclear | Complete transparency | 100% improvement |
| Family financial coordination | Ad-hoc discussions | Shared data system | Structured process |
| Financial decision confidence | Low (guesswork) | High (data-driven) | Qualitative improvement |

---
<br />


## 12. Approval and Sign-Off
| Name | Role | Date | Signature |
|------|------|------|------------|
| Igor Kudinov | Project Owner / BSA | November 5, 2025 |  |
| Igor Kudinov | Primary User Representative | November 5, 2025 |  |
| Wife | Primary User Representative | November 5, 2025 |  |

---
