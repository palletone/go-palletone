# Build Gptn in a stock Go builder container
FROM palletone/middle:step2 as builder

# Pull Gptn into a second stage deploy alpine container
FROM ubuntu:latest

# Maintainer information
MAINTAINER palletone "contract@pallet.one"

# Copy gptn from stock Go builder container
COPY --from=builder /go-palletone/images/scripts/result /go-palletone

# Working in the go-palletone directory
WORKDIR /go-palletone

# Exposing 8545 8546 30303 30303/udp ports
#EXPOSE 8545 8546 30303 30303/udp
#
## Default start command
#ENTRYPOINT ["gptn"]
