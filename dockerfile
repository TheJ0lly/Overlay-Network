# We get the official Go:1.24.7 version image
FROM golang:1.24.7 AS builder

# We copy the mod and sum files and download dependencies
WORKDIR /Overlay-Network
# COPY go.mod ./
# RUN go mod download

# We copy the source code into Docker WORKDIR
COPY . .
RUN python3 build.py -build

# We install an OS image so that Go and Docker have a filesystem
FROM fedora:42

# Install only runtime dependencies - not required yet
# RUN dnf install update

WORKDIR /Overlay-Network

# Copy the compiled binary from the builder
COPY --from=builder /Overlay-Network/bin/nodemgmt .

# Ensure binary is executable
RUN chmod +x ./nodemgmt

# Run the application
CMD ["./nodemgmt"]