# Guía de Autohospedaje (Self-Hosting Tracker)

Aldea es **100% gratuito y auto-hospedado**. No existen servidores centralizados ni suscripciones de pago. Esta guía detalla cómo desplegar el coordinador (`trackerd`) en una máquina virtual de capa gratuita (ej. Oracle Cloud Always Free, AWS Free Tier o cualquier VPS propio).

---

## 1. Requisitos Previos

- Un servidor Linux (Ubuntu/Debian) con una dirección IP pública.
- Puerto TCP `9090` abierto en el firewall (Ingress Rule).

---

## 2. Instalación Rápida con Binario Único

Descarga la última versión del binario `trackerd`:

```bash
# Descargar binario compilado para Linux AMD64
curl -L -o trackerd https://gitlab.com/giotld-group/Aldea/-/releases/permalink/latest/downloads/trackerd-linux-amd64
chmod +x trackerd
```

---

## 3. Generar Configuración Inicial

Ejecuta el asistente de configuración para inicializar la base de datos y generar la clave secreta de la red:

```bash
./aldea init --role=tracker --listen="0.0.0.0:9090" --out="tracker.yaml"
```

Esto generará un archivo `tracker.yaml` similar a este:

```yaml
listen_addr: 0.0.0.0:9090
db_path: ./tracker.db
network_key: a96f55f0680d0a7ef329f45418996549
```

---

## 4. Ejecución en Segundo Plano (systemd)

Crea un servicio systemd en `/etc/systemd/system/trackerd.service`:

```ini
[Unit]
Description=Aldea Tracker Coordinator Daemon
After=network.target

[Service]
Type=simple
User=ubuntu
WorkingDirectory=/opt/aldea
ExecStart=/opt/aldea/trackerd --config /opt/aldea/tracker.yaml
Restart=always
RestartSec=5

[Install]
WantedBy=multi-user.target
```

Habilita e inicia el servicio:

```bash
sudo systemctl daemon-reload
sudo systemctl enable --now trackerd
sudo systemctl status trackerd
```

¡Listo! El tracker ya está coordinando la topología P2P de tu grupo sin almacenar ningún fragmento de datos.
