# SIPA Backend (Sistem Informasi Pelayanan Akademik)

This repository contains the backend service for SIPA, a robust academic and administrative information system[cite: 3]. Built with Go (Golang), this RESTful API handles complex business logic, user authentication, and data management to power the SIPA frontend application[cite: 3].

## 🚀 Key Features

As the backend engineer for this project, I architected a system capable of handling various administrative workflows securely and efficiently[cite: 3]. Key features include:

- **Robust Role-Based Access Control (RBAC):** Secure API endpoints protected by custom middleware to differentiate access levels for Admin, Kaprodi (Head of Program), Tendik (Educational Staff), and Mahasiswa (Students)[cite: 3].
- **Complex Workflow Automation:** 
  - Automated scheduling and cron jobs (`scheduler.go`) to manage time-sensitive administrative tasks[cite: 3].
  - SLA (Service Level Agreement) tracking (`sla.go`) to ensure document processing meets required deadlines[cite: 3].
- **Document & Notification Systems:**
  - Automated PDF generation (`pdf_generator.go`) for official academic letters[cite: 3].
  - Asynchronous email notifications (`mailer.go`) to keep users updated on their submission statuses[cite: 3].
  - Secure file upload handling (`upload.go`) for student document submissions[cite: 3].
- **Comprehensive Logging:** Centralized logging mechanism (`log.go`) to track system activities and user actions for security and auditing[cite: 3].

## 🛠 Tech Stack

- **Language:** [Go (Golang)](https://golang.org/)[cite: 3]
- **Database ORM:** GORM (implied by typical Go structures, assuming `db.go` handles this)[cite: 3]
- **Authentication:** JWT (JSON Web Tokens) via custom middleware[cite: 3]
- **Task Scheduling:** Custom Cron implementations[cite: 3]

## 📂 Project Structure

The project follows a clean architecture pattern, separating handlers, models, and utility services[cite: 3]:

```text
├── config/           # Database configuration and connection setups (`db.go`)[cite: 3]
├── cron/             # Background jobs and automated task schedulers (`scheduler.go`)[cite: 3]
├── handlers/         # HTTP request handlers grouped by feature (Auth, Surat, Notification, etc.)[cite: 3]
├── models/           # Database schemas and data structures (User, Surat, Log, etc.)[cite: 3]
├── utils/            # Shared utilities (Mailer, PDF Generator, SLA calculator)[cite: 3]
├── backup.sql        # Database schema/backup file[cite: 3]
└── main.go           # Application entry point[cite: 3]
