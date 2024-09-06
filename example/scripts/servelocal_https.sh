#!/bin/bash

set -euxo pipefail

script_dir="$(dirname $0)"
"${script_dir}/../../../gohttp/simple_serve.sh" "${script_dir}/../output"
