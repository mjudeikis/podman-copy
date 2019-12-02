build:
	GOARCH=arm go build -o demo

image:
	podman build -t quay.io/mangirdas/midori:arm .

image-push: build image
	podman push quay.io/mangirdas/midori:arm
