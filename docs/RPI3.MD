# Raspbery PI 3 

This document explains how to setup RPI3 to be used as edge node for podman virtual kubelet

## Install 

1. [Download](https://arm.fedoraproject.org/) Fedora ARM minial image
2. Install [ARM installer](https://fedoraproject.org/wiki/Architectures/ARM/Installation)
3. "Bake ARM image into RPI SD Card:
```
arm-image-installer --addkey=/home/$USER/.ssh/id_rsa.pub \
--target=rpi3 --media=/dev/sda \
--image=/home/mjudeiki/Documents/arm/Fedora-Server-armhfp-31-1.9-sda.raw.xz \
--resizefs
```
4. Mount root filesystem of the SD card with R/W 
cp ./deploy/misc/nodes/rpi3/* /run/media/$USER/__/etc/sysconfig/network-scripts/
rm /run/media/$USER/__/etc/systemd/system/multi-user.target.wants/initial-setup.service

5. Feel free to play with shadow files and all other configs

6. When you reach ssh phase, lets setup podman and varlink
```
yum install --enablerepo=updates-testing podman libvarlink-util libvarlink -y
systemctl enable io.podman.service
systemctl start io.podman.service
```