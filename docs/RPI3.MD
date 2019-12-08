# Raspberry PI 3

This document explains how to setup RPI3 to be used as edge node for
podman virtual kubelet node

## Install

1. [Download](https://arm.fedoraproject.org/) Fedora ARM minial image.
   * Note: Fedora 31 is not supported on RPI3 and RPI4. Only valid configuration currently is RPI3 & Fedora 30*
2. Install [ARM installer](https://fedoraproject.org/wiki/Architectures/ARM/Installation)
3. Bake ARM image into RPI SD Card:

```
arm-image-installer --addkey=/home/$USER/.ssh/id_rsa.pub \
--target=rpi3 --media=/dev/sda \
--image=/home/$USER/Documents/arm/Fedora-Server-armhfp-30-1.2-sda.raw.xz \
--resizefs
```

4. Mount root filesystem of the SD card with R/W and change static files to boot
without promts

```
# copy network scripts, so WIFI would be "like home"
cp ./deploy/misc/nodes/rpi3/* /run/media/$USER/__/etc/sysconfig/network-scripts/
```

Set Hostname
```
echo "rpi3-walldisplay" > /etc/hostname

```

Copy boostrap script and systemd file. This is automated way to bootstrap the node
```
cat <<EOF >/usr/lib/systemd/system/bootstrap.service
[Unit]

[Service]
ExecStart=curl https://gist.githubusercontent.com/mjudeikis/868dbd82bd738e1e926b7c1b25c9555d/raw/7cb063747b1fd80b5f132e879f060cdc0353b681/gistfile1.txt | sh

[Install]
WantedBy=default.target
EOF
```
Enable service:
```

```

5. Feel free to play with shadow files and all other configs

TODO: Add root passwd config and non-priv user add

6. When you reach ssh phase, lets setup podman and varlink.
   TODO: Move all this to boostrap script
```
yum install --enablerepo=updates-testing podman libvarlink-util libvarlink containers-common     -y
systemctl enable io.podman.service
systemctl start io.podman.service
```

Configure root and RPI users
```
# as root
passwd
useradd rpi

# make user autologin
vi  /etc/lxdm/lxdm.conf
autologin=rpi
```

7. Configure graphical target

https://raspberrytips.com/install-fedora-raspberry-pi/

```
sudo dnf groupinstall -y "LXDE Desktop"
sudo systemctl set-default graphical.target
sudo systemctl mask sleep.target suspend.target hibernate.target hybrid-sleep.target
```

8. Copy config files from dev machine to destination
TODO: Move to boostrap script
```
ssh root@192.168.10.177 "mkdir -p /etc/kubernetes/ /etc/vkubelet/"
scp deploy/systemd/vkubelet-podman.service root@192.168.10.177:/usr/lib/systemd/system/vkubelet-podman.service
scp ~/.kube/config root@192.168.10.177:/etc/kubernetes/admin.conf
scp deploy/systemd/podman-cfg.json root@192.168.10.177:/etc/vkubelet/podman-cfg.json
scp bin/virtual-kubelet root@192.168.10.177:/usr/local/bin/virtual-kubelet

```
