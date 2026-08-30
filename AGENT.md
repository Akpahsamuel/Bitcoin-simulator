# AGENT.md — Agent & AI Assistant Startup Protocol

> **CRITICAL DIRECTIVE FOR ALL AI ASSISTANTS & AGENTS:**
> At the start of every session, conversation, or task in this repository, **you MUST immediately read and follow [`CLAUDE.md`](./CLAUDE.md)** before planning, executing commands, editing code, or responding to the user.

---

## 1. Startup Checklist

Whenever an agent session begins:

1. **Read [`CLAUDE.md`](./CLAUDE.md)**:
   - Understand the repository architecture, locked design decisions (D1–D7), and golden rules.
   - Familiarize yourself with the **Language Port Contract** before authoring or modifying any example in `examples/`.
   - Adhere to the strict Git commit & signing policies (all commits must be signed with `-S`, never push directly to `main`, never add AI co-author trailers).

2. **Verify Working Environment**:
   - Bitcoin Core version is pinned to **31.1** in **regtest** mode.
   - Credentials (`bitcoinrpc:bitcoinrpcpassword`) and ports (`18443`, `28332`, `28333`, `18444`) are standard and regtest-only.
   - Do not connect to mainnet, testnet, or signet under any circumstances.
   - Ensure the node boots at height 0 (never mine inside `init-lab.sh`; only the example bootstrap mines 101 blocks for coinbase maturity).

3. **Check Git Status & Branch**:
   - Ensure you are working on a feature branch (not `main`).
   - Check if `.env` exists; if missing, run `bash scripts/init-lab.sh` or copy from `.env.example`.

---

## 2. Quick Reference Links

- **Working Rules & Invariants**: [`CLAUDE.md`](./CLAUDE.md)
- **Roadmap & Master Plan**: `plan.md` (local / git-ignored)
- **Contributing & Commit Signing Guide**: [`CONTRIBUTING.md`](./CONTRIBUTING.md)
- **Troubleshooting**: [`docs/troubleshooting.md`](./docs/troubleshooting.md)
- **CLI Workshop Walkthrough**: [`examples/00-cli-workshop/README.md`](./examples/00-cli-workshop/README.md)
