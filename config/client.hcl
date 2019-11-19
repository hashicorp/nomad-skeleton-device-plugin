log_level = "TRACE"

plugin "example-device-hello-world" {
  config {
    fingerprint_period = "20m"
    greeting           = "¡Buenos dias!"
    greetings_per_node = 5
  }
}
