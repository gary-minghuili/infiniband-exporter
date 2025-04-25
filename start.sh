#!/usr/bin/env bash

go build .

# ./infiniband_exporter -w $(pwd) -m dev -g true

# cp ./config/sequoia_config.yaml ./config/config.yaml
# cp ./config/sequoia_default.yaml ./config/default.yaml

./infiniband_exporter -w $(pwd) -m dev
