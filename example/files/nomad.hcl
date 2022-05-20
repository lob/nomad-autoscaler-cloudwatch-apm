datacenter = "dc1"

data_dir = "/opt/nomad"

server {
  enabled          = true
  bootstrap_expect = 1
}

client {
  enabled = true
}

plugin "docker" {
  config {
    volumes {
      enabled = true
    }
  }
}

