variable "dns_zone_name" {}
variable "system_domain_suffix" {}

resource "google_dns_record_set" "default" {
  name = "*.${var.env_id}.${var.system_domain_suffix}."
  type = "A"
  ttl = 300

  managed_zone = var.dns_zone_name
  rrdatas = [ google_compute_address.bosh-director-ip.address ]
}