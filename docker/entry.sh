#!/bin/bash

# set the config directory
datadir="/opt/data"

# create config dir
mkdir -p "${datadir}"

# check for config file
if [ ! -f "${datadir}/config.toml" ]; then
  echo "Config File not found, adding default."
  cp "/docker/config.toml" "${datadir}/config.toml"
  # generate new seed
  seed=$(accept-nano -seed)
  echo "# Your generated seed" >> "${datadir}/config.toml"
  echo "Seed = \"${seed}\"" >> "${datadir}/config.toml"
fi

# start service
accept-nano -config /opt/data/config.toml
