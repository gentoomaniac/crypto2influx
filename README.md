# Template go Repository

## tl;dr

This is a template go repository with actions already set up to create compiled releases

## What does this Template provide?

* a basic cli application with subcommands based on [Kong](https://github.com/alecthomas/kong)
* logging using zerolog
* some customisations like a more descriptive version info
* GitHub workflow to run tests on every push
* GitHub Workflow to build binary releases with [goreleaser](https://github.com/goreleaser/goreleaser)

## What is missing?

* A sample for a spinner using [spinner](https://github.com/briandowns/spinner)
* some sample code for a cli interactive selection dialogue using [promptui](https://github.com/manifoldco/promptui)

## How to use this template

### Fetch the project

```bash
git clone https://github.com/gentoomaniac/go-template.git ./
rm -r .git
```

### update all references to the template

```
# goreleaser IDs and binary names
sed 's/template-application/my-new-application/g' .goreleaser.yaml

# go.mod
sed 's#gentoomaniac/go-template#githubuser/reponame#g' go.mod
```

### check in the code

```
git init
git add -A
git commit -m 'import template'
```

## How to build locally

```
goreleaser build --single-target --snapshot
```

## Example runs

### help

```
> template-application --help
Usage: template-application <command>

Flags:
  -h, --help             Show context-sensitive help.
  -v, --verbosity=INT    Increase verbosity.
  -q, --quiet            Do not run upgrades.
      --json             Log as json
      --debug            shortcut for -vvvv

Commands:
  foo
    FooBar command

Run "template-application <command> --help" for more information on a command.
```

### logging

```
> template-application -vvv
8:40PM INF Default command
```

```
> template-application -vvv --json foo
{"level":"info","time":"2021-11-05T20:41:33+01:00","message":"foo command"}
```

### version

```
template-application -V
template-application commit:bf9d771 release:snapshot build:workflow/1 date:2021-11-05T19:58:00Z goVersion:go1.17.2 platform:linux/amd64
```
