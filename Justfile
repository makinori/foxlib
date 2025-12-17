default:
	@just --list

# build and publish base image
container:
	cat base.Dockerfile | podman build -t ghcr.io/makinori/foxlib:base -
	podman push ghcr.io/makinori/foxlib:base