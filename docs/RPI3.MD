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

TODO: Add root passwd condif and non-priv user add

6. When you reach ssh phase, lets setup podman and varlink
```
yum install --enablerepo=updates-testing podman libvarlink-util libvarlink containers-common     -y
systemctl enable io.podman.service
systemctl start io.podman.service
```

```
# as root
passwd
useradd rpi

# make user autologin
vi  /etc/lxdm/lxdm.conf
autologin=rpi
```

7. Disable first boot

```
sudo rm /etc/systemd/system/graphical.target.wants/initial-setup.service.
```

8. Configure graphical target

https://raspberrytips.com/install-fedora-raspberry-pi/

```
sudo dnf groupinstall -y "LXDE Desktop"
sudo systemctl set-default graphical.target
sudo rm /etc/systemd/system/graphical.target.wants/initial-setup.service.
```



7. Copy config files from dev machine to
destination
```
ssh root@192.168.10.177 "mkdir -p /etc/kubernetes/ /etc/vkubelet/"
scp deploy/systemd/vkubelet-podman.service root@192.168.10.177:/usr/lib/systemd/system/vkubelet-podman.service
scp ~/.kube/config root@192.168.10.177:/etc/kubernetes/admin.conf
scp deploy/systemd/podman-cfg.json root@192.168.10.177:/etc/vkubelet/podman-cfg.json
scp bin/virtual-kubelet root@192.168.10.177:/usr/local/bin/virtual-kubelet

```
