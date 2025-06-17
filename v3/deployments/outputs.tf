output "internal_cloud_run_url" {
  value = google_cloud_run_v2_service.personal-internal.uri
}

output "external_cloud_run_url" {
  value = google_cloud_run_v2_service.personal-external.uri
}

output "api_gateway_url" {
  value = "https://${google_api_gateway_gateway.gateway.default_hostname}"
}