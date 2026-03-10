# 🚀 AWS S3-to-SES Lambda Processor

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=for-the-badge&logo=go&logoColor=white)](https://go.dev/)
[![AWS Lambda](https://img.shields.io/badge/AWS-Lambda-FF9900?style=for-the-badge&logo=amazon-aws&logoColor=white)](https://aws.amazon.com/lambda/)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg?style=for-the-badge)](https://opensource.org/licenses/MIT)

A high-performance AWS Lambda function written in Go that automatically notifies customers via SES when a new contract is uploaded to S3.

## 📖 Overview

This service automates the notification workflow for contract processing. When a file is uploaded to a specific S3 bucket, this Lambda extracts the customer's email from the object's metadata and sends a confirmation email using AWS SES v2.

### Workflow
1.  **S3 Upload**: A contract is uploaded to `distributed-jobs-platform-contracts`.
2.  **Event Trigger**: S3 triggers this Lambda function.
3.  **Metadata Extraction**: Lambda fetches the `client-email` from S3 user metadata.
4.  **Notification**: An email is sent via SES to the extracted address.

---

## 🛠️ Technical Stack

- **Language**: Go (Golang)
- **SDK**: AWS SDK for Go v2
- **Infrastructure**: AWS Lambda (Amazon Linux 2023)
- **Services**: S3 (Trigger), SES v2 (Email)

---

## 🚀 Deployment Guide

### 1. Build & Package
Run these commands in your terminal to compile the binary and create the deployment package.

```bash
# Clean and sync dependencies
go mod tidy

# Compile for AWS Lambda (Linux x64)
# Note: 'bootstrap' is required for Amazon Linux 2023 runtime
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap main.go

# Create the deployment zip
# Windows (PowerShell):
powershell -Command "Compress-Archive -Path bootstrap -DestinationPath function.zip -Force"

# Linux/macOS/Git Bash:
zip function.zip bootstrap
```

### 2. AWS Lambda Setup
Configure the following settings in the AWS Console:

| Setting | Value |
| :--- | :--- |
| **Runtime** | Amazon Linux 2023 |
| **Architecture** | x86_64 |
| **Handler** | `bootstrap` |
| **Code** | Upload `function.zip` |

### 3. Permissions (IAM Policy)
Ensure the Lambda role has the following inline policy:

```json
{
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "S3Access",
            "Effect": "Allow",
            "Action": ["s3:GetObject", "s3:HeadObject"],
            "Resource": "arn:aws:s3:::distributed-jobs-platform-contracts/*"
        },
        {
            "Sid": "SESAccess",
            "Effect": "Allow",
            "Action": [
                "ses:SendEmail",
                "ses:SendRawEmail"
            ],
            "Resource": "*"
        }
    ]
}
```

---

## 🔔 S3 Trigger Configuration

1.  Navigate to your Lambda function in the AWS Console.
2.  Click **Add Trigger** and select **S3**.
3.  **Bucket**: Select `distributed-jobs-platform-contracts`.
4.  **Event type**: `All object create events` (or PUT/POST).
5.  **Acknowledge** the recursive invocation warning (safe as we don't write back to the same bucket).

---

## 📝 Usage Note: S3 Metadata

For the Lambda to work correctly, files uploaded to S3 **MUST** include the following user metadata:

| Key | Value | Description |
| :--- | :--- | :--- |
| `client-email` | `user@example.com` | The destination for the notification. |

---

## 📧 SES Configuration

> [!IMPORTANT]
> Ensure the sender email (`castillofranciscodaniel@gmail.com`) is **verified** in your AWS SES console. If you are in the SES Sandbox, the recipient email must also be verified.

---

## 📜 License

Distributed under the MIT License. See `LICENSE` for more information.
