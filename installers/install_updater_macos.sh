#!/bin/bash

sudo_cmd=

KERNEL_NAME=$(uname -s)
ARCHITECTURE=$(uname -m)
FILE_INDEX_URL="https://docp-agent-k8s.s3.us-east-1.amazonaws.com/index.json"
BINARY_URL="https://github.com/DelfiaProducts/docp-agent-os-instance/releases/download"
VERSION="${VERSION:-${VERSION:-latest}}"
DOCP_FILES_PATH=/opt/docp-agent
USER_GROUP_NAME=docp-agent

# Root user detection
if [ "$UID" == "0" ]; then
    sudo_cmd=''
else
    sudo_cmd='sudo'
fi

#analize options
while [[ $# -gt 0 ]]; do
    key="$1"
    case $key in --version)
        VERSION="$2"
        shift 
        shift 
        ;;
    esac
done


#verify is linux kernel
function verify_kernel(){
if [[ "$KERNEL_NAME" != "Linux" ]]; then
  printf "\033[31mInvalid installer for machine\033[0m\n"
  exit 0
fi
}

#verify is architecture and get binary
function verify_architecture(){
if [[ "$ARCHITECTURE" == "arm64" ]]; then
  get_binary_arch64
fi
if [[ "$ARCHITECTURE" == "x86_64" ]]; then
  get_binary_amd64
fi
}

#setup configure e verify machine
function setup(){
  verify_kernel
}

# Corrige a função resolve_version para extrair corretamente o campo "latest" do JSON
function resolve_version() {
  local version="$1"
  if [[ "$version" == "latest" ]]; then
    version=$(curl -s "$FILE_INDEX_URL" | sed -n 's/.*"latest"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p')
    if [[ -z "$version" ]]; then
      echo "Failed to fetch latest version." >&2
      exit 1
    fi
  fi
  echo "$version"
}

#add permission workdir
function add_perm_work_dir(){
  $sudo_cmd chown -R $USER_GROUP_NAME:$USER_GROUP_NAME /opt/docp-agent/
}

#get binary arm64
function get_binary_arch64(){
  sudo curl -s -o $DOCP_FILES_PATH/bin/releases/$VERSION/updater "$BINARY_URL/$VERSION/updater-macos-arm64"
  sudo chmod +x $DOCP_FILES_PATH/bin/releases/$VERSION/updater
}
#get binary amd64
function get_binary_amd64(){
  sudo curl -s -o $DOCP_FILES_PATH/bin/releases/$VERSION/updater "$BINARY_URL/$VERSION/updater-macos-amd64"
  sudo chmod +x $DOCP_FILES_PATH/bin/releases/$VERSION/updater
}
# create symbolic link
function create_link_simbolic(){
  sudo ln -sfn $DOCP_FILES_PATH/bin/releases/$VERSION/updater $DOCP_FILES_PATH/bin/current/updater
}

#set content service
function set_content_service() {
  printf '
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
    <dict>
        <key>KeepAlive</key>
        <dict>
            <key>SuccessfulExit</key>
            <false/>
        </dict>
        <key>Label</key>
        <string>com.docp.manager</string>
        <key>EnvironmentVariables</key>
        <dict>
            <key>DOCP_AGENT_PORT</key>
            <string>12012</string>
            <key>LOG_LEVEL</key>
            <string>info</string>
            <key>ERROR_LEVEL</key>
            <string>high</string>
        </dict>
        <key>ProgramArguments</key>
        <array>
            <string>/opt/docp-agent/bin/current/updater</string>
        </array>
        <key>StandardOutPath</key>
        <string>/opt/docp-agent/logs/launchd.log</string>
        <key>StandardErrorPath</key>
        <string>/opt/docp-agent/logs/launchd.log</string>
        <key>ExitTimeOut</key>
        <integer>10</integer>
    </dict>
    </plist>' | sudo tee ~/Library/LaunchAgents/com.docp.updater.plist > /dev/null

}

#prepare launchd
function prepare_launchd() {
  launchctl load ~/Library/LaunchAgents/com.docp.updater.plist
  launchctl start gui/$(id -u)/com.docp.updater
}
#actions
VERSION=$(resolve_version "$VERSION")
setup
verify_architecture
create_link_simbolic
add_perm_work_dir
set_content_service
prepare_launchd
