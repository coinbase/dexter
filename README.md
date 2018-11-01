# Dexter

[![CircleCI](https://circleci.com/gh/coinbase/dexter/tree/master.svg?style=svg)](https://circleci.com/gh/coinbase/dexter/tree/master)

Your friendly forensics expert.

Dexter is a forensics acquisition framework designed to be extensible and secure.

Dexter runs as an agent backed by S3.  Investigators use Dexter on the command line to issue investigations and retrieve reports.  Investigations define facts that must be true about the systems in scope, and tasks that will be ran on the host.  After tasks are ran, Dexter generates reports that are individually encrypted back to the investigators that are authorized to view the data.

## Architecture Overview

![](doc/dexter.png?raw=true "")

## Building

### Prerequisites

#### A working go environment

You must have go installed.  Please follow the [installation instructions](https://golang.org/doc/install) or use a alternative method such that you can successfully run `go` and have a properly setup `$GOPATH` defined in your environment.

#### dep

Install the [dep](https://github.com/golang/dep) package manager.

### Download the repository

Clone the repository into the correct place in your `$GOPATH`.

```
cd $GOPATH/src
mkdir -p github.com/coinbase
cd github.com/coinbase
git clone github.com/coinbase/dexter
cd dexter
```

### Install dependencies

Install dependencies into the vendor directory with dep.

```
dep ensure
```

### Run tests

The Makefile contains functionality to run all available tests without testing vendor.

```
make test
```

### Install

Dexter can be installed with:

```
make install
```

On linux, a bash completion script can be installed with `make bash`.

Dexter will need to be configured before it can be used.

#### Environment variables

Dexter is configured with the following environment variables.  Some are only required when Dexter is running as a daemon, others are required both when acting as a daemon as well as a command line client.

|Envar|Use|Daemon|Client|
|---|---|:---:|:---:|
|`DEXTER_AWS_S3_BUCKET`|The S3 bucket Dexter will use|✓|✓|
|`DEXTER_POLL_INTERVAL_SECONDS`|The number of seconds in between Dexter S3 polls|✓||
|`DEXTER_PROJECT_NAME_CONFIG`|Instructs Dexter on how to look up a local host's project name.  Contents must being with `file://`, followed by a local path, or `envar://`, followed by an envar name.|✓||
|`DEXTER_OSQUERY_SOCKET`|Path to the local osquery socket|✓||
|`DEXTER_AWS_ACCESS_KEY_ID`|AWS access key, used to override `AWS_ACCESS_KEY_ID`.  If not set, `AWS_ACCESS_KEY_ID` will be used instead.|✓|✓|
|`DEXTER_AWS_SECRET_ACCESS_KEY`|AWS access key, used to override `AWS_SECRET_ACCESS_KEY`.  If not set, `AWS_SECRET_ACCESS_KEY` will be used instead.|✓|✓|
|`DEXTER_AWS_REGION`|AWS access key, used to override `AWS_REGION`.  If not set, `AWS_REGION` will be used instead.|✓|✓|

#### Amazon S3 access

In order to use Dexter, you will need to have access to an S3 bucket.

Dexter usage can be divided into three roles: daemon, investigator, and admin.

##### Daemon

Dexter daemons will need to the following aws permissions to use the S3 bucket:

* `ListBucket` on the entire bucket
* `GetObject` on `investigations`
* `GetObject` on `investigations/*`
* `GetObject` on `investigators`
* `GetObject` on `investigators/*`
* `PutObject` on `reports/*`
* `PutObjectAcl` on `reports/*`

##### Investigators

Investigators will require the following permissions to use Dexter:

* `GetObject` on the entire bucket
* `ListBucket` on the entire bucket
* `PutObject` on `investigations/*`
* `PutObjectAcl` on `investigation/*`
* `PutObject` on `investigators/*`
* `PutObjectAcl` on `investigators/*`

##### Admins

Dexter admins should have all the permissions of investigators, as well as the following additional permissions:

* `DeleteObject` on the entire bucket

This makes it possible for admins to prune old investigations and reports.

## Usage

Full documentation for dexter is auto-generated [here](doc/dexter.md).

### Setting up an investigator

The command [`dexter investigator init`](doc/dexter_investigator_init.md) can be used to create a new investigator on a new system.

Create the new investigator from within the Dexter repository, and make sure the `investigators` directory exists already.  You will set a new password which will be used when investigations are signed and reports are downloaded.

```
$ ./dexter investigator init hayden
Initializing new investigator "hayden" on local system...
Set a new password >
Confirm >
hayden.json has been generated in the investigators directory,
submit this change as a pull request.
``` 

This will create a `~/.dexter` directory and place the investigator file in the investigators directory.

Now, re-create the embedded file and create a pull request

```
make embed
git checkout -b new-investigator
git add --all
git commit -m "new investigator"
```

Once this updated version of dexter is deployed as a daemon, the new investigator will be active.

### Revoking investigators

The command [`dexter investigator emergency-revoke`](doc/dexter_investigator_emergency-revoke.md) can be used to revoke a compromised investigator.

This command issues an investigation with a single task: destroy the investigator's public key in all dexter daemons.  This will make it impossible for Dexter to create reports this investigator can read, and to validate investigations with the investigator's signature.  It also removes all currently generated reports that can be read by this investigator.  The investigator will need to be permanently removed by deleting their key from the `investigators` directory in Dexter and redeploying all instances of Dexter.

### Deploying the daemon

The command [`dexter daemon`](doc/dexter_daemon.md) is used to start a daemon.

Dexter daemon can be deployed either as a binary or as a docker container.  When deployed via docker, it is important to provide Dexter with access to the docker socket and osquery socket, if you intend on using those features.  The Dockerfile included in this repo is a good place to start, but will require the configuration file to be edited before building.

### Creating an investigation

The command [`dexter investigation create`](doc/dexter_investigation_create.md) is used to create new investigations.

Running this command will enter into an interactive cli where an investigation can be configured, signed, and uploaded.

### Listing investigations

The command [`dexter investigation list`](doc/dexter_investigation_list.md) is used to list all investigations stored in the Dexter bucket.

```
$ dexter investigation list
+---------------+--------+-------------------------+------------------------+-----------+-------------+
| INVESTIGATION | ISSUER |          TASKS          |         SCOPE          | CONSENSUS | REVIEWED BY |
+---------------+--------+-------------------------+------------------------+-----------+-------------+
| 1e8b73bb      | bob    | docker-filesystem-diff, | platform-is("linux"),  | 1/1       | alice       |
|               |        | osquery-collect         | user-exists(REDACTED)  |           |             |
+---------------+--------+-------------------------+------------------------+-----------+-------------+
```

### Approving investigations

The command [`dexter investigation approve`](doc/dexter_investigation_approve.md) is used to preview and sign investigations that require consensus approval.

```
$ dexter investigation approve 1
Provide your password to approve the following investigation:
+------------------+--------------------------------+
|      FIELD       |             VALUE              |
+------------------+--------------------------------+
| ID               | 1e8b73bb                       |
| Issued By        | bob                            |
| Tasks            | osquery-collect,               |
|                  | docker-filesystem-diff         |
| Scope            | platform-is("linux"),          |
|                  | user-exists(REDACTED)          |
| Kill Containers? | false                          |
| Kill Host?       | false                          |
| Recipients       | alice, bob                     |
| Approvers        |                                |
+------------------+--------------------------------+
Password >
```

### Pruning investigations

The command [`dexter investigation prune`](doc/dexter_investigation_prune.md) is used to delete old investigations.

When this command is ran, all past investigations will be downloaded into a new directory called `InvestigationArchive`.

### Listing reports

The command [`dexter report list`](doc/dexter_report_list.md) is used to print a table of reports.

```
$ dexter report list
+---------------+--------+-------------------------+-----------------------+------------+----------------+
| INVESTIGATION | ISSUER |          TASKS          |         SCOPE         | RECIPIENTS | HOSTS UPLOADED |
+---------------+--------+-------------------------+-----------------------+------------+----------------+
| 1e8b73bb      | bob    | docker-filesystem-diff, | platform-is("linux"), | alice,     | 1              |
|               |        | osquery-collect         | user-exists(REDACTED) | bob        |                |
+---------------+--------+-------------------------+-----------------------+------------+----------------+
```

### Downloading reports

The command [`dexter report retrieve`](doc/dexter_report_retrieve.md) is used to download reports.

The encrypted report will be downloaded, and you will be prompted for your password.  Once provided, the report will be populated in a new directory.

The report format is:

```
DexterReport-<ID>/<hostname>/<taskname>/...
```

### Pruning reports

The command [`dexter report prune`](doc/dexter_report_prune.md) is used to delete old reports.

No reports will be saved, this should be used with caution.

## Development

### Adding facts

New facts can be added very easily.  Make a copy of the [example fact](facts/example.go) and replace the contents as needed with your new fact.  Rebuild and re-deploy dexter, and your fact will be available for use.

### Adding tasks

New tasks can be added just like new facts.  Make a copy of the [example task](tasks/example.go), replacing the content as needed, and redeploy.
