output "cloud_run_url" {
  value = google_cloud_run_service.personal.status[0].url
}