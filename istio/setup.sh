# Copyright (c) [2024] Fergal Somers
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
# 
#      http://www.apache.org/licenses/LICENSE-2.0
#  
#  Unless required by applicable law or agreed to in writing, software
#  distributed under the License is distributed on an "AS IS" BASIS,
#  WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
#  See the License for the specific language governing permissions and
#  limitations under the License.

cd "$(dirname "$0")"

docker pull istio/istioctl:1.23.0

docker run -v $PWD/..:/wd \
 -e KUBECONFIG=/wd/kubeconfig \
    --network=host \
 istio/istioctl:1.23.0 \
 install -f /wd/istio/istio-profile.yaml -y

kubectl --kubeconfig=$PWD/../kubeconfig apply -k resources/overrides
