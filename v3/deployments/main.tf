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
}

resource "google_api_gateway_api" "api" {
  provider = google-beta
  api_id   = "beaconchain-api"
  project  = var.project_id
}

resource "google_api_gateway_api_config" "api_cfg" {
  provider = google-beta
  project  = var.project_id
  api      = google_api_gateway_api.api.api_id
  # omitting causes tf to create random ids which won't be cleaned up, has to be done for dependency / uptime reasons
  # clean up manually or add a script
  # api_config_id = "beaconchain-api-config"

  openapi_documents {
    document {
      path     = "api-gateway.json"
      contents = base64encode(local.merged_swagger)
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
  location = google_cloud_run_v2_service.personal-external.location
  project  = var.project_id
  service  = google_cloud_run_v2_service.personal-external.name

  role   = "roles/run.invoker"
  member = "allUsers"
}
