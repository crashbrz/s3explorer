![License](https://img.shields.io/badge/license-sushiware-red)
![Issues open](https://img.shields.io/github/issues/crashbrz/s3explorer)
![GitHub pull requests](https://img.shields.io/github/issues-pr-raw/crashbrz/s3explorer)
![GitHub closed issues](https://img.shields.io/github/issues-closed-raw/crashbrz/s3explorer)
![GitHub last commit](https://img.shields.io/github/last-commit/crashbrz/s3explorer)

# S3Explorer

S3Explorer is a command-line tool written in Go that allows you to list and download objects from S3-compatible bucket URLs. It supports filtering, multi-threaded downloads, and handling multiple bucket URLs provided in a file.

## Features

- Retrieve and list objects from an S3 bucket.
- Filter objects by substring match.
- Download individual or all objects concurrently with configurable thread limits.
- Support for multiple bucket URLs via file input.
- Debug mode for detailed error messages.

## Installation

1. Clone the repository:

   ```bash
   git clone <repository-url>
   cd <repository-name>
   ```

2. Build the binary:

   ```bash
   go build -o s3explorer
   ```

## Usage

### Command-line Flags

| Flag     | Description                                   | Example                              |
| -------- | --------------------------------------------- | ------------------------------------ |
| `-u`     | S3 bucket URL to retrieve keys from           | `-u https://bucket.s3.amazonaws.com` |
| `-U`     | File containing a list of S3 bucket URLs      | `-U buckets.txt`                     |
| `-t`     | Number of goroutines for concurrent downloads | `-t 30`                              |
| `-l`     | Limit the number of keys to retrieve          | `-l 50`                              |
| `-d`     | Download a single key                         | `-d example/key.txt`                 |
| `-D`     | Download all keys found                       | `-D`                                 |
| `-f`     | Filter keys by substring match                | `-f log`                             |
| `-debug` | Enable debug mode for detailed error messages | `-debug`                             |

### Examples

#### List First 10 keys from an S3 Bucket

```bash
./s3explorer -u https://bucket.s3.amazonaws.com -l 10
```

#### Filter Keys Containing a Specific Substring

```bash
./s3explorer -u https://bucket.s3.amazonaws.com -f "passwd"
```

#### Download a Single Key

```bash
./s3explorer -u https://bucket.s3.amazonaws.com -d example/key.txt
```

#### Download All Keys Concurrently

```bash
./s3explorer -u https://bucket.s3.amazonaws.com -D -t 50
```

#### Use a File with Multiple Bucket URLs

```bash
./s3explorer -U buckets.txt -l 20
```

### Debug Mode

Enable debug mode for troubleshooting:

```bash
./s3explorer -u https://bucket.s3.amazonaws.com -debug
```

## License

S3Explorer is licensed under the SushiWare license. For more information, check [docs/license.txt](docs/license.txt).

## Acknowledgments

- [pb](https://github.com/cheggaaa/pb) for progress bar integration.
