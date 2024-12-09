# Birthsday Sceduler
Un semplice service per mandare dei messaggi (via relay-mail e/o telegram) che mi 
ricordano i compleanni ai quali voglio mandare gli auguri.
L'allarme è configurabile via template e può essere simulato (mostra il messaggio senza mandarlo). 
Se la modalità Telegram o Mail non è configurata, l'allarme non viene inviato attraverso quel canale.

### Stop del service
Per stoppare il sevice si usa:

    sudo systemctl stop birthday-scheduler

## Deployment su ubuntu direttamente

    cd ~/build/birthday-scheduler
    git pull --all
    ./publish-service.sh
Oppure uso Visual Code in remoto dove uso il synch di git. Qui nel terminal mi basta usare:

    ./publish-service.sh

## Service Config
Per prima cosa va creato il file birthday-scheduler.service.
Il contenuto l'ho messo sotto in una sezione apposita.

    sudo nano /lib/systemd/system/birthday-scheduler.service
Poi si fa l'enable:

    sudo systemctl enable birthday-scheduler.service
E infine lo start:

    sudo systemctl start birthday-scheduler
Logs sono disponibili con:

    sudo journalctl -f -u birthday-scheduler

## birthday-scheduler.service
Qui segue il contenuto del file birthday-scheduler.service
Nota il Type=idle che è meglio di simple. Così 
viene fatto partire dopo che la wlan ha ottenuto l'IP intranet
e così si ha l'accesso.

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
Oppure edito il file direttamente con Visual Code con copia e incolla dal mio PC.

### config_custom.toml
È il file che mi esegue un ovveride del file config.toml. 
Mi serve in quanto config.toml si trova su gitHub, mentre config_custom.toml è
solo locale fuori da git. Si trova in:

    /home/igor/app/go/birthday-scheduler/current/

## Visual Code
Per lo sviluppo iniziale ho usato windows, poi, per l'update del service,
ho usato Visual Code Remote nella directory ~/build/birth-scheduler.

