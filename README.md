# Aldea

Aldea es una red de almacenamiento distribuido y cómputo peer-to-peer (P2P), auto-hospedada y de código abierto, diseñada para grupos cerrados (amigos, familias, pequeños equipos). Permite agrupar el espacio de disco no utilizado y la capacidad de cómputo de dispositivos personales a través de diferentes ubicaciones sin depender de servidores VPS de pago, proveedores en la nube centralizados ni infraestructura de terceros.

---

## Características Principales

### 1. Almacenamiento Distribuido con Cero Conocimiento (Zero-Knowledge)
- **Cifrado en el Cliente**: Cifrado simétrico de alta seguridad mediante XChaCha20-Poly1305 con derivación de claves Argon2id. Los datos se cifran localmente antes de salir de la máquina de origen; los nodos de almacenamiento nunca ven contenido en texto plano.
- **Redundancia y Tolerancia a Fallos**: Fragmentación y codificación de borrado mediante Reed-Solomon ($K=4, M=4$). Cada archivo se divide en 8 fragmentos (100% de redundancia), lo que permite reconstruir la información original incluso si hasta 4 nodos fallan o se desconectan simultáneamente.
- **División en Chunks Fijos**: Los archivos se dividen en bloques de 1 MB para una distribución uniforme y eficiente en la red P2P.
- **Auto-Reparación de Datos**: El motor de monitoreo detecta nodos desconectados (churn) y reconstruye automáticamente los fragmentos faltantes en nodos saludables activos.

### 2. Capa de Cómputo Distribuido (MicroVMs & Contenedores)
- **Aislamiento Seguro**: Ejecución de cargas de trabajo en contenedores utilizando Kata Containers y microVMs Firecracker (`internal/runtime`), garantizando aislamiento a nivel de kernel en entornos Linux.
- **Planificador P2P**: Distribución equilibrada de instancias de cómputo entre nodos pares en función de la capacidad de CPU y memoria RAM disponible.
- **Enrutamiento e Ingress P2P**: Enrutamiento HTTP público/privado a servicios en ejecución con conmutación por error (stateless failover).
- **Respaldos de Volumen de Estado**: Capturas de pantalla (snapshots) periódicas de volúmenes persistentes almacenadas de forma cifrada en la red de almacenamiento P2P.

### 3. Red P2P y Traversal de Red (NAT)
- **Descubrimiento y Enrutamiento DHT**: Tabla de hash distribuida (Kademlia DHT) para el descubrimiento directo entre pares.
- **Travesía de NAT y Routers**: Soporte integrado para UPnP y STUN para conexiones P2P directas en redes domésticas.
- **Servidores de Relevo (Relay Fallback)**: Mecanismo de retransmisión autenticado con HMAC-SHA256 (TURN/WebSocket) para entornos con NAT simétrico restrictivo.
- **Control de Ancho de Banda**: Limitación de tasas de subida y bajada por nodo y gestión de cuotas de disco.

### 4. Interfaces de Usuario
- **CLI (`aldea`)**: Cliente de línea de comandos para gestión de credenciales, transferencia de archivos y despliegue de contenedores.
- **GUI de Escritorio (`aldea-desktop`)**: Aplicación nativa multiplataforma desarrollada en Wails v2 (Go + HTML/CSS/JS) con integración a la bandeja del sistema (System Tray GTK/Windows/macOS), métricas en tiempo real y panel de control visual.

---

## Arquitectura del Sistema

```
                       +-------------------+
                       |     trackerd      |
                       |   (Coordinador)   |
                       +---------+---------+
                                 | Metadatos / Ubicación
                                 v
    +-------------------+ +-------------------+ +-------------------+
    |       noded       | |       noded       | |       noded       |
    |    (Almacén A)    |<|    (Almacén B)    |<|    (Almacén C)    |
    +-------------------+ +-------------------+ +-------------------+
              ^                     ^                     ^
              |                     |                     |
              +---------------------+---------------------+
                          Red P2P Encubierta
```

### Componentes del Código
- `cmd/aldea`: CLI interactiva para clientes.
- `cmd/noded`: Daemon de nodo de almacenamiento y cómputo que se ejecuta en los dispositivos participantes (base de datos integrada `bbolt`).
- `cmd/trackerd`: Daemon coordinador que gestiona el estado de salud de los nodos y mantiene el catálogo de metadatos de ubicación (base de datos integrada `bbolt`).
- `gui/`: Aplicación de escritorio Wails v2 con panel de métricas, explorador de archivos P2P y gestor de microVMs.

---

## Requisitos de Instalación

