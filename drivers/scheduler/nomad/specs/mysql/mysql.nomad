job "mysql" {
  datacenters = ["dc1"]
  type        = "service"

  group "mysql" {
    count = 1

    restart {
      attempts = 10
      interval = "5m"
      delay    = "25s"
      mode     = "delay"
    }

    task "mysql" {
      driver = "docker"

      env = {
          "MYSQL_ROOT_PASSWORD" = "secret"
      }
      
      config {
        image = "mysql/mysql-server:8.0"

        port_map {
          db = 3306
        }

        volumes = [
          "size=5,repl=3,name=mysql_vol:/var/lib/mysql",
        ]
        volume_driver = "pxd"
      }

      resources {
        cpu    = 500
        memory = 1024
        network {
          port "db" {}
        }
      }

      service {
        name = "mysql"
        port = "db"

        check {
          type     = "tcp"
          interval = "10s"
          timeout  = "2s"
        }
      }
    }
  }
}

