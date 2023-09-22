# Delly: File Cleanup Tool

Delly is a command-line utility designed to help you efficiently clean up your file system by recursively deleting files with specific file extensions in a given directory. This tool is especially handy when you need to reclaim storage space by removing files matching certain criteria.

## Table of Contents

- [Installation](#installation)
- [Usage](#usage)
- [Example](#example)
- [License](#license)

## Installation

To use Delly, you'll need to have it installed on your system. You can do this by following these steps:

1 Install the binary into your GOBIN with go install:

   ```shell
   go install github.com/bxffour/delly@latest
   ```

Now, Delly is ready to use on your system.

## Usage

Delly is simple to use and takes two main parameters: file extension(s) to match and a directory to start the search from. Here's the basic usage:

```shell
delly -e <extensions> <directory>
```

- `-e <extensions>`: Specify the file extensions to match, separated by commas (e.g., "mp4,zip").
- `<directory>`: Provide the directory where Delly should begin its search for matching files.

Delly will then provide a list of matching files along with their sizes and ask for your confirmation before deleting them. Additionally, it reports the bytes saved per directory after the deletion process.

## Example

Let's walk through a typical usage scenario. Suppose you want to delete all `.mp4`, `ttf` and `.zip` files from your `~/Downloads` directory:

```shell
delly -e mp4,zip,ttf ~/Downloads
```

Delly will display a list of matching files, their sizes, and ask for your confirmation before proceeding with the deletion. After successfully deleting the files, it will report the bytes saved per directory.

Output:

```shell
FILE                                                                                  SIZE
----                                                                                  ----
/home/user/Downloads/go1.21.1.linux-amd64.tar.gz                                         67 MB
/home/user/Downloads/Cantarell.zip                                                       185 kB
/home/user/Downloads/Inter.zip                                                           3.6 MB
/home/user/Downloads/JetBrainsMono-2.304/fonts/variable/JetBrainsMono-Italic[wght].ttf   309 kB
/home/user/Downloads/JetBrainsMono-2.304/fonts/variable/JetBrainsMono[wght].ttf          303 kB
/home/user/Downloads/JetBrainsMono-2.304.zip                                             5.6 MB
/home/user/Downloads/bundleservice.zip                                                   12 MB
----                                                                                  ----
TOTAL                                                                                 89 MB

do you want to go ahead with deleting these files? [y/n]: y

DIRECTORY                                              OLDSIZE     NEWSIZE     BYTES SAVED
---------                                              -------     -------     -----------
/home/user/Downloads                                      332 MB      244 MB      88 MB
/home/user/Downloads/JetBrainsMono-2.304/fonts/variable   612 kB      0 B         612 kB
```

## Contributing

We welcome contributions to Delly! If you find a bug, have an idea for an improvement, or want to add a new feature, please open an issue or create a pull request on the [Delly GitHub repository](https://github.com/bxffour/delly).

## License

Delly is open-source software licensed under the [MIT License](LICENSE). You are free to use, modify, and distribute this tool according to the terms of the license.
