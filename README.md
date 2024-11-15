<!---
Copyright (c) [2024] Fergal Somers
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

     http://www.apache.org/licenses/LICENSE-2.0
 
 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
-->

# RouteLayer
A simple controller that configures Istio Routes to manage service isolation. 

# Context 
Systems are typically made up of many services. Best practice is to independently continuously build and deploy each service independently or each other (i.e. avoid testing and releasing all services in a single monolithic releaase). This provides a number of advantages: 

- Service can be released more frequently. 
- Teams can move independently, optimize their releases to ensure they can provide value quickly. 
- Since services are released more frequently, each update contains less changes and therefore less risk. 
- Services deployments can support A/B testing, progressive releases and rollbacks as risk mitigation strategies. 

Service testing is usually done at 3 levels: 

1. Unit-testing - generally happens within each service's repository as part of the Continuous Integration (CI) release process. 
2. Integration-testing - also happens within each service's repository again as part of the Continuous Integration (CI) release process. 
3. End-to-End testing - generally happens in some shared testing / QA environmment where multiple services are available to run end-to-end functional tests. 

The monolithic release process generally has a single QA environment for end-to-end testing. Each testing cycle (e.g. each week) release new Release Candidate (RC) versions of each service. Then test, and either rollback any RC versions that cause issues or fix-forward. At the end of the testing cycle the RC's are released to Staging / Production environments. This is a `maximum damage` approach to software release. Each week change everything. 

A better approach is independent service deployment, but where do you do end-to-end testing in this world. Ideally each service would like it's own QA environment that contains just the current production version of all the services. Then the service team could deploy the next version of the software into this QA environment and run the end-to-end tests and qualify the release. 

However, this quickly gets very costly if you are allocating each team or each PR it's own cluster / environment. A better approach is to build service isolation into your environment so that you can deploy N versions of a given service together. This allows each PR to share the same environment, which:

- reduces execution costs
- reduces operational overhead of maintaining multiple environments. 
- leads to faster testing cycles (less wait time to deploy software).

# RouteLayer 

RouteLayer provides a simple mechanism for service isolation within an Istio Mesh. It provides a simple controaller that manages Istio Virtual Services to provide service-routing capability. 
