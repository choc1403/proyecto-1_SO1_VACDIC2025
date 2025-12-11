```bash
/etc/systemd/system/so1daemon.service
[Unit]
Description=SO1 Daemon (Container Manager)
After=network.target docker.service
Requires=docker.service

[Service]
ExecStart=/ruta/a/tu/go-daemon/so1-daemon
WorkingDirectory=/ruta/a/tu/go-daemon
Restart=always
User=root
Environment=CGO_ENABLED=1

[Install]
WantedBy=multi-user.target

```