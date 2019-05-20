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

Dexter uses Go modules, so you must `export GO111MODULE=on` in your environment.

### Download the repository

Clone the repository into the correct place in your `$GOPATH`.

```
cd $GOPATH/src
mkdir -p github.com/coinbase
cd github.com/coinbase
git clone github.com/coinbase/dexter
cd dexter
```

### Run tests

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

* `ListBucket` on `investigations`
* `ListBucket` on `investigations/*`
* `ListBucket` on `investigators`
* `ListBucket` on `investigators/*`
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

##### Admins

Dexter admins should have all the permissions of investigators, as well as the following additional permissions:

* `PutObject` on `investigators/*`
* `PutObjectAcl` on `investigators/*`
* `DeleteObject` on the entire bucket

This makes it possible for admins to prune old investigations and reports.

## Usage

Full documentation for dexter is auto-generated [here](doc/dexter.md).

### Setting up an investigator

The command [`dexter investigator init`](doc/dexter_investigator_init.md) can be used to create a new investigator on a new system.  You will set a new password which will be used when investigations are signed and reports are downloaded.

```
$ ./dexter investigator init hayden
Initializing new investigator "hayden" on local system...
Set a new password >
Confirm >
New investigator file created: hayden.json
This must be uploaded to Dexter by your Dexter administrator.
``` 

A dexter admin can now place this file in the investigators directory of the S3 bucket.

This will create a `~/.dexter` directory locally containing your encrypted private key.

### Revoking investigators

The command [`dexter investigator emergency-revoke`](doc/dexter_investigator_emergency-revoke.md) can be used to revoke an investigator.

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
