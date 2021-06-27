<!-- markdownlint-disable MD024 MD041 -->

[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]

<!-- PROJECT LOGO -->
<br />
<p align="center">
  <h3 align="center">agones-mc</h3>

  <p align="center">
  Minecraft server CLI for Agones GameServers
    <br />
    <a href="https://github.com/saulmaldonado/agones-mc"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://github.com/saulmaldonado/agones-mc/tree/main/example/mc-server.yml">View Example</a>
    ·
    <a href="https://github.com/saulmaldonado/agones-mc/issues">Report Bug</a>
    ·
    <a href="https://github.com/saulmaldonado/agones-mc/issues">Request Feature</a>
  </p>
</p>

<!-- TABLE OF CONTENTS -->
<details open="open">
  <summary><h2 style="display: inline-block">Table of Contents</h2></summary>
  <ol>
    <li>
      <a href="#about-the-project">About The Project</a>
      <ul>
        <li><a href="#built-with">Built With</a></li>
      </ul>
    </li>
    <li>
      <a href="#getting-started">Getting Started</a>
      <ul>
        <li><a href="#prerequisites">Prerequisites</a></li>
        <li><a href="#installation">Installation</a></li>
      </ul>
    </li>
    <li><a href="#usage">Usage</a></li>
    <li><a href="#roadmap">Roadmap</a></li>
    <li><a href="#contributing">Contributing</a></li>
    <li><a href="#license">License</a></li>
    <li><a href="#acknowledgements">Acknowledgements</a></li>
    <li><a href="#author">Author</a></li>
  </ol>
</details>

<!-- ABOUT THE PROJECT -->

## About The Project

This application was built to run as a sidecar container alongside Minecraft server in an Agones GameServer Pod to integrate it with the Agones SDK and assist with lifecycle management, health checking, world loading and backup, and configuration editing.

### Built With

