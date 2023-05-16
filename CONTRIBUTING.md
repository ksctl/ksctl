# PRs
- if it's documentation changes set target to `main` branch
- if it's a API change then make sure you create the Issue then create 2 PR's one for API which will get merged then another for CLI
> Reason 🧠: First we need to update the API as its merged in main branch then only it can use by go module to fetch latest API which can be used for CLI

# Issues
- mention the domain and concise subject

# Documentation
- mention the function comments
- use `go fmt` command to format code

# Formating for PR & Issue subject line
```markdown
# Releated to enhancement
enhancement(<Core Component>): <Title>

# Related to feature
feat(<Core Component>): <Title>

# Related to Bug fix or other types of fixes
fix(<Core Component>): <Title>

## Core Component
- Docs
- API
- CLI
- CI

```
