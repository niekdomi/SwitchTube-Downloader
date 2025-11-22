# SwitchTube-Downloader: A Streamlined CLI for SwitchTube Video Downloads

**SwitchTube-Downloader** is a lightweight, efficient command-line tool designed
to easily download videos from [SwitchTube](https://tube.switch.ch/).

## Getting Started

1. **Download the binary**: Visit the [releases page](https://github.com/niekdomi/SwitchTube-Downloader/releases)
   to obtain the appropriate binary for your operating system (Linux, MacOS,
   Windows).

   Arch-Linux users can also use the AUR package:

   ```bash
   yay -S switchtube-downloader-bin
   ```

   **NOTE**: The AUR executable is named `swdl` for convenience.

2. **Make executable**: After downloading, ensure the binary is executable. For
   Linux and MacOS, run:

   ```bash
   chmod +x switchtube-downloader
   ```

3. **Usage**: Run `./switchtube-downloader` to access the help menu,
   which provides clear guidance on available commands.

4. **Create access token**: A SwitchTube access token is required. Generate
   one [here](https://tube.switch.ch/access_tokens) to authenticate your
   requests.

<details>
  <summary>[Click me] for detailed usage instructions</summary>

Running the SwitchTube Downloader without arguments displays available commands:

<pre><code>
./switchtube-downloader
A CLI downloader for SwitchTube videos

Usage:
  SwitchTube-Downloader [command]

Available Commands:
  download    Download a video or channel
  help        Help about any command
  token       Manage the SwitchTube access token
  version     Print the version number of the SwitchTube downloader

Flags:
  -h, --help   help for SwitchTube-Downloader

Use "SwitchTube-Downloader [command] --help" for more information about a command.
</code></pre>

## Downloading a video or a channel

To download a video or channel, use the `download` command with either the
video/channel ID or its full URL:

<pre><code>./switchtube-downloader download {id or url}</code></pre>

For example, for the URL `https://tube.switch.ch/channels/dh0sX6Fj1I`, the ID is
`dh0sX6Fj1I`. You can use either:

- **URL**: More convenient, directly copied from the browser:
  <pre><code>./switchtube-downloader download https://tube.switch.ch/channels/dh0sX6Fj1I</code></pre>

- **ID**: Shorter, but requires extracting the ID:
  <pre><code>./switchtube-downloader download dh0sX6Fj1I</code></pre>

To view detailed help for the `download` command:

<pre><code>
./switchtube-downloader download --help
Download a video or channel. Automatically detects if input is a video or channel.
You can also pass the whole URL instead of the ID for convenience.

Usage:
SwitchTube-Downloader download <id|url> [flags]

Flags:
  -a, --all             Download the whole content of a channel
  -e, --episode         Prefixes the video with episode-number e.g. 01_OR_Mapping.mp4
  -f, --force           Force overwrite if file already exist
  -h, --help            help for download
  -o, --output string   Output directory for downloaded files
  -s, --skip            Skip video if it already exists
</code></pre>

### Using Flags

You can add optional flags to customize the download. For example:

- Single flag:
  <pre><code>./switchtube-downloader download dh0sX6Fj1I -f</code></pre>

- Multiple flags combined:
  <pre><code>./switchtube-downloader download dh0sX6Fj1I -a -f -e</code></pre>

### Available Flags

- `-a`, `--all`: Download all videos from a channel. This means that if you
  provide a channel ID, it will download all videos in that channel. You can
  also add this flag to a video ID, but with no effect.

- `-e`, `--episode`: Prefixes the video filename with the episode number, e.g.,
  `01_OR_Mapping.mp4`. This is useful for channels with multiple videos. So you
  keep track of the order of the videos.

  Keep in mind that the prefix might look like `04ar`. This is **not** a bug,
  but the name set by the video uploader.

- `-f`, `--force`: Forces the download to overwrite existing files. Use this
  flag with caution, as it will replace any existing files without confirmation.
  Force has also precedence over the `--skip` flag, meaning that if you use both
  flags, the file will be overwritten.

- `-h`, `--help`: Displays help information for the `download` command. Running
  a command without a flag, e.g. `./switchtube-downloader download` will
  automatically trigger the help menu.

- `-o`, `--output`: Specifies the output directory for downloaded files. Per
  default the current working directory is used (cwd). If you want to change the
  output directory you can pass the path like this:
  - Absolute path: `./switchtube-downloader download dh0sX6Fj1I -o /path/to/dir`
  - Relative path:
    - Current dir:
      - `./switchtube-downloader download dh0sX6Fj1I -o path/to/dir`
      - `./switchtube-downloader download dh0sX6Fj1I -o ./path/to/dir`
    - Parent dir: `./switchtube-downloader download dh0sX6Fj1I -o ../path/to/dir`

- `-s`, `--skip`: Skips the download if the video already exists in the output
  directory. This is useful to avoid re-downloading videos.

## Managing access token

The `token` command manages the SwitchTube access token stored in the system
keyring:

<pre><code>
./switchtube-downloader token
Manage the SwitchTube access token stored in the system keyring

Usage:
  SwitchTube-Downloader token [flags]
  SwitchTube-Downloader token [command]

Available Commands:
  delete      Delete access token from the keyring
  get         Get the current access token
  set         Set a new access token

Flags:
  -h, --help   help for token

Use "SwitchTube-Downloader token [command] --help" for more information about a command.
</code></pre>

**Note**: The `delete` subcommand removes the token without a confirmation
prompt, so use it carefully.

</details>

## Why to choose (this) SwitchTube-Downloader?

| Feature                        | [SwitchTube-Downloader](https://github.com/niekdomi/SwitchTube-Downloader) |
| ------------------------------ | -------------------------------------------------------------------------- |
| **Binary Size**                | 7.2MB (light)                                                              |
| **Store Access Token**         | Automatic storage                                                          |
| **Encrypted Access Token**     | Secure encryption                                                          |
| **Intuitive Downloads**        | One simple command                                                         |
| **Video download**             | Supported                                                                  |
| **Channel download**           | Supported                                                                  |
| **Select videos of a channel** | Supported                                                                  |
| **Support ID and URL**         | Supported                                                                  |

## FAQ

> Can we select the video quality?

Multiple video quality options are usually available, but to keep the downloader
simple, I chose not to include a quality selection flag, since most users will
use the highest quality available anyway.

> Can we download multiple videos at once?

Currently, the downloader doesn't support batch downloads (e.g.
`./switchtube-downloader download dh5sX1Fj3I qu0fK6Sw1V dh0sX6Fj1I`). If
there is enough interest, I might implement this feature in the future.

> Is it possible to configure default settings such as output directory?

Currently, the downloader does not support this. If there is enough interest,
I might implement this feature in the future.

## Testing the SwitchTube API

For developers or curious users, you can interact directly with the SwitchTube
API using the following command:

```bash
curl -H "Authorization: Token {your_token}" \
        https://tube.switch.ch/api/v1/xxx
```

E.g., you can write the output to a file to examine the JSON structure:

```bash
curl -H "Authorization: Token {your_token}" \
        https://tube.switch.ch/api/v1/browse/channels/{channel_id}/videos | tee tmp.json
```
