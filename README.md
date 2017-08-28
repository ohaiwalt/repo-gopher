Repo Gopher
============

[![Build Status](https://travis-ci.org/ohaiwalt/repo-gopher.svg?branch=master)](https://travis-ci.org/ohaiwalt/repo-gopher)

A utility for ensuring that a GitHub repository has the correct set of labels. The owner can create a toml file to configure a list of labels, and a list of mappings to rename older labels correctly.

This tool is shamelessly ~ripped off~ ported from https://github.com/thommay/repo_man so I didn't have to mess around with running Ruby locally. All credit to the author.

Configuration
--------------

Given the following configuration file:

```toml
repositories = [ "example/fox", "example/wolf" ]

[[label]]
name = "bug"
color = "f29513"
mappings = [ "defect", "error" ]

[[label]]
name = "Jump In"
color = "123456"

[[label]]
name = "An Old Label"
delete = true
color = "123456"

```

repo-gopher would create two labels, `bug` and `Jump In`, and would ensure
any existing issues labelled as `defect` or `error` were relabelled as
`bug`.

The config file can either be specified from the command line, or will use `/etc/repo-gopher/config.toml` by default.

Syntax
------

The config file consists of two arrays, `repositories`, and `label`.

`repositories` is an array of `organization`/`repository` names.

A label may have the following keys: 

* `name` string, required
* `color` string, required
* `delete` bool

There is a special key, `mappings`, that accepts an array of existing labels that should be renamed to the current one. Renaming is done by applying the new label and then removing the old one, so it should be idempotent in the face of failed runs.

Running
--------

Repo Gopher expects you to have the environment variable `GITHUB_AUTH_TOKEN` set. To get a GitHub API token, go [here](https://github.com/settings/tokens).


To apply a config file to a repository, run
```
repo-gopher -c config.toml
```

To run using the provided Dockerfile, run
```
docker run -v /local/path/to/config.toml:/etc/repo-gopher/config.toml -e GITHUB_AUTH_TOKEN=$GITHUB_AUTH_TOKEN ohaiwalt/repo-gopher
```
