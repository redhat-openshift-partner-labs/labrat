# LABRAT Documentation

Welcome to the LABRAT documentation. This directory contains comprehensive guides for users, developers, and administrators.

## Documentation Structure

### Commands

User-facing command reference documentation.

- [hub managedclusters](commands/hub-managedclusters.md) - List ACM managed clusters

### Architecture

Technical architecture and design documentation.

- [Kubernetes Integration](architecture/kubernetes-integration.md) - K8s client architecture and design

### Development

Developer guides for contributing to LABRAT.

- [Hub ManagedClusters Implementation](development/hub-managedclusters-implementation.md) - Implementation guide for the managedclusters command

## Quick Links

### For Users

- **Getting Started**: See main [README.md](../README.md)
- **Configuration**: `~/.labrat/config.yaml` setup
- **Commands**: Browse [commands/](commands/) directory

### For Developers

- **Implementation Plans**: See [`.claude/plans/`](../.claude/plans/)
- **Testing Guide**: [TESTING.md](../TESTING.md)
- **Linting Guide**: [LINTING.md](../LINTING.md)
- **Code Style**: Follow Go conventions and project linting rules

### For Contributors

- **Development Setup**: Install Go 1.25+, Task, and run `task init`
- **TDD Workflow**: Use `task test:watch` for continuous testing
- **Quality Gates**: Run `task check` before committing

## Command Documentation

All CLI commands have detailed documentation in the [commands/](commands/) directory:

| Command | Description | Status |
|---------|-------------|--------|
| `hub status` | Check ACM hub health | Implemented (placeholder) |
| `hub managedclusters` | List managed clusters | In Development |
| `spoke create` | Provision partner cluster | Implemented (placeholder) |
| `bootstrap init` | Initialize environment | Implemented (placeholder) |

## Architecture Documentation

Technical design documents explaining LABRAT's architecture:

| Document | Description |
|----------|-------------|
| [Kubernetes Integration](architecture/kubernetes-integration.md) | K8s client design, CRD handling, status derivation |

## Development Documentation

Guides for implementing features and contributing code:

| Document | Description |
|----------|-------------|
| [Hub ManagedClusters Implementation](development/hub-managedclusters-implementation.md) | TDD guide for managedclusters command |

## Contributing to Documentation

### Adding New Documentation

1. **Command Docs**: Create `commands/<command-name>.md`
   - Include synopsis, description, flags, examples
   - Follow the template in `hub-managedclusters.md`

2. **Architecture Docs**: Create `architecture/<topic>.md`
   - Explain design decisions
   - Include diagrams (Mermaid or ASCII)
   - Document trade-offs

3. **Development Docs**: Create `development/<feature>.md`
   - TDD workflow
   - Implementation phases
   - Testing strategy

### Documentation Standards

- **Markdown**: Use GitHub-flavored Markdown
- **Code Blocks**: Always specify language (```go, ```bash, ```yaml)
- **Links**: Use relative paths for internal links
- **Examples**: Include working, copy-pasteable examples
- **Clarity**: Write for the intended audience (users vs developers)

### Building Documentation Site (Future)

Currently, documentation is in Markdown files. In the future, we may:
- Use MkDocs or similar for a documentation site
- Add API documentation generation
- Include auto-generated command reference

## Documentation TODO

- [ ] Add configuration reference guide
- [ ] Document RBAC requirements for ACM
- [ ] Add troubleshooting guide
- [ ] Create architecture diagrams
- [ ] Add API documentation (when API is implemented)
- [ ] Create video tutorials or screencasts
- [ ] Add FAQ section
- [ ] Document deployment strategies

## Feedback

Found a documentation issue? Please:
1. Check if it's already reported in GitHub Issues
2. Create a new issue with label `documentation`
3. Or submit a PR with the fix

## License

Documentation is licensed under the same terms as the LABRAT project.
