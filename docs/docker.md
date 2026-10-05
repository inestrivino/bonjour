# Docker

> Learn how to effectively use Docker alongside `bonjour`. Run it, test it, and create releases.

## Trying bonjour with Docker

You can try the application in your terminal without doing a full installation or storing persistent data using the official Docker image of the project:

```bash
docker run --rm -it ghcr.io/inestrivino/bonjour:latest
```

Note: Any configuration filled out during an ephemeral run will disappear as soon as the container exits.

## Persistently running bonjour with Docker

If you want to use this project long-term via Docker, and keep your saved configuration intact:

1. Create and name the container instance

```bash
docker run -it --name bonjour ghcr.io/inestrivino/bonjour:latest
```

2. Start and re-attach your instance on subsequent runs:

```bash
docker start -ai bonjour
```

## Creating a local Docker image of the project

You can create a local Docker image of the project for personal use or testing, by following these two steps (also present in `docs/development.md`):

* (**GoReleaser necessary**) Creating the installation binaries for your OS and architecture:

```bash
goreleaser build --snapshot --single-target --clean -p 1
```

* (**GoReleaser and Docker necessary**) Create a local Docker image from the installation binaries built in the previous point:

```bash
docker build --build-arg TARGETPLATFORM=$(find dist -type d -name "bonjour_*" | head -n 1) -t bonjour:local .
docker run --rm -it bonjour:local
```

Now you can run the image or even create a container instance as before for persistent data.

## Deleting bonjour-related images and data

If you want to completely remove the Docker image, the container instance, and all image-related data in your computer, follow these steps:

1. Make sure the container instance is stopped.
2. Run this command to delete the container instance:

```bash
docker rm -f bonjour
```

Verification: Running `docker ps -a | grep my-bonjour` should return nothing.

3. Run the command to delete all bonjour-related images:

```bash
docker rmi -f $(docker images --format '{{.Repository}}:{{.Tag}}' | grep '^bonjour')
```

Verification: Running `docker images | grep bonjour` should return nothing.

**WARNING: THE FOLLOWING STEP WILL DELETE ALL DOCKER IMAGE RELATED DATA FROM YOUR COMPUTER, NOT JUST BONJOUR RELATED ONES. THIS IS A DESTRUCTIVE ACTION AND CAN'T BE UNDONE!**

If you want to delete all information associated to your docker images then run:

```bash
docker system prune -a --volumes -f
```

Verification: Check that there is no space still occupied with `docker system df`.