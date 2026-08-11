# SIPA Backend (Sistem Informasi Pelayanan Akademik)

This repository contains the backend service for SIPA, a robust academic and administrative information system. Built with Go (Golang), this RESTful API handles complex business logic, user authentication, and data management to power the SIPA frontend application.

## 🚀 Key Features

As the backend engineer for this project, I architected a system capable of handling various administrative workflows securely and efficiently. Key features include:

* **Robust Role-Based Access Control (RBAC):** Secure API endpoints protected by custom middleware to differentiate access levels for Admin, Kaprodi (Head of Program), Tendik (Educational Staff), and Mahasiswa (Students).
* **Complex Workflow Automation:**
* Automated scheduling and cron jobs (`scheduler.go`) to manage time-sensitive administrative tasks.
* SLA (Service Level Agreement) tracking (`sla.go`) to ensure document processing meets required deadlines.


* **Document & Notification Systems:**
* Automated PDF generation (`pdf_generator.go`) for official academic letters.
* Asynchronous email notifications (`mailer.go`) to keep users updated on their submission statuses.
* Secure file upload handling (`upload.go`) for student document submissions.


* **Comprehensive Logging:** Centralized logging mechanism (`log.go`) to track system activities and user actions for security and auditing.

## 🛠 Tech Stack

* **Language:** [Go (Golang)](https://golang.org/)
* **Database ORM:** GORM (implied by typical Go structures, configured in `db.go`)
* **Authentication:** JWT (JSON Web Tokens) via custom middleware
* **Task Scheduling:** Custom Cron implementations

## 📂 Project Structure

The project follows a clean architecture pattern, separating handlers, models, and utility services:

```text
├── config/           # Database configuration and connection setups (`db.go`)
├── cron/             # Background jobs and automated task schedulers (`scheduler.go`)
├── handlers/         # HTTP request handlers grouped by feature (Auth, Surat, Notification, etc.)
├── models/           # Database schemas and data structures (User, Surat, Log, etc.)
├── utils/            # Shared utilities (Mailer, PDF Generator, SLA calculator)
├── backup.sql        # Database schema/backup file
└── main.go           # Application entry point

```

## 💻 Getting Started

To run this backend service locally, follow these steps:

### Prerequisites

* [Go](https://golang.org/doc/install) (v1.18+ recommended)
* PostgreSQL or MySQL (depending on the `backup.sql` configuration)

### Installation & Setup

1. **Clone the repository:**
```bash
git clone https://github.com/yourusername/be-sipa.git
cd be-sipa

```


2. **Install dependencies:**
```bash
go mod download

```


3. **Database Setup:**
* Create a new database in your local RDBMS.
* Import the provided database structure:
```bash
psql -U username -d database_name < backup.sql
# OR mysql -u username -p database_name < backup.sql

```




4. **Environment Variables:**
* Create a `.env` file in the root directory and configure your database credentials, JWT secrets, and SMTP settings for the mailer.


5. **Run the server:**
```bash
go run main.go

```


The API should now be running (usually on `http://localhost:8080`).
