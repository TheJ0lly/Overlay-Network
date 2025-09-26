import subprocess
import os
from enum import Enum
import sys
import re
import argparse

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

def unit_testing(verbose: bool):
    subprocess.run(["go", "clean", "-fuzzcache"])
    subprocess.run(["go", "clean", "-testcache"])
    subprocess.run(["go", "clean", "-cache"])

    cmd = ["go", "test", "./..."]

    if verbose:
        cmd.append("-v")

    status = subprocess.run(cmd)
    if status.returncode != 0:
        log_error("failed to run the unit tests")
        os._exit(1)


parser = argparse.ArgumentParser(prog="Overlay-Network control script",
                                 description="This script helps with building/cleaning the binaries, generating coverage reports, and testing, both UT's and simulations(not yet)")
parser.add_argument("-build", action="store_true")
parser.add_argument("-clean", action="store_true")
parser.add_argument("-cover", action="store_true")

testing_sub_parser = parser.add_subparsers(dest="testmode")
test_parser = testing_sub_parser.add_parser("test", help="run tests")
test_parser.add_argument("type", choices=["ut", "nt"], help="ut(Unit Testing) | nt(Network Testing)")
test_parser.add_argument("-v", "--verbose", action="store_true", help="enables verbose output")


parser.add_argument("-test", type=str, help="the only available choices are: ut(Unit Tests)/nt(Network Tests)")

args = parser.parse_args()

if args.build:
    build()
elif args.clean:
    clean()
elif args.cover:
    coverage()
elif args.testmode == "test":
    if args.type == "ut":
        unit_testing(args.verbose)
    elif args.type == "nt":
        print("not working yet")
    else:
        print("unknown test flag value")