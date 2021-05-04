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

[Full GameServer specification example](./example/mc-server.yml)

<!-- USAGE EXAMPLES -->

## Usage

### Agones GameServer

A Minecraft server container will serve as the GameServer container. Every GameServer Pod also contains the an sdkServer that will report lifecycle changes to the Agones controller. To signal Minecraft server lifecycle updates, a seperate application integrated with the Agones SDK will ping the Minecraft server and report lifecycle updates to the sdkServer.

#### GameServer Pod template example

```yml
template:
  spec:
    containers:
      - name: mc-server
        image: itzg/minecraft-server # Minecraft server image
        imagePullPolicy: Always
        env: # Full list of ENV variables at https://github.com/itzg/docker-minecraft-server
          - name: EULA
            value: 'TRUE'

      - name: mc-monitor
        image: saulmaldonado/agones-mc
        args:
          - monitor
        imagePullPolicy: Always
```

[Full Java GameServer specification example](./example/mc-server.yml)

[Full Bedrock GameServer specification example](./example/mc-bedrock-server.yml)

### Run Locally with Docker

Run an example Minecraft GameServer Pod locally with `docker-compose`

```sh
docker-compose -f monitor.docker-compose.yml up

# or

make docker-compose.monitor
```

### Flags

```
Usage:
  agones-mc monitor [flags]

Flags:
--attempts uint            Ping attempt limit. Process will end after failing the last attempt (default 5)

--edition string           Minecraft server edition. java or bedrock (default "java")

-h, --help                     help for monitor

--host string              Minecraft server host (default "localhost")

--initial-delay duration   Initial startup delay before first ping (default 1m0s)

--interval duration        Server ping interval (default 10s)

--port uint                Minecraft server port (default 25565)
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
