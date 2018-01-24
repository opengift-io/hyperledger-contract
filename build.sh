#!/usr/bin/env bash
rm -rf build
cd contracts
solc --abi --optimize --bin --overwrite *.sol -o ../build