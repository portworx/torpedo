job "nginx" {
  datacenters = ["dc1"]

  group "nginx" {
    count = 4

    task "nginx" {
      driver = "docker"

      config {
        image = "nginx"
        port_map {
          http = 80
        }
        volumes = [
          "size=5,repl=3,shared=true,name=nomad_vol:/mnt/test",
        ]
        volume_driver = "pxd"
      }
      artifact = {
        source = "https://gist.githubusercontent.com/michaelalhilly/121cfe78ee6d9c6b80e4043e36458082/raw/3e808a9d498aba551fdfee4523331f8f8378b10f/index.html"
        destination = "index.html"
      }

      service {
        name = "nginx"
        tags = ["global"]
        port = "http"

        check {
          name     = "nginx alive"
          type     = "tcp"
          interval = "10s"
          timeout  = "2s"
        }
      }

      resources {
        cpu = 500
        memory = 64

        network {
            mbits = 10
            port "http" {}
        }
      }
    }
  }
}
