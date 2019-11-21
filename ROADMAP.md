# Roadmap for Podman Virtual Kubelet provider

## Easy

1. Move all errors to errdefs
2. Sometimes podman operations hangs. Add context to all operations with timeouts
3. Add support for 2 containers in the ppodman pod

## Not so easy

1. Implement `GetContainerLogs` `RunInContainer`
2. Add better kube yaml file support. Like enable volumes, secrets, configmaps.
3. Add routine to check if podman is alive and update node status based on that
4. Configure node and schedule pods based on configures limits
5. Add support for "remote vkubelet podman" where vkubelet is running in the cluster as a pod
and it reaches to podman node via remote varlink api
6. Add support for better pod status. Like "CrashBackLoop", "Failed", "ImageNotFound", etc.