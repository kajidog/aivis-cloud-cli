---
"@kajidog/aivis-cloud-cli": patch
---

Optimize CI/CD performance and reduce unnecessary testing

- Replace slow cross-platform build test with fast single-platform validation
- Simplify release workflow by removing complex retry logic
- Focus PR tests on essential functionality: Go client tests + CLI build verification
- Reduce CI execution time while maintaining code quality assurance