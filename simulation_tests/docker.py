import subprocess
import os

def is_docker_installed() -> bool:
    cmd = ["docker"]
    status = subprocess.run(cmd, capture_output=True, text=True)
    if status.returncode != 0:
        print(f"It seems that Docker is not installed on the system:\n\n{status.stdout}")
        return False
    return True

def is_docker_network_exists() -> bool:
    cmd = ["docker", "network", "ls"]
    status = subprocess.run(cmd, capture_output=True, text=True)

    if status.returncode != 0:
        print(f"Cannot show the Docker networks:\n\n{status.stdout}")
        return False
    return status.stdout.find("OverlayNetwork") != -1

def docker_create_network() -> bool:
    cmd = ["docker", "network", "create", "-d", "bridge", "OverlayNetwork"]
    status = subprocess.run(cmd, capture_output=True, text=True)

    if status.returncode != 0:
        print(f"Failed to create Docker network used for testing:\n\n{status.stdout}")
        return False
    return True

if not is_docker_installed():
    os._exit(1)

if not is_docker_network_exists():
    if not docker_create_network():
        os._exit(1)