[![Codacy Badge](https://app.codacy.com/project/badge/Grade/54e6788204d54ffeb627e2da1958c9cc)](https://app.codacy.com/gh/B87/file-bridge/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade) [![Codacy Badge](https://app.codacy.com/project/badge/Coverage/54e6788204d54ffeb627e2da1958c9cc)](https://app.codacy.com/gh/B87/file-bridge/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_coverage) ![build main](https://github.com/B87/file-bridge/actions/workflows/go.yml/badge.svg?branch=main)
# File Bridge: Go Multi-tool for File Manipulation across File Systems

File Bridge (fileb) is a versatile tool developed in Go (Golang) designed for efficient file and directory manipulation. It provides both a Command-Line Interface (CLI) and a library. 
This tool is capable of handling basic operations across various file systems, streamlining the process of managing files seamlessly.

## Key Features

- **Multi-File System Support:** File Bridge is built to work effortlessly across multiple file systems simultaneously, ensuring a flexible and integrated experience.
- **Dual Interface:** Offers both a CLI for direct command execution and a library for other Go projects integration.

## Supported File Systems

- **Local File System:** Directly manage files on your local machine.
- **Google Cloud Platform (GCP):** Use Google Storage (GS) buckets as file systems.


## Installation

Download binary or build from source, test the installation with `fileb` or `go run .`

Set GOOGLE_APPLICATION_CREDENTALS env var if you plan to use GS filesystems.

## [CLI](https://github.com/B87/file-bridge/wiki/CLI)

The CLI allows to easily manage files from multiple file systems or storages from the terminal.

`fileb cp -r ~/folder gs://mybucket`

See also `fileb -h`

## [Packages (pkg)](https://github.com/B87/file-bridge/wiki/Packages)

This is a collection of code used in the application meant to be reused in other go projects.

### filesys

Manage files and file systems.

### image

Adapted version of [imaging](https://github.com/disintegration/imaging) to manipulate images.
