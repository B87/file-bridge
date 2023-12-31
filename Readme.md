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

### 1. Setup env variables

|            Env Var             | Default |                Description                |
| :----------------------------: | :-----: | :---------------------------------------: |
| GOOGLE_APPLICATION_CREDENTIALS |  None   | (Optional) Used by GGP Storage FileSystem |


### 2. (Optional) Run bash install script 

`sudo bash scripts/install.sh`

This script will build from source and add the binary to /usr/local/bin 

Assumptions:
- Go is installed and available in your path
- /usr/local/bin exists and is inside $PATH
  
### 3. Run

`fileb` or `go run .`

## CLI

### List

`fileb help ls`

|            Example            |           Description            |
| :---------------------------: | :------------------------------: |
|         `fileb ls ~/`         | Lists local filesystem user home |
| `fileb ls gs://bucket/folder` |    Lists folder of GS bucket     |


### Copy
`fileb help cp`


|                   Example                   |                 Description                 |
| :-----------------------------------------: | :-----------------------------------------: |
|      `fileb cp ~/example.txt ~/folder`      | Copy local file example.txt to local folder |
| `fileb cp ~/example.txt gs://bucket/folder` |    Lists "folder" of GS bucket "bucket"     |

### Move
`fileb help mv`

|                   Example                   |                 Description                 |
| :-----------------------------------------: | :-----------------------------------------: |
|      `fileb cp ~/example.txt ~/folder`      | Copy local file example.txt to local folder |
| `fileb cp ~/example.txt gs://bucket/folder` |   Lists "folder" from GS bucket "bucket"    |

### Remove
`fileb help rm`

|             Example              |               Description               |
| :------------------------------: | :-------------------------------------: |
|     `fileb rm ~/example.txt`     |      Remove local file example.txt      |
| `fileb rm -r gs://bucket/folder` | Remove "folder" from GS bucket "bucket" |

### Make Dir
`fileb help mkdir`
|             Example              |          Description           |
| :------------------------------: | :----------------------------: |
|      `fileb mkdir ~/folder`      |   Creates local empty folder   |
| `fileb mkdir gs://bucket/folder` | Creates empty folder GS bucket |

## Packages (pkg)

### filesys

Manage files and file systems

### image

Adapted version of [imaging](https://github.com/disintegration/imaging) to manipulate images