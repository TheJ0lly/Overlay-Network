import subprocess
import os
from enum import Enum
import sys


class Color(Enum):
    WHITE=30
    RED=31
    GREEN=32
    YELLOW=33

def print_with_color(message: str, endline: str = "", bold: bool = False, color: Color = Color.WHITE):
    print(f"\033[{color.value};{1 if bold is True else ""}m{message}\033[0m", end=endline)

def log_info(message: str):
    print_with_color(message="INFO", bold=True, endline="|")
    print(f"{message}")

def log_warn(message: str):
    print_with_color(message="WARN", bold=True, endline="|", color=Color.YELLOW)
    print(f"{message}")

def log_error(message: str):
    print_with_color(message="ERRO", bold=True, endline="|", color=Color.RED)
    print(f"{message}")

def log_success(message: str):
    print_with_color(message="SUCD", bold=True, endline="|", color=Color.GREEN)
    print(f"{message}")




def build():
    log_info("reading the cmd folder")
    files = os.listdir("./cmd")
    if len(files) == 0:
        log_error("something has gone VERY WRONG: there are no folders in the cmd folder")
        os._exit(1)

    for file in os.listdir("./cmd"):
        log_info(f"compiling binary: {file}")
        status = subprocess.run(["go", "build", "-o", "./bin/", f"./cmd/{file}"])
        if status.returncode != 0:
            log_warn(f"failed to compile binary: {file} - stopping build")
            os._exit(1)
        log_success(f"compiled binary succesfully: {file}")

    log_success("all binaries have been built")

def clean():
    files = os.listdir("./bin")
    for file in files:
        status = subprocess.run(["rm", f"./bin/{file}"])
        if status.returncode != 0:
            log_error("failed to clean build folder - please do it manually")
            os._exit(1)
    log_success("build folder cleaned successfully")

if len(sys.argv) != 2:
    log_error("build script requires one argument:\n\tbuild\n\tclean\n\ttest(not working yet :))")
    os._exit(1)

match sys.argv[1]:
    case "build":
        build()
    case "clean":
        clean()
    case "test":
        print("testing not working yet")
