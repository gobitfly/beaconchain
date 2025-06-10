variable "project_id" {}
variable "image" {
  default = "latest"
}
variable "db_password" {}
variable "region" {
  default = "us-central1"
}
variable "gcs_bucket" {}