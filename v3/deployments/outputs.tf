output "cloud_run_url" {
  value = google_cloud_run_v2_service.personal.uri
}

output "api_gateway_url" {
  value = "https://${google_api_gateway_gateway.gateway.default_hostname}"
}