- [cobra](https://github.com/spf13/cobra)
- [mc-pinger](https://github.com/Raqbit/mc-pinger)
- [go-bedrockping](https://github.com/ZeroErrors/go-bedrockping)
- [Agones Go SDK](agones.dev/agones/sdks/go)

<!-- GETTING STARTED -->

## Getting Started

### Prerequisites

- Kubernetes

  - Set up a Kubernetes cluster using local or cloud solutions
  - [https://agones.dev/site/docs/installation/creating-cluster/](https://agones.dev/site/docs/installation/creating-cluster/)

- Agones

  - Set up Agones in your cluster
  - [https://agones.dev/site/docs/installation/install-agones/](https://agones.dev/site/docs/installation/install-agones/)

- Minecraft Server
  - Compatible with Java and Bedrock editions
  - You'll need a container to run the Minecraft server. Highly recommend using [itzg/docker-minecraft-server](https://github.com/itzg/docker-minecraft-server) or [itzg/minecraft-bedrock-server](https://github.com/itzg/docker-minecraft-bedrock-server)

### Installation

#### Create a new Minecraft GameServer With Sidecar

```sh
kubectl create -f example/mc-server.yml
```

[Full Java GameServer specification example](./example/mc-server.yml)
[Full Bedrock GameServer specification example](./example/mc-server-bedrock.yml)

<!-- USAGE EXAMPLES -->

## Usage

### Monitor

```sh
agones-mc monitor
```

### Environment variables

- `INITIAL_DELAY`: Initial startup delay before first ping (default `30s`)
- `HOST`: Minecraft server host (default `"localhost"`)
- `PORT`: Minecraft server port (default `25565`)
- `EDITION`: Minecraft server edition. java or bedrock (default `"java"`)
- `INTERVAL`: Server ping interval (default `10s`)
- `TIMEOUT`: Max ping duration before timeout (default `10s`)
- `MAX_ATTEMPTS`: Ping attempt limit. Process will end after failing the last (default `5`)

To utilize Agones GameServer health checking, game containers need to interact with the SDK server sidecar. This sidecar process will ping Minecraft Java/Bedrock game containers and report container health to the SDK server.

On Pod creation it will repeatedly ping minecraft server in the same Pod network (`localhost`) every `interval` (defaults to `10s`). The first successful ping will call `Ready()`. Every subsequent ping will call `Health()`. For an unsuccessful ping, the process will attempt to ping server until successful or until a total of `attempts` (default `5`) consecutive failed pings at which the process will exit

If the server is pinged while starting up (initial world generation), the ping will be considered successful but `Ready()` would not be called.

#### GameServer Pod template example

```yml
template:
  spec:
    containers:
      - name: mc-server
        image: itzg/minecraft-server # Minecraft Java server image
        env: # Full list of ENV variables at https://github.com/itzg/docker-minecraft-server
          - name: EULA
            value: 'TRUE'

      - name: mc-monitor
        image: saulmaldonado/agones-mc # monitor
        args:
          - monitor
        env:
          - name: INITIAL_DELAY
            value: 30s
          - name: MAX_ATTEMPTS
            value: 5
          - name: INTERVAL
            value: 10s
          - name: TIMEOUT
            value: 10s
        imagePullPolicy: Always
```

[Full Java GameServer specification example](./example/mc-server.yml)

[Full Bedrock GameServer specification example](./example/mc-bedrock-server.yml)

### Run Locally with Docker

Run an example Minecraft GameServer, Agones SDK, and monitor Pod locally with `docker-compose`

```sh
docker-compose -f monitor.docker-compose.yml up

# or

make docker-compose.monitor
```

### Backup

```sh
agones-mc backup
```

### Environment variables

- `INITIAL_DELAY`: Initial startup delay before first ping (default `30s`)
- `VOLUME`: Mounted volume path into the server's minecraft data directory (default `"/data"`)
- `HOST`: Minecraft server host (default `"localhost"`)
- `PORT`: Minecraft server port (default `25565`)
- `EDITION`: Minecraft server edition. java or bedrock (default `"java"`)
- `BUCKET_NAME`: GCP bucket name for backups (default `""`)
- `BACKUP_NAME`: Archived world backup name (default `""`)
- `BACKUP_CRON`: crontab for the backup job (default will run job once)
- `RCON_PASSWORD`: Password for server's RCON (default `"minecraft"`)
- `POD_NAME`: Pod name for logging (default `""`)

`backup` will creates zip archives of world for backup to Google Cloud Storage. The process will use the host's Application Default Credentials (ADC) or attached service account (provided by GCE, GKE, etc.). To run as a sidecar, the container will need a shared volume with the minecraft server's `/data` directory.

If a crontab is provided through `BACKUP_CRON` the process will schedule backup job according to it, otherwise the backup job will only run once at startup.

If an `RCON_PASSWORD` env variable is set on the container, the process will attempt to call `save-all` on the minecraft server before backing up

When starting a backup job the process will copy the world data at `/data/world` into a zip with the name `<SERVER_NAME>-<UTC_TIMESTAMP>.zip`. The zip will then be uploaded to Google Cloud Storage into the bucket specified by `BUCKET_NAME`

#### GameServer Pod template example

```yml
template:
  spec:
    containers:
      - name: mc-server
        image: itzg/minecraft-server # Minecraft Java server image
        env: # Full list of ENV variables at https://github.com/itzg/docker-minecraft-server
          - name: EULA
            value: 'TRUE'
        volumeMounts:
          - mountPath: /data # shared vol with mc-load and mc-backup
            name: world-vol

      - name: mc-backup
        image: saulmaldonado/agones-mc # backup
        args:
          - backup
        env:
          - name: BUCKET_NAME
            value: agones-minecraft-mc-worlds
          - name: BACKUP_CRON
            value: 0 */6 * * *
          - name: INITIAL_DELAY
            value: 60s
          - name: VOLUME
            value: /data
          - name: POD_NAME
            valueFrom:
              fieldRef:
                fieldPath: metadata.name # GameServer ref for naming backup zip files
          - name: RCON_PASSWORD
            value: minecraft # default rcon password. If provided RCON connection will be used to execute 'save-all' before a backup job.
            # Change the rcon password when exposing RCON port outside the pod
        imagePullPolicy: Always
        volumeMounts:
          - mountPath: /data # shared vol with mc-server
            name: world-vol

      volumes:
      - name: world-vol # shared vol between containers. will not persist between restarts
        emptyDir: {}
```

[Full Java GameServer specification example](./example/mc-server.yml)

[Full Bedrock GameServer specification example](./example/mc-bedrock-server.yml)

### Run Locally with Docker

Run an example Minecraft GameServer Pod locally with `docker-compose`

```sh
docker-compose -f backup.docker-compose.yml up

# or

make docker-compose.backup
```

### Load

```sh
  agones-mc load
```

### Environment variables

- `BUCKET_NAME`: GCP bucket name for backups (default `""`)
- `BACKUP_NAME`: Archived world backup name to load (default `""`)
- `VOLUME`: volume mount path to load minecraft world into (default `"/data"`)
- `POD_NAME`: Pod name for logging (default `""`)

Load is an initContainer process that will download an archived world from Google Cloud Storage and load it into the Minecraft container's world directory.

The name of the archived world must be specified using the `BACKUP` env variable. This can be done in a Pod template using a `fieldRef` to a Pod annotation

For example:

The name of the archived world can be specified using `'agones.dev/sdk-backup'` annotation on the pod template (`template.metadata.annotations['agones.dev/sdk-backup']`) and referenced using `metadata.annotations['agones.dev/sdk-backup']`

When downloaded the zip file will be placed into `/data/world.zip` on the current container. A shared volume between the container and the minecraft server's container should be used to place the zip into the minecraft server's `/data` directory. When using the `itzg/minecraft-server` container image, specifying a `WORLD` environment variable that points to the location of an archived zip file will cause the startup script to unzip the world and load it into the `/data/world` directory

#### GameServer Pod template example

```yml
template:
  metadata:
    annotations:
      agones.dev/sdk-backup: mc-server-qfsgr-2021-05-09T09:35:00Z.zip # mc-load will download this archived world from storage
  spec:
    initContainers:
      - name: mc-load
        image: saulmaldonado/agones-mc # backup
        args:
          - load
        env:
          - name: BUCKET_NAME
            value: agones-minecraft-mc-worlds
          - name: POD_NAME
            valueFrom:
              fieldRef:
                fieldPath: metadata.name # GameServer name ref for logging
          - name: BACKUP_NAME
            valueFrom:
              fieldRef:
                fieldPath: metadata.annotations['agones.dev/sdk-backup'] # ref to agones.dev/sdk-backup to download archived world
          - name: VOLUME
            value: /data
        imagePullPolicy: Always
        volumeMounts:
          - mountPath: /data # shared vol with mc-server
            name: world-vol

    containers:
      - name: mc-server
        image: itzg/minecraft-server # Minecraft Java server image
        env: # Full list of ENV variables at https://github.com/itzg/docker-minecraft-server
          - name: EULA
            value: "TRUE"
          - name: WORLD
            value: /data/world.zip # path to archived world in shared vol. mc-load initcontainer will download and place the archive at /data/world.zip
        volumeMounts:
          - mountPath: /data # shared vol with mc-load and mc-backup
            name: world-vol

    volumes:
      - name: world-vol # shared vol between containers. will not persist between restarts
        emptyDir: {}
```

[Full Java GameServer specification with world loading example](./example/mc-server-load.yml)

[Full Bedrock GameServer specification with world loading example](./example/mc-bedrock-server-load.yml)

### Run Locally with Docker

Run an example Minecraft GameServer Pod locally with `docker-compose`

```sh
docker-compose -f load.docker-compose.yml up

# or

make docker-compose.load
```

### Fileserver

```sh
  agones-mc fileserver
```

### Environment variables

- `VOLUME`: volume mount path to load minecraft world into (default `"/data"`)

fileserver is a Go http fileserver for viewing, adding, editing, and deleting configuration files in the shared minecraft data volume in a pod.

`GET: /:path-to-file`

Response: `Content-Type: application/json`

Access file or traverse directory in the volume

`POST: /:filename`

Request: `Content-Type: multipart/form-data`

Creates a new file in the volume

`PUT: /:filename`

Request: `Content-Type: multipart/form-data`

edits existing file in the volume

`DELETE: /filename`

deletes existing file in the volume

For example:

`GET: /whitelist.json` will download the `/data/whitelist.json`

#### GameServer Pod template example

```yml
template:
  spec:
    containers:
      - name: mc-server
        image: itzg/minecraft-server # Minecraft Java server image
        env: # Full list of ENV variables at https://github.com/itzg/docker-minecraft-server
          - name: EULA
            value: "TRUE"
        volumeMounts:
          - mountPath: /data # shared vol with mc-load and mc-backup
            name: world-vol

      - name: mc-fileserver # fileserver
          image: saulmaldonado/agones-mc
          args:
            - fileserver
          env:
            - name: VOLUME
              value: /data
          imagePullPolicy: Always
          volumeMounts:
            - mountPath: /data # shared vol with mc-server
              name: world-vol

    volumes:
      - name: world-vol # shared vol between containers. will not persist between restarts
        emptyDir: {}
```

[Full Java GameServer specification with world loading example](./example/mc-server.yml)

[Full Bedrock GameServer specification with world loading example](./example/mc-bedrock-server.yml)

### Run Locally with Docker

Run an example Minecraft GameServer Pod locally with `docker-compose`

```sh
docker-compose -f fileserver.docker-compose.yml up

# or

make docker-compose.load
```

<!-- ROADMAP -->

## Roadmap

See the [open issues](https://github.com/saulmaldonado/agones-mc/issues) for a list of proposed features (and known issues).

<!-- CONTRIBUTING -->

## Contributing

Contributions are what make the open source community such an amazing place to be learn, inspire, and create. Any contributions you make are **greatly appreciated**.

1. Fork the Project
2. Clone the Project
3. Create your Feature or Fix Branch (`git checkout -b (feat|fix)/AmazingFeatureOrFix`)
4. Commit your Changes (`git commit -m 'Add some AmazingFeatureOrFix'`)
5. Push to the Branch (`git push origin (feat|fix)/AmazingFeatureOrFix`)
6. Open a Pull Request

### Build from source

1. Clone the repo

   ```sh
   git clone https://github.com/saulmaldonado/agones-mc.git
   ```

2. Build

   ```sh
   make build

   # or

   CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o build/agones-mc .
   ```

### Build from Dockerfile

1. Clone the repo

   ```sh
   git clone https://github.com/saulmaldonado/agones-mc.git
   ```

2. Build

   ```sh
   docker build -t <hub-user>/agones-mc:latest .
   ```

3. Push to Docker repo

   ```sh
   docker push <hub-user>/agones-mc:latest
   ```

<!-- LICENSE -->

## License

Distributed under the MIT License. See [LICENSE](./LICENSE) for more information.

<!-- ACKNOWLEDGEMENTS -->

## Acknowledgements

- [Raqbit/mc-pinger](https://github.com/Raqbit/mc-pinger)
- [ZeroErrors/go-bedrockping](https://github.com/ZeroErrors/go-bedrockping)
- [itzg/docker-minecraft-server](https://github.com/itzg/docker-minecraft-server)
- [itzg/minecraft-bedrock-server](https://github.com/itzg/docker-minecraft-bedrock-server)

## Author

### Saul Maldonado

- 🐱 Github: [@saulmaldonado](https://github.com/saulmaldonado)
- 🤝 LinkedIn: [@saulmaldonado4](https://www.linkedin.com/in/saulmaldonado4/)
- 🐦 Twitter: [@saul_mal](https://twitter.com/saul_mal)
- 💻 Website: [saulmaldonado.com](https://saulmaldonado.com/)

## Show your support

Give a ⭐️ if this project helped you!

<!-- MARKDOWN LINKS & IMAGES -->
<!-- https://www.markdownguide.org/basic-syntax/#reference-style-links -->

[contributors-shield]: https://img.shields.io/github/contributors/saulmaldonado/agones-mc.svg?style=for-the-badge
[contributors-url]: https://github.com/saulmaldonado/agones-mc/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/saulmaldonado/agones-mc.svg?style=for-the-badge
[forks-url]: https://github.com/saulmaldonado/agones-mc/network/members
[stars-shield]: https://img.shields.io/github/stars/saulmaldonado/agones-mc.svg?style=for-the-badge
[stars-url]: https://github.com/saulmaldonado/agones-mc/stargazers
[issues-shield]: https://img.shields.io/github/issues/saulmaldonado/agones-mc.svg?style=for-the-badge
[issues-url]: https://github.com/saulmaldonado/agones-mc/issues
[license-shield]: https://img.shields.io/github/license/saulmaldonado/agones-mc.svg?style=for-the-badge
[license-url]: https://github.com/saulmaldonado/agones-mc/blob/master/LICENSE.txt
