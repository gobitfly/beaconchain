output "internal_cloud_run_url" {
  description = "URL of the internal Cloud Run service"
  value = google_cloud_run_v2_service.personal-internal.uri
}

output "external_cloud_run_url" {
  description = "URL of the external Cloud Run service"
  value = google_cloud_run_v2_service.personal-external.uri
}
