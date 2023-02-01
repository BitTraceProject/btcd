FROM ubuntu:20.04

WORKDIR /bittrace

COPY ./output/* /bittrace/

VOLUME ["/root/.bittrace"]
VOLUME ["/root/.btcd"]


# for testnet
EXPOSE 18555 18556
# for mainnet
EXPOSE 8333 8334
# for simnet
EXPOSE 18333 18334
# for signet
EXPOSE 38333 38334

ENTRYPOINT ["/bittrace/btcd"]