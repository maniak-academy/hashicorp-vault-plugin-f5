# Git Instructions for HashiCorp Vault F5 BIG-IP Token Plugin

This document provides instructions for managing the plugin's source code using Git.

## Initial Setup (Already Completed)

The repository has been initialized with Git and all files have been committed. The following steps have been completed:

```bash
git init
git add .
git commit -m "Initial commit of HashiCorp Vault F5 BIG-IP Token Plugin"
```

## Pushing to GitHub or GitLab

### GitHub

1. Create a new repository on GitHub (without README, license, or gitignore)
2. Push your local repository:

```bash
# Add the remote repository
git remote add origin https://github.com/YOUR_USERNAME/vault-plugin-f5-token.git

# Push the code
git push -u origin main
```

### GitLab

1. Create a new project on GitLab (without README, license, or gitignore)
2. Push your local repository:

```bash
# Add the remote repository
git remote add origin https://gitlab.com/YOUR_USERNAME/vault-plugin-f5-token.git

# Push the code
git push -u origin main
```

## Daily Git Workflow

### Making Changes

1. Make your changes to the code
2. Check status to see what files have changed:
   ```bash
   git status
   ```
3. Add changed files to staging:
   ```bash
   git add <filename>
   # Or add all changes
   git add .
   ```
4. Commit your changes:
   ```bash
   git commit -m "Descriptive message about your changes"
   ```
5. Push to the remote repository:
   ```bash
   git push
   ```

### Creating Branches

For significant changes, create a feature branch:

```bash
# Create and checkout a new branch
git checkout -b feature/new-feature-name

# Work on your changes...

# Push branch to remote
git push -u origin feature/new-feature-name
```

### Best Practices

1. **Commit Messages**:
   - Use clear, descriptive commit messages
   - Begin with a verb (e.g., "Add", "Fix", "Update")
   - Keep the first line under 50 characters
   - Use the extended description for more details

2. **Branching Strategy**:
   - `main` or `master`: Production-ready code
   - `develop`: Integration branch for features
   - `feature/...`: New features
   - `bugfix/...`: Bug fixes
   - `release/...`: Release preparation

3. **Git Ignore**:
   - The `.gitignore` file is configured to exclude build artifacts, temporary files, and sensitive information
   - Update it as needed

## Git Resources

- [GitHub Documentation](https://docs.github.com/en)
- [GitLab Documentation](https://docs.gitlab.com/)
- [Git Book](https://git-scm.com/book/en/v2)
- [Git Cheat Sheet](https://education.github.com/git-cheat-sheet-education.pdf) 