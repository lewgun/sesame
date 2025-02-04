#! /usr/bin/env bash

# Copyright Project Sesame Authors
#
# Licensed under the Apache License, Version 2.0 (the "License"); you may
# not use this file except in compliance with the License.  You may obtain
# a copy of the License at
#
# http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
# WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.  See the
# License for the specific language governing permissions and limitations
# under the License.

# install-sesame-release.sh: Install a specific release of Sesame.

set -o pipefail
set -o errexit
set -o nounset

readonly KUBECTL=${KUBECTL:-kubectl}
readonly WAITTIME=${WAITTIME:-5m}

readonly PROGNAME=$(basename "$0")
readonly VERS=${1:-}

if [ -z "$VERS" ] ; then
        printf "Usage: %s VERSION\n" $PROGNAME
        exit 1
fi

# Install the Sesame version.
${KUBECTL} apply -f "https://projectsesame.io/quickstart/$VERS/sesame.yaml"

${KUBECTL} wait --timeout="${WAITTIME}" -n projectsesame -l app=sesame deployments --for=condition=Available
${KUBECTL} wait --timeout="${WAITTIME}" -n projectsesame -l app=envoy pods --for=condition=Ready
