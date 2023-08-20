# Contributions

Provide a generic tasks for new and existing contributors

## Types of changes

There are many ways to contribute to the ksctl project. Here are a few examples:

* **New changes to docs:** You can contribute by writing new documentation, fixing typos, or improving the clarity of existing documentation.
* **New features:** You can contribute by proposing new features, implementing new features, or fixing bugs.
* **Cloud support:** You can contribute by adding support for new cloud providers.
* **Kubernetes distribution support:** You can contribute by adding support for new Kubernetes distributions.

Phases a change / feature goes through

1. Raise a issue regarding it (used for prioritizing)
2. what all changes does it demands
3. if all goes well you will be assigned
4. If its about adding Cloud Support then usages of CloudFactory is needed and sperate the logic of vm, firewall, etc. to their respective files and do have a helper file for behind the scenes logic for ease of use
5. If its about adding Distribution support do check its compatability with different cloud providers vm configs and firewall rules which needs to be done

## Formating for PR & Issue subject line

### Subject / Title

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

### Body
Follow the PR or Issue template
add all the significant changes to the PR description

### Commit messages
mention the detailed description in the git commits.
what? why? How?

**each commits must be sign-off**

# Development
First you have to fork the ksctl repository. [fork](https://github.com/kubesimplify/ksctl/fork)
```bash
cd <path> # to you directory where you want to clone ksctl
mkdir <directory name> # create a directory
cd <directory name> # go inside the directory
git clone git@github.com:${YOUR_GITHUB_USERNAME}/ksctl.git # clone you fork repository
cd ksctl # go inside the ksctl directory
git remote add upstream https://github.com/kubesimplify/ksctl.git # set upstream
git remote set-url --push upstream no_push # no push to upstream
```

# Making Build
### Linux
```bash
make install_linux # for linux
```
### Mac OS
```bash
make install_macos # for macos
```
### Windows
```bash
.\builder.ps1 # for windows
```
## for website
```bash
cd website # to go inside the directory
```
### Install Dependencies
```bash
npm install # install npm  
npm install --global yarn # install yarn
```

### Start the Server
```bash
yarn start # to run a development server
```
## Trying out code changes

Before submitting a code change, it is important to test your changes thoroughly. You can do this by running the unit tests and integration tests.

### Unit tests
```bash
make test
```

## Submitting changes

Once you have tested your changes, you can submit them to the ksctl project by creating a pull request.
Make sure you use the provided PR template

## Getting help

If you need help contributing to the ksctl project, you can ask for help on the kubesimplify Discord server, ksctl-cli channel or else raise issue or discussion

## Thank you for contributing!

We appreciate your contributions to the ksctl project!

Some of our contributors [ksctl contributors](https://github.com/kubesimplify/ksctl/graphs/contributors)
