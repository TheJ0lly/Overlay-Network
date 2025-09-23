import subprocess
import os
from enum import Enum
import sys
import re

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


def remove_file(filepath: str) -> int:
    status = subprocess.run(["rm", filepath])
    if status.returncode != 0:
        log_warn(f"failed to remove file: {filepath} - please do it manually")
        return status.returncode
    return 0

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
            log_error(f"failed to compile binary: {file} - stopping build")
            os._exit(status.returncode)
        log_success(f"compiled binary succesfully: {file}")

    log_success("all binaries have been built")


def clean():
    files = os.listdir("./bin")
    for file in files:
        status = remove_file(f"./bin/{file}")
        if status != 0:
            log_error("failed to clean build folder - please do it manually")
            os._exit(status)
    log_success("build folder cleaned successfully")


def coverage():
    status = subprocess.run(["go", "test", "./...", "-coverprofile=coverage.out"], capture_output=True)
    if status.returncode != 0:
        log_error("failed to generate the coverage report")
        os._exit(1)

    cov_output = str(status.stdout)
    coverages = re.findall(rf'capture:\s(\d+\.\d)%', cov_output)
    all_covered: bool = all(cov == "100.0" for cov in coverages)

    if all_covered == True:
        log_success(r"source code is 100% covered - no need for coverage report")
        remove_file("./coverage.out")
        os._exit(0)

    status = subprocess.run(["go", "tool", "cover", "-html=coverage.out", "-o", "coverage.html"])
    if status.returncode != 0:
        log_error("failed to convert the coverage report to HTML")
        os._exit(1)

    status = subprocess.run(["rm", "coverage.out"])
    if status.returncode != 0:
        log_error("failed to remove the original coverage report")
        os._exit(1)


if len(sys.argv) != 2:
    log_error("build script requires one argument:\n\tbuild\n\tclean\n\ttest(not working yet :))")
    os._exit(1)

match sys.argv[1]:
    case "build":
        build()
    case "clean":
        clean()
    case "coverage":
        coverage()
    case "test":
        print("testing not working yet")
