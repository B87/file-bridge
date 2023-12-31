[![Codacy Badge](https://app.codacy.com/project/badge/Grade/54e6788204d54ffeb627e2da1958c9cc)](https://app.codacy.com/gh/B87/file-bridge/dashboard?utm_source=gh&utm_medium=referral&utm_content=&utm_campaign=Badge_grade) 
# File Bridge: Go Multi-tool for File Manipulation across File Systems

File Bridge (fileb) is a versatile tool developed in Go (Golang) designed for efficient file and directory manipulation. It provides both a Command-Line Interface (CLI) and a library. 
This tool is capable of handling basic operations across various file systems, streamlining the process of managing files seamlessly.

## Key Features

- **Multi-File System Support:** File Bridge is built to work effortlessly across multiple file systems simultaneously, ensuring a flexible and integrated experience.
- **Dual Interface:** Offers both a CLI for direct command execution and a library for other Go projects integration.

## Supported File Systems

- **Local File System:** Directly manage files on your local machine.
- **Google Cloud Platform (GCP):** Use Google Storage buckets as file systems.

## Installation

### 1. Setup env variables

| Env Var                        | Default | Description                               |
| ------------------------------ | ------- | ----------------------------------------- |
| GOOGLE_APPLICATION_CREDENTIALS | None    | (Optional) Used by GGP Storage FileSystem |


### 2. Run bash install script ``sudo bash scripts/install.sh``

This script will build from source and add the binary to /usr/local/bin 

Assumptions:
- Go is installed and available in your path
- /usr/local/bin exists and is inside $PATH

## CLI 

### List

`fileb help ls`

#### List Examples
| Command                       | Description                                      |
| ----------------------------- | ------------------------------------------------ |
| `fileb ls ~/`                 | Lists local filesystem user home                 |
| `fileb ls gs://bucket/folder` | Lists "folder" of Google Storage bucket "bucket" |


### Copy

`fileb help cp`

#### Copy Examples

| Command                                     | Description                                      |
| ------------------------------------------- | ------------------------------------------------ |
| `fileb cp ~/example.txt ~/folder`           | Copy local file example.txt to local folder      |
| `fileb cp ~/example.txt gs://bucket/folder` | Lists "folder" of Google Storage bucket "bucket" |

### Move
`fileb help mv`

| Command                                     | Description                                        |
| ------------------------------------------- | -------------------------------------------------- |
| `fileb cp ~/example.txt ~/folder`           | Copy local file example.txt to local folder        |
| `fileb cp ~/example.txt gs://bucket/folder` | Lists "folder" from Google Storage bucket "bucket" |

### Remove
`fileb help rm`

| Command                          | Description                                         |
| -------------------------------- | --------------------------------------------------- |
| `fileb rm ~/example.txt`         | Remove local file example.txt                       |
| `fileb rm -r gs://bucket/folder` | Remove "folder" from Google Storage bucket "bucket" |

### Make Dir
| Command                          | Description                                |
| -------------------------------- | ------------------------------------------ |
| `fileb mkdir ~/folder`           | Creates local empty folder                 |
| `fileb mkdir gs://bucket/folder` | Creates empty folder Google Storage bucket |


## Packages (pkg)

### Filesys

Manage files and file systems

### Image

Adapted version of [imaging](https://github.com/disintegration/imaging) to manipulate images