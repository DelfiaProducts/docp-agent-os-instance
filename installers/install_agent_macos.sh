#!/bin/bash

sudo_cmd=

KERNEL_NAME=$(uname -s)
ARCHITECTURE=$(uname -m)
BINARY_URL="https://test-docp-agent-data.s3.amazonaws.com/agent"
VERSION="${VERSION:-${VERSION:-latest}}"
AGENT_IS_RUNNING=$(ps aux | grep -v grep | grep docp-agent/bin/agent)
DOCP_FILES_PATH=/opt/docp-agent
USER_GROUP_NAME=docp-agent

# Root user detection
if [ "$UID" == "0" ]; then
    sudo_cmd=''
else
    sudo_cmd='sudo'
fi

# verify is already running manager
function already_running(){
if [[ "$AGENT_IS_RUNNING" ]]; then
  printf "\033[31mAlready running agent\033[0m\n"
  exit 0
fi
}

#verify is linux kernel
function verify_kernel(){
if [[ "$KERNEL_NAME" != "Darwin" ]]; then
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
  already_running
}

#get binary arm64
function get_binary_arch64(){
  sudo curl -s -o $DOCP_FILES_PATH/bin/agent "$BINARY_URL/$VERSION/macos_arm64"
  sudo chmod +x $DOCP_FILES_PATH/bin/agent
}
#get binary amd64
function get_binary_amd64(){
  sudo curl -s -o $DOCP_FILES_PATH/bin/agent "$BINARY_URL/$VERSION/macos_amd64"
  sudo chmod +x $DOCP_FILES_PATH/bin/agent
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
        <string>com.docp.agent</string>
        <key>EnvironmentVariables</key>
        <dict>
            <key>DOCP_REGISTER_URL</key>
            <string>https://msapi.sandbox.docphq.tech/agents</string>
            <key>DOCP_STATE_CHECK_URL</key>
            <string>https://msapi.sandbox.docphq.tech/agents</string>
            <key>DOCP_AGENT_PORT</key>
            <string>12012</string>
            <key>LOG_LEVEL</key>
            <string>info</string>
            <key>ERROR_LEVEL</key>
            <string>high</string>
        </dict>
        <key>ProgramArguments</key>
        <array>
            <string>/opt/docp-agent/bin/agent</string>
        </array>
        <key>StandardOutPath</key>
        <string>/opt/docp-agent/logs/launchd.log</string>
        <key>StandardErrorPath</key>
        <string>/opt/docp-agent/logs/launchd.log</string>
        <key>ExitTimeOut</key>
        <integer>10</integer>
    </dict>
    </plist>' | sudo tee ~/Library/LaunchAgents/com.docp.agent.plist > /dev/null
}

#prepare launchd
function prepare_launchd() {
  launchctl load ~/Library/LaunchAgents/com.docp.agent.plist
  launchctl start gui/$(id -u)/com.docp.agent
}
#actions
setup
verify_architecture
set_content_service
prepare_launchd
