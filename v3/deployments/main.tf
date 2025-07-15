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

resource "google_project_service" "vpcaccess" {
  service = "vpcaccess.googleapis.com"
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
  depends_on    = [google_project_service.sql]
}

resource "google_compute_network" "mono-vpc" {
  name = "vpc-network"
}

resource "google_compute_firewall" "default" {
  name    = "default"
  network = google_compute_network.mono-vpc.name

  allow {
    protocol = "tcp"
    ports    = ["22"]
  }

  source_ranges = ["35.235.240.0/20"]
}

// service resources
locals {
  internal_swagger = templatefile("../api/gen/api_service/v1/internal.swagger.json", {
    CLOUD_RUN_URL = google_cloud_run_v2_service.personal-internal.uri
  })
  external_swagger = templatefile("../api/gen/api_service/v1/external.swagger.json", {
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

  # internal access only
  provider = google-beta
  default_uri_disabled = false
  ingress  = "INGRESS_TRAFFIC_INTERNAL_ONLY"

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
      env {
        name  = "GRPCPORT"
        value = "8080"
      }
      
      volume_mounts {
        name       = "swagger-vol"
        mount_path = "/api/gen/"
      }
      volume_mounts {
        name = "cloudsql"
        mount_path = "/cloudsql"
      }
    }

    volumes {
      name = "swagger-vol"

      gcs {
        bucket    = var.gcs_bucket
        read_only = true
      }
    }

    volumes {
      name = "cloudsql"
      cloud_sql_instance {
        instances = [google_sql_database_instance.beaconchain-db.connection_name]
      }
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
  
  # no public access for now
  ingress  = "INGRESS_TRAFFIC_INTERNAL_ONLY"

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

      volume_mounts {
        name = "cloudsql"
        mount_path = "/cloudsql"
      }
    }

    volumes {
      name = "swagger-vol"

      gcs {
        bucket    = var.gcs_bucket
        read_only = true
      }
    }

    volumes {
      name = "cloudsql"
      cloud_sql_instance {
        instances = [google_sql_database_instance.beaconchain-db.connection_name]
      }
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
