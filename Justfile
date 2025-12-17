default:
	@just --list

# build and publish base image
container:
	cat base.Dockerfile | podman build --dns 1.1.1.1 \
	-t ghcr.io/makinori/foxlib:base -
	podman push ghcr.io/makinori/foxlib:base