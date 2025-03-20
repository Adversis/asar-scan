# Electron ASAR Scanner

A command-line tool that scans your Windows or macOS for Electron applications and checks if they're using ASAR integrity protection along with .node files.

Pairs nicely with https://github.com/adversis/NodeLoader

## Installation

### From Source

```bash
git clone https://github.com/adversis/asar-scan.git
cd asar-scan
go build -o asarscan cmd/asarscan/*.go

GOOS=windows go build -o asarscan.exe cmd/asarscan/

# Or build using Make
make run  # Builds and runs immediately

# Or build for all platforms
make all  # Creates builds in the dist/ directory
```

## Usage

```bash
# Basic scan
./asarscan

# Enable verbose output
./asarscan -verbose

# Output results in JSON format
./asarscan -json

Results:
========

[1] /Applications/Beeper.app
  Is Electron App: true
  Electron Version: 27.0.2
  Has ASAR File: true
  ASAR Integrity Enabled: true
  OnlyLoadFromAsar Enabled: false
  .node Files (5 found):
    1. /Applications/Beeper.app/Contents/Resources/app.asar.unpacked/.hak/hakModules/keytar/build/Release/keytar.node
    2. /Applications/Beeper.app/Contents/Resources/app.asar.unpacked/.hak/hakModules/matrix-seshat/native/index.node

Summary Table:
===================================================================================
Application                    | Version    | ASAR File  | Integrity  | OnlyLoadFromAsar
===================================================================================
Beeper.app                     | 27.0.2     | Yes        | Yes        | No             
Slack.app                      | 31.0.0     | Yes        | Yes        | No             
Visual Studio Code.app         | 32.2.7     | No         | N/A        | N/A    
```

You might then do something like the following assuming the Terminal has Full Disk Access TCC permissions.

```
cp /Applications/Obsidian.app/Contents/Resources/app.asar.unpacked/node_modules/btime/binding.node binding.node.bak
cp launcher.node /Applications/Obsidian.app/Contents/Resources/app.asar.unpacked/node_modules/btime/binding.node
```

## Resources
  - https://www.adversis.io/blogs/living-off-node-js-addons
  - https://www.atredis.com/blog/2025/3/7/node-is-a-loader
  - https://text.tchncs.de/ioi/backdooring-electron-applications
  - https://github.com/ctxis/beemka 
  
