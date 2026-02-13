# Standard Application Reference Architecture

> Parent Project: [Enterprise GitOps Platform Blueprint](../gitops-platform-blueprint)

## Overview
This repository demonstrates the implementation of a high-performance microservice within the GitOps ecosystem. It includes the application source code, comprehensive test suites, and the CI pipeline configuration.

## Key Features
- **Application Stack**: Built on Kotlin and Quarkus for cloud-native performance.
- **Containerization**: Optimized multi-stage Docker builds for minimal image size.
- **CI/CD Integration**: Automated pipeline that builds images, scans for vulnerabilities, and updates the `gitops-manifests` repository to trigger deployment.

---
**Component**: Application Reference  
**Status**: Reference Implementation