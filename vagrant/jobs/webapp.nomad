job "webapp" {
  datacenters = ["dc1"]

  group "demo" {
    count = 3

    network {
      port "webapp_http" {}
    }

    scaling {
      enabled = true
      min     = 1
      max     = 20

      policy {
        cooldown = "20s"

        check "cloudwatch" {
          source = "cloudwatch"
          query  = "SELECT MAX(ApproximateNumberOfMessagesVisible) FROM SCHEMA(\"AWS/SQS\", QueueName) WHERE QueueName = 'MY_SQS_QUEUE'"

          strategy "target-value" {
            target = 50
          }
        }
      }
    }

    task "webapp" {
      driver = "docker"

      config {
        image = "hashicorp/demo-webapp-lb-guide"
        ports = ["webapp_http"]
      }

      env {
        PORT    = "${NOMAD_PORT_webapp_http}"
        NODE_IP = "${NOMAD_IP_webapp_http}"
      }

      resources {
        cpu    = 100
        memory = 16
      }
    }
  }
}
