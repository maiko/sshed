# SSHed - ssh connections manager and config editor
A visual cross-platform editor to manage list of SSH hosts in ssh config file and execute commands on those hosts. SSHed uses native ``ssh_config`` format to store connections information and supports all available ssh options.

# Acknowledgments

Special thanks to Eugene Terentev @trntv ([eugene@terentev.net](mailto:eugene@terentev.net)) for creating the original project that this work is based upon. His contributions and vision laid the foundation for the ongoing development of this tool.

# Disclaimer

This project is a personal learning initiative, my first experience with Go and I'm no profesionnal developper. As such, the code may not follow all best practices. While I've prioritized security and reliability, I cannot assure the tool is free from vulnerabilities. Users should take appropriate measures to secure their systems when using SSHed, such as backing up configurations and using encrypted keychains for sensitive data. I accept no liability for any damages resulting from the use of this software.

---

# Installation
download binary [here](https://github.com/maiko/sshed/releases)

or run in console
```
curl -sf https://gobinaries.com/maiko/sshed | sh
```

or install with ``go get``
```
go get -u github.com/maiko/sshed
```

# Features
- add, show, list, remove ssh hosts in ssh_config file
- show, edit ssh config via preferred text editor
- connect to host by key
- transfer files between your computer and host using key
- execute commands via ssh (on single or multiple hosts)
- use a Jumphost (also known as Bastion or ProxyHost)
- encrypted keychain to store ssh passwords and private keys

# Usage
```
NAME:
   sshed - SSH config editor and hosts manager

USAGE:
   help [global options] command [command options] [arguments...]

AUTHORS:
   Eugene Terentev <eugene@terentev.net>
   Maiko BOSSUYT <hello@maiko-bossuyt.eu>

COMMANDS:
   show      Shows host
   list      Lists all hosts
   add       Add or edit host
   remove    Removes host
   to        Connects to host
   at        Executes commands
   transfer  Transfers files to/from a host
   encrypt   Encrypts keychain
   config    Shows SSH config
   backup    Backs up SSH configuration and keychain into a .tgz file
   restore   Restores SSH configuration and keychain from a backup
   help, h   Shows a list of commands or help for one command

GLOBAL OPTIONS:
   --keychain value  path to keychain database (default: "/home/user/.sshed") [$SSHED_KEYCHAIN]
   --config value    path to SSH config file (default: "/home/user/.ssh/config") [$SSHED_CONFIG_FILE]
   --ssh-path value  path to SSH binary (default: "ssh") [$SSHED_SSH_BIN]
   --scp-path value  path to SCP binary (default: "scp") [$SSHED_SCP_BIN]
   --help, -h        show help
```

# Bash (ZSH) autocomplete
to enable autocomplete run
```
PROG=sshed source completions/autocomplete.sh
```
if installed with brew, just add those lines to ``.bash_profile`` (``.zshrc``) file
```
PROG=sshed source $(brew --prefix sshed)/autocomplete.sh
```

# Tips
1. To use passwords stored in your keychain, you must install `sshpass`, which allows for the automatic input of a password to SSH.
    However, this method is **strongly discouraged** because it involves passing the password in plain text, which can be exposed to other system users via simple process monitoring tools like `ps`. For better security, it is recommended to use SSH keys for authentication, as they provide a more secure method of connection without exposing sensitive information.

    to install `sshpass` with brew use
    ```
    brew install http://git.io/sshpass.rb
    ```
    for other options see: [https://github.com/kevinburke/sshpass](https://github.com/kevinburke/sshpass)

2. To see all available ssh options run ``man ssh_config``

# TODO
 - [ ] replace sshpass with native go implementation
 - [ ] manage port forwarding
 - [ ] replace scp with something else
 - [ ] handling of ssh options (-c, -E, -f, -T, -t)
 - [ ] key, password generation
