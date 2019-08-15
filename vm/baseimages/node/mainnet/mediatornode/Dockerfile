FROM ubuntu:18.04

MAINTAINER palletone "contract@pallet.one"

ADD ./tardata/* /palletone/

RUN apt-get update && apt-get install -y vim

WORKDIR /palletone/mainnet

EXPOSE 8545
