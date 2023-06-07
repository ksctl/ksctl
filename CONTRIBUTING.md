# PRs
- must follow PR template
- need to sign-off each commits
- add all the significant changes to the PR description
- if it's documentation changes set target to `main` branch

# Issues
- mention the domain and concise subject

# Documentation
- mention the function comments
- use `go fmt` command to format code

# Formating for PR & Issue subject line

## Subject / Title

```markdown
# Releated to enhancement
enhancement: <Title>

# Related to feature
feat: <Title>

# Related to Bug fix or other types of fixes
fix: <Title>

# Related to update
update: <Title>
```

## Tash Description
must be elloborate on what is doing to be added/removed/modified and why
and in solutions part how? it is changed

# Git commits

mention the detailed description in the git commits
what? why? How?

# Workflow when working on source code changes
## Command Line

> There are 2 approaches

### test the working directly on the api/ directory

> **Note**
create some kind of handler function which can be used to test specific provider or feature
and once its done **remove** it

### test by building the entire CLI
```bash
make install_<linux,macos_intel,macos>
# choose based on your platform and install as dev version
```

## Website
make sure you test your changes before creating the PR / Issue
```bash
cd website
npm install
yarn start # to run a development server
```