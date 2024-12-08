# Birthsday Sceduler
Un semplice service per mandare dei messaggi (al momento via mail) che mi 
ricordano dei compleanni.

### Stop del service
Per stoppare il sevice si usa:

    sudo systemctl stop birthday-scheduler

## Deployment su ubuntu direttamente

    cd ~/build/birthday-scheduler
    git pull --all
    ./publish-service.sh

## Service setup
Ora bisogna abilitare il service:

    sudo systemctl enable birthday-scheduler.service
Ora si fa partire il service (resistente al reboot):

    sudo systemctl start birthday-scheduler
Per vedere i logs si usa:

    sudo journalctl -f -u birthday-scheduler

## Service Config
Questo il conetnuto del file che compare con:

    sudo nano /lib/systemd/system/birthday-scheduler.service
Poi si fa l'enable:

    sudo systemctl enable birthday-scheduler.service
E infine lo start:

    sudo systemctl start birthday-scheduler
Logs sono disponibili con:

    sudo journalctl -f -u birthday-scheduler

Qui segue il contenuto del file birthday-scheduler.service
Nota il Type=idle che è meglio di simple in quanto così 
viene fatto partire quando anche la wlan ha ottenuto l'IP intranet
per consentire l'accesso.

```
[Install]
WantedBy=multi-user.target

[Unit]
Description=birthday-scheduler service
ConditionPathExists=/home/igor/app/go/birthday-scheduler/current/birthday-scheduler.bin
After=network.target

[Service]
Type=idle
User=igor
Group=igor
LimitNOFILE=1024

Restart=on-failure
RestartSec=10
startLimitIntervalSec=60

WorkingDirectory=/home/igor/app/go/birthday-scheduler/current/
ExecStart=/home/igor/app/go/birthday-scheduler/current/birthday-scheduler.bin

# make sure log directory exists and owned by syslog
PermissionsStartOnly=true
ExecStartPre=/bin/mkdir -p /var/log/birthday-scheduler
ExecStartPre=/bin/chown igor:igor /var/log/birthday-scheduler
ExecStartPre=/bin/chmod 755 /var/log/birthday-scheduler
StandardOutput=syslog
StandardError=syslog

```

## Data.json
Nel file data.json ho messo la lista dei compleanni che mi devo ricordare.

Per aggiornare il server di invido, mi piazzo locale nella dir cert e mando:

    rsync -av data.json <user>@<server>:/home/igor/app/go/birthday-scheduler/current/

