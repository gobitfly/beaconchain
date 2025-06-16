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

locals {
  rendered_swagger = templatefile("../api/gen/beaconchain_api.swagger.json", {
    CLOUD_RUN_URL = google_cloud_run_v2_service.personal.uri
  })
  rendered_swagger_v1 = templatefile("../api/gen/beaconchain_api_v1.swagger.json", {
    CLOUD_RUN_URL = google_cloud_run_v2_service.personal.uri
  })

  swagger_files = {
    "beaconchain_api"    = local.rendered_swagger
    "beaconchain_api_v1" = local.rendered_swagger_v1
  }
}

resource "google_storage_bucket_object" "swagger_files" {
  for_each = local.swagger_files

  name         = "${each.key}.swagger.json"
  content      = each.value
  content_type = "application/json"
  bucket       = var.gcs_bucket
}

resource "google_cloud_run_v2_service" "personal" {
  name     = "beaconchain-api"
  location = var.region

  template {
    containers {
      image = "us-central1-docker.pkg.dev/${var.project_id}/service/beaconchain-api@${var.image}"
      args  = ["--environment", "personal_cloudrun"]
      env {
        name  = "INSTANCE_CONNECTION_NAME"
        value = "${var.project_id}:${var.region}:${google_sql_database_instance.beaconchain-db.name}"
      }
      volume_mounts {
        name       = "swagger-vol"
        mount_path = "/api/gen/"
      }
    }

    volumes {
      name = "swagger-vol"

      gcs {
        bucket    = var.gcs_bucket
        read_only = true
      }
    }

    annotations = {
      "run.googleapis.com/cloudsql-instances" = "${var.project_id}:${var.region}:${google_sql_database_instance.beaconchain-db.name}"
    }
  }

  traffic {
    percent = 100
    type    = "TRAFFIC_TARGET_ALLOCATION_TYPE_LATEST"
  }
}

resource "google_api_gateway_api" "api" {
  provider   = google-beta
  api_id     = "beaconchain-api"
  project    = var.project_id
}

resource "google_api_gateway_api_config" "api_cfg" {
  provider      = google-beta
  project       = var.project_id
  api           = google_api_gateway_api.api.api_id
  # api_config_id = "beaconchain-api-config"

  openapi_documents {
    document {
      path     = "beaconchain_api.swagger.json"
      contents = base64encode(local.rendered_swagger_v1)
    }
  }
  lifecycle {
    create_before_destroy = true
  }
}

resource "google_api_gateway_gateway" "gateway" {
  provider   = google-beta
  gateway_id = "beaconchain-gateway"
  api_config = google_api_gateway_api_config.api_cfg.id
  project    = var.project_id
  region     = var.region
}

resource "google_cloud_run_service_iam_member" "noauth" {
  location = google_cloud_run_v2_service.personal.location
  project  = var.project_id
  service  = google_cloud_run_v2_service.personal.name

  role   = "roles/run.invoker"
  member = "allUsers"
}
