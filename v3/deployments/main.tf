provider "google" {
  project = var.project_id
  region  = var.region
}

terraform {
  backend "gcs" {}
}

// enable cloud APIs
resource "google_project_service" "run" {
  service = "run.googleapis.com"
}

resource "google_project_service" "sql" {
  service = "sqladmin.googleapis.com"
}

resource "google_sql_database_instance" "beaconchain-db" {
  name             = "beaconchain-db"
  region           = var.region
  database_version = "POSTGRES_17"

  settings {
    edition = "ENTERPRISE"
    tier    = "db-f1-micro"
  }

  root_password = var.db_password
}

resource "google_cloud_run_service" "personal" {
  name     = "personal-service"
  location = var.region

  template {
    spec {
      containers {
        image = var.image_url
        args  = ["--environment", "personal_cloudrun"]
        env {
          name  = "INSTANCE_CONNECTION_NAME"
          value = "${var.project_id}:${var.region}:${google_sql_database_instance.beaconchain-db.name}"
        }
      }
    }

    metadata {
      annotations = {
        "run.googleapis.com/cloudsql-instances" = "${var.project_id}:${var.region}:${google_sql_database_instance.beaconchain-db.name}"
      }
    }
  }

  traffic {
    percent         = 100
    latest_revision = true
  }
}
