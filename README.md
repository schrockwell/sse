# SSE - Stupidly Simple Environments

SSE is a highly-opinionated, platform-agnostic tool for managing secret environment variables.

It's a single-file executable with no external dependencies, so it's easy to integrate with development, build, and deployment environments.

Secrets are saved in an encrypted TOML file that is safe to check into version control. Only the `master.key` file is required to decrypt and edit it. Multiple environments are supported; `development` is the default.

SSE draws inspiration from [Rails credentials](https://thoughtbot.com/blog/switching-from-env-files-to-rails-credentials) and [SOPS](https://github.com/getsops/sops). It uses [age](https://github.com/FiloSottile/age) for encryption.

## Available Commands

Run `sse help [command]` for details.

```
$ sse help
Stupidly Simple Environments (sse) manages encrypted environment variables
for small projects using age encryption.

Files:
  master.key  - age keypair (add to .gitignore)
  env.toml    - environment file with encrypted values (safe to commit)

The env.toml file contains sections for each environment:
  [development]
  API_KEY = "ENC[...]"

  [production]
  API_KEY = "ENC[...]"

Keys are human-readable, only values are encrypted.

Usage:
  sse [command]

Available Commands:
  completion  Generate the autocompletion script for the specified shell
  edit        Edit env.toml
  help        Help about any command
  init        Initialize a new project
  load        Export variables to current shell
  show        Print decrypted env.toml
  with        Run a command with decrypted environment

Flags:
  -h, --help   help for sse

Use "sse [command] --help" for more information about a command.
```