```bash
sudo go build -o /usr/local/bin/mydaemon main.go


sudo nano /etc/systemd/system/mydaemon.service
[Unit]
Description=Mi daemon en Go
After=network.target

[Service]
ExecStart=/usr/local/bin/mydaemon
Restart=always

[Install]
WantedBy=multi-user.target

```