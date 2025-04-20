#!/usr/bin/env bash

go build .
./infiniband_exporter -w $(pwd) -m dev
