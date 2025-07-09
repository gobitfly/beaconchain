provider "google" {
  project = var.project_id
  region  = var.region
}

terraform {
  backend "gcs" {}
}

// cloud APIs
resource "google_project_service" "run" {
  service = "run.googleapis.com"
}

resource "google_project_service" "sql" {
  service = "sqladmin.googleapis.com"
}

resource "google_project_service" "cloudbuild" {
  service = "cloudbuild.googleapis.com"
}

resource "google_project_service" "service_control" {
  service = "servicecontrol.googleapis.com"
}

// dependency resources
resource "google_sql_database_instance" "beaconchain-db" {
  name             = "beaconchain-db"
  region           = var.region
  database_version = "POSTGRES_17"

  settings {
    edition = "ENTERPRISE"
    tier    = "db-f1-micro"
  }

  root_password = var.db_password
  depends_on = [google_project_service.sql]
}


// service resources
locals {
  internal_swagger = templatefile("../api/gen/internal.swagger.json", {
    CLOUD_RUN_URL = google_cloud_run_v2_service.personal-internal.uri
  })
  external_swagger = templatefile("../api/gen/external.swagger.json", {
    CLOUD_RUN_URL = google_cloud_run_v2_service.personal-external.uri
  })

  merged_swagger = jsonencode(merge(
    jsondecode(local.internal_swagger),
    {
      paths = merge(
        try(jsondecode(local.internal_swagger).paths, {}),
        try(jsondecode(local.external_swagger).paths, {})
      )
    },
    {
      definitions = merge(
        try(jsondecode(local.internal_swagger).definitions, {}),
        try(jsondecode(local.external_swagger).definitions, {})
      )
    },
    {
      tags = concat(
        try(jsondecode(local.internal_swagger).tags, {}),
        try(jsondecode(local.external_swagger).tags, {})
      )
    }
  ))

  swagger_files = {
    "internal" = local.internal_swagger
    "external" = local.external_swagger
  }
}

resource "google_storage_bucket_object" "swagger_files" {
  for_each = local.swagger_files

  name         = "${each.key}.swagger.json"
  content      = each.value
  content_type = "application/json"
  bucket       = var.gcs_bucket
}

resource "google_cloud_run_v2_service" "personal-internal" {
  name     = "beaconchain-api-internal"
  location = var.region

  template {
    containers {
      image = "us-central1-docker.pkg.dev/${var.project_id}/service/beaconchain-api@${var.image}"
      args = [
        "--environment", "personal_cloudrun",
        "--type", "internal",
      ]
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
  deletion_protection = false

  depends_on = [google_project_service.run]
}

resource "google_cloud_run_v2_service" "personal-external" {
  name     = "beaconchain-api-external"
  location = var.region

  template {
    containers {
      image = "us-central1-docker.pkg.dev/${var.project_id}/service/beaconchain-api@${var.image}"
      args = [
        "--environment", "personal_cloudrun",
        "--type", "external",
      ]
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
  deletion_protection = false

  depends_on = [google_project_service.run]
}

resource "google_cloud_run_service_iam_member" "noauth" {
  location = google_cloud_run_v2_service.personal-external.location
  project  = var.project_id
  service  = google_cloud_run_v2_service.personal-external.name

  role   = "roles/run.invoker"
  member = "allUsers"
}
