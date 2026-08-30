# Contributing to Bitcoin Simulator

Thank you for your interest in contributing! This project is a hands-on sandbox and starter kit for Bitcoin developers.

To maintain code quality and security, our repository enforces strict GitHub repository rules. Please review these guidelines before submitting a Pull Request.

---

## 🔒 Mandatory Repository Requirements

The `main` branch is protected by automated rules:
1. **Pull Requests Only**: Direct pushes to `main` are rejected. All changes must go through a PR.
2. **Verified Commits Required**: **Every commit in your PR must be cryptographically signed and verified by GitHub.** Commits without valid signatures will block merging.

---

## 🔑 Setting Up Verified Commit Signing

GitHub supports signing commits using either **SSH keys** (recommended, easiest) or **GPG keys**.

### Option A: Using SSH Signing (Recommended & Fastest)

If you already have an SSH key used for GitHub authentication, you can use it to sign commits in Git 2.34+:

1. **Check if you have an SSH key:**
   ```bash
   ls -la ~/.ssh/*.pub
   ```
   *(If you don't have one, generate one with `ssh-keygen -t ed25519 -C "your_email@example.com"`)*

2. **Configure Git to use SSH for signing:**
   ```bash
   git config --global gpg.format ssh
   git config --global user.signingkey ~/.ssh/id_ed25519.pub
   git config --global commit.gpgsign true
   ```

3. **Add your SSH key to GitHub as a "Signing Key":**
   - Copy your public key:
     ```bash
     cat ~/.ssh/id_ed25519.pub
     ```
   - Go to GitHub: **Settings** → **SSH and GPG keys** → **New SSH Key**.
   - Set **Key type** to **"Signing Key"** (or add it as both Authentication and Signing).
   - Paste your key and save.

---

### Option B: Using GPG Signing

1. **Generate a GPG key:**
   ```bash
   gpg --full-generate-key
   ```
   *(Choose RSA and RSA, 4096 bits, and enter the exact name and email associated with your GitHub account).*

2. **List your key ID:**
   ```bash
   gpg --list-secret-keys --keyid-format=long
   ```
   Look for the `sec` line, e.g. `sec rsa4096/3AA5C34371567BD2`. The ID is `3AA5C34371567BD2`.

3. **Configure Git to sign with your GPG key:**
   ```bash
   git config --global user.signingkey <YOUR_KEY_ID>
   git config --global commit.gpgsign true
   ```

4. **Export and add your public GPG key to GitHub:**
   ```bash
   gpg --armor --export <YOUR_KEY_ID>
   ```
   Copy the output block (`-----BEGIN PGP PUBLIC KEY BLOCK-----` to `-----END PGP PUBLIC KEY BLOCK-----`) and add it under GitHub **Settings** → **SSH and GPG keys** → **New GPG Key**.

---

## 🛠 Fixing Unsigned Commits on an Existing Branch

If your Pull Request is blocked because previous commits are unsigned:

### Re-sign the most recent commit:
```bash
git commit --amend --no-edit -S
git push --force-with-lease
```

### Re-sign multiple past commits on your feature branch:
```bash
git rebase --exec 'git commit --amend --no-edit -S' -i origin/main
git push --force-with-lease
```

---

## 🚀 Development & Contribution Workflow

1. **Fork or branch**:
   ```bash
   git checkout -b feature/my-new-feature
   ```

2. **Develop & Test**:
   - Make sure your node is running (`bash scripts/init-lab.sh`).
   - If adding or changing examples, verify they pass locally:
     ```bash
     bash scripts/test-examples.sh
     ```

3. **Commit with signatures**:
   ```bash
   git commit -S -m "feat(examples): add Taproot multi-path spend example"
   ```

4. **Push and create a Pull Request**:
   ```bash
   git push -u origin feature/my-new-feature
   ```
   Open the PR against the `main` branch. Ensure the GitHub check confirms your commits are **Verified**.
