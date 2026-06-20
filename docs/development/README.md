# Development Documentation

This folder contains engineering and agent-facing guidance for building Mercadia POS.

## Documents

- [AI agent rules](ai-agent-rules.md) - rules for Codex/AI agents working in the repository.
- [API contract workflow](api-contract-workflow.md) - OpenAPI generation, Scalar docs, and Orval client generation.
- [Repository structure](repository-structure.md) - POS folder ownership and app/service/package layout.
- [Dependency and version policy](dependency-policy.md) - current, maintained, and secure component rules.
- [Continuous integration](ci.md) - GitHub Actions jobs, path filters, and branch protection.
- [Documentation backlog](documentation-backlog.md) - important topics that are not yet described deeply enough.
- [UI components and theme](ui-components.md) - `@mercadia/ui` package, token layers, and accent presets.

The short entry point for automated coding agents is:

- `../../AGENTS.md` from the POS project root.

Build and contract commands are documented in `../../README.md`.

These documents are intentionally practical. They should be updated whenever architecture
decisions change or implementation reveals a better workflow.