- **Go**: Versión 1.22 o superior.
- **CGO & Herramientas de Compilación**: `gcc`, `pkg-config`.
- **Dependencias de GUI (Linux)**: `webkit2gtk-4.1` (o `webkit2gtk-4.0` en distribuciones compatibles), `gtk3`.
- **Motor de Cómputo (Opcional para nodos con soporte microVM)**: Linux kernel con soporte KVM / Kata Containers / Firecracker.

---

## Guía de Inicio Rápido

### 1. Compilación de Binarios

Para compilar las herramientas de línea de comandos y demonios:

```bash
./scripts/build_all.sh
```

Los ejecutables generados se ubicarán en `build/bin/`:
- `build/bin/aldea`
- `build/bin/noded`
- `build/bin/trackerd`

Para compilar la aplicación de escritorio nativa (Wails):

```bash
./scripts/build_desktop.sh
```

El ejecutable resultante estará en `build/bin/aldea-desktop-linux-amd64` (o la arquitectura correspondiente).

---

### 2. Prueba Local con Docker Compose

Es posible desplegar un clúster local de prueba compuesto por 8 nodos de almacenamiento y 1 coordinador mediante Docker Compose:

```bash
docker compose up -d --build
```

Para verificar el estado de los contenedores:

```bash
docker compose ps
```

---

### 3. Uso desde la Línea de Comandos (CLI)

#### Inicializar configuración de cliente:
```bash
./build/bin/aldea init --tracker "http://localhost:8080" --key "clave-secreta-de-32-caracteres!"
```

#### Consultar estado de la red y nodos activos:
```bash
./build/bin/aldea status
```

#### Subir un archivo a la red P2P:
```bash
./build/bin/aldea put /ruta/a/mi_imagen.png
```
*Salida:* `uploaded: /ruta/a/mi_imagen.png -> <fileID>`

#### Descargar un archivo desde la red:
```bash
./build/bin/aldea get <fileID> /ruta/a/imagen_recuperada.png
```

#### Desplegar un contenedor / microVM de cómputo:
```bash
./build/bin/aldea compute deploy --image nginx:alpine --name mi-servidor --cpus 1 --ram 512
```

#### Listar cargas de trabajo en ejecución:
```bash
./build/bin/aldea compute list
```

#### Detener y eliminar una microVM:
```bash
./build/bin/aldea compute terminate <vmID>
```

---

### 4. Demostración de Resiliencia Automática

El repositorio incluye un script para verificar la reconstrucción automática de fragmentos mediante Reed-Solomon ($K=4, M=4$) ante la caída intencionada de nodos:

```bash
./scripts/demo.sh
```

---

## Estructura del Repositorio

```
.
├── cmd/
│   ├── aldea/          # Cliente CLI principal
│   ├── noded/          # Daemon de nodo de almacenamiento/cómputo
│   └── trackerd/       # Daemon coordinador de metadatos
├── deploy/
│   └── docker/         # Configuraciones de despliegue Docker
├── gui/                # Aplicación de escritorio nativa Wails v2 (Go + Frontend Web)
├── internal/
│   ├── bandwidth/      # Control de cuotas y límite de tasa de transferencia
│   ├── chunker/        # Fragmentación de archivos en bloques de 1 MB
│   ├── config/         # Carga y validación de configuración
│   ├── crypto/         # Cifrado cliente XChaCha20-Poly1305 y Argon2id
│   ├── dht/            # Tabla de Hash Distribuida Kademlia
│   ├── erasure/        # Codificación Reed-Solomon (klauspost/reedsolomon)
│   ├── ingress/        # Enrutamiento de entrada P2P HTTP
│   ├── invite/         # Generación y validación de tokens de invitación
│   ├── metrics/        # Recolección de métricas de sistema y rendimiento
│   ├── nat/            # Traversal de NAT (STUN y UPnP)
│   ├── protocol/       # Protocolos wire autenticados (HMAC-SHA256)
│   ├── relay/          # Servidores de relevo TURN/WebSocket
│   ├── repair/         # Motor de auto-reparación de fragmentos
│   ├── runtime/        # Aislamiento de ejecución en contenedores (Kata/Firecracker)
│   ├── scheduler/      # Planificador de cargas de trabajo de cómputo
│   ├── snapshot/       # Copias de seguridad de volúmenes persistentes
│   └── tracker/        # Motor de metadatos y asignación de bloques (bbolt)
├── scripts/            # Scripts de compilación y demostración
└── test/               # Pruebas de integración E2E
```

---

## Pruebas

Para ejecutar el conjunto completo de pruebas unitarias e integración con detección de carreras (`-race`):

```bash
go test -v -race ./...
```

---

## Licencia

Este proyecto está distribuido bajo los términos de la licencia MIT.
