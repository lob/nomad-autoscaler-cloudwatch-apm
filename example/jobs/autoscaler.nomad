job "autoscaler" {
  type = "service"

  datacenters = ["dc1"]

  group "autoscaler" {
    count = 1

    network {
      port "http" {}
    }

    task "autoscaler" {
      driver = "docker"

      config {
        image   = "hashicorp/nomad-autoscaler:0.3.6"
        command = "nomad-autoscaler"
        ports   = ["http"]

        args = [
          "agent",
          "-config",
          "${NOMAD_TASK_DIR}/config.hcl",
          "-http-bind-address",
          "0.0.0.0",
          "-http-bind-port",
          "${NOMAD_PORT_http}",
          "-plugin-dir",
          "${NOMAD_TASK_DIR}/plugins",
        ]

        volumes = [
          "/home/vagrant/nomad-autoscaler-plugins/nomad-autoscaler-cloudwatch-apm_linux_amd64:${NOMAD_TASK_DIR}/plugins/nomad-autoscaler-cloudwatch-apm",
        ]
      }

      template {
        data = <<EOF
nomad {
  address = "http://{{env "attr.unique.network.ip-address" }}:4646"
}


apm "cloudwatch" {
  driver = "nomad-autoscaler-cloudwatch-apm"

  config = {
    aws_region            = "us-east-1"
    aws_access_key_id     = "<AWS_ACCESS_KEY_ID>"
    aws_secret_access_key = "<AWS_SECRET_ACCESS_KEY>"
    aws_session_token     = "<AWS_SESSION_TOKEN>"
  }
}

strategy "target-value" {
  driver = "target-value"
}
          EOF

        destination = "${NOMAD_TASK_DIR}/config.hcl"
      }

      resources {
        cpu    = 50
        memory = 128
      }

      service {
        name = "autoscaler"
        port = "http"

        check {
          type     = "http"
          path     = "/v1/health"
          interval = "3s"
          timeout  = "1s"
        }
      }
    }

  }
}
