This section outlines production-ready deployment steps for cloud providers, RBAC, and CI/CD. Use in docs or PRs for reviewers.

- Production install scripts:
  - deploy/linux_production_install.sh
  - deploy/windows_production_install.ps1
- Systemd service:
  - deploy/systemd/novabackup.service
- Docker Compose for prod:
  - docker-compose-prod.yml (to start API + UI + DB)
- Docker/CI: GH Actions pipelines for DB dialects and cloud
- Documentation: migration guide, cloud provider examples, RBAC docs
