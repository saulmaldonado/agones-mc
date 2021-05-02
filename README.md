[![Contributors][contributors-shield]][contributors-url]
[![Forks][forks-shield]][forks-url]
[![Stargazers][stars-shield]][stars-url]
[![Issues][issues-shield]][issues-url]

<!-- PROJECT LOGO -->
<br />
<p align="center">
  <h3 align="center">agones-mc-monitor</h3>

  <p align="center">
  Minecraft server monitor sidecar for Agones GameServers
    <br />
    <a href="https://github.com/saulmaldonado/agones-mc-monitor"><strong>Explore the docs »</strong></a>
    <br />
    <br />
    <a href="https://github.com/saulmaldonado/agones-mc-monitor/tree/main/example/mc-server.yml">View Example</a>
    ·
    <a href="https://github.com/saulmaldonado/agones-mc-monitor/issues">Report Bug</a>
    ·
    <a href="https://github.com/saulmaldonado/agones-mc-monitor/issues">Request Feature</a>
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

This application is meant to be run as a sidecar container alongside Minecraft servers in an Agones GameServer pod. This application integrates the Agones SDK with the Minecraft server and assists with lifecycle management such as health checking, allocations, and shutdowns.

### Built With

- [mc-pinger](https://github.com/Raqbit/mc-pinger)
- [go-bedrockping](https://github.com/ZeroErrors/go-bedrockping)
- [Agones Go SDK](agones.dev/agones/sdks/go)

<!-- GETTING STARTED -->

## Getting Started

To get a copy up and running follow these steps.

### Prerequisites

This is an example of how to list things you need to use the software and how to install them.

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

#### Create a new Minecraft GameServer

```sh
kubectl create -f example/mc-server.yml
```

[Full GameServer specification example](./example/mc-server.yml)

<!-- USAGE EXAMPLES -->

## Usage

### Agones GameServer

A GameServer requires a game server container. This will be the Mincraft server container. Every GameServer also contains the an sdkServer that will report lifecycle changes to the Agones controller. To signal Minecraft server lifecycle updates, a seperate application integrated with the Agones SDK needs to ping the Minecraft server and report lifecycle updates to the sdkServer.

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
        image: saulmaldonado/agones-mc-monitor
        imagePullPolicy: Always
```

[Full Java GameServer specification example](./example/mc-server.yml)

[Full Bedrock GameServer specification example](./example/mc-bedrock-server.yml)

### Run Locally with Docker

Run alongside a Minecraft server and an [Agones SDK sidecar](https://agones.dev/site/docs/guides/client-sdks/local/) in the same network

```sh
docker run -it --rm saulmaldonado/agones-mc-monitor
```

### Flags

```
  --attempts uint
        Ping attempt limit. Process will end after failing the last attempt (default 5)

  --edition string
      Minecraft server edition. java or bedrock (default "java")

  --host string
        Minecraft server host (default "localhost")

  --initial-delay duration
        Initial startup delay before first ping (default 1m0s)

  --interval duration
        Server ping interval (default 10s)

  --port uint
        Minecraft server port (default 25565)

  --timeout duration
        Ping timeout (default 10s)
```

<!-- ROADMAP -->

## Roadmap

See the [open issues](https://github.com/saulmaldonado/agones-mc-monitor/issues) for a list of proposed features (and known issues).

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
   git clone https://github.com/saulmaldonado/agones-mc-monitor.git
   ```

2. Build

   ```sh
   go build -o agones-mc-monitor ./cmd/main.go
   ```

### Build from Dockerfile

1. Clone the repo

   ```sh
   git clone https://github.com/saulmaldonado/agones-mc-monitor.git
   ```

2. Build

   ```sh
   docker build -t <hub-user>/agones-mc-monitor:latest .
   ```

3. Push to Docker repo

   ```sh
   docker push <hub-user>/agones-mc-monitor:latest
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

[contributors-shield]: https://img.shields.io/github/contributors/saulmaldonado/agones-mc-monitor.svg?style=for-the-badge
[contributors-url]: https://github.com/saulmaldonado/agones-mc-monitor/graphs/contributors
[forks-shield]: https://img.shields.io/github/forks/saulmaldonado/agones-mc-monitor.svg?style=for-the-badge
[forks-url]: https://github.com/saulmaldonado/agones-mc-monitor/network/members
[stars-shield]: https://img.shields.io/github/stars/saulmaldonado/agones-mc-monitor.svg?style=for-the-badge
[stars-url]: https://github.com/saulmaldonado/agones-mc-monitor/stargazers
[issues-shield]: https://img.shields.io/github/issues/saulmaldonado/agones-mc-monitor.svg?style=for-the-badge
[issues-url]: https://github.com/saulmaldonado/agones-mc-monitor/issues
[license-shield]: https://img.shields.io/github/license/saulmaldonado/agones-mc-monitor.svg?style=for-the-badge
[license-url]: https://github.com/saulmaldonado/agones-mc-monitor/blob/master/LICENSE.txt
