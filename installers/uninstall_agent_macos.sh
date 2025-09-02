#!/bin/bash

sudo_cmd=

KERNEL_NAME=$(uname -s)
ARCHITECTURE=$(uname -m)
AGENT_IS_RUNNING=$(ps aux | grep -v grep | grep docp-agent/bin/agent)
DOCP_FILES_PATH=/opt/docp-agent
USER_GROUP_NAME=docp-agent

# Root user detection
if [ "$UID" == "0" ]; then
    sudo_cmd=''
else
    sudo_cmd='sudo'
fi

#stop and disable launchd
function stop_and_disable() {
  launchctl bootout gui/$(id -u)/com.docp.agent 
  launchctl unload ~/Library/LaunchAgents/com.docp.agent.plist
}

#remove file service
function remove_file_service(){
  $sudo_cmd rm ~/Library/LaunchAgents/com.docp.agent.plist
}

#setup configure e verify machine
function setup(){
  verify_kernel
}

# uninstaller the agent
function _uninstaller(){
  stop_and_disable
  remove_file_service
}

#uninstall agent
function uninstall(){
  setup
 _uninstaller 
}

#run uninstaller
function run(){
  uninstall
}

#verify is linux kernel
function verify_kernel(){
if [[ "$KERNEL_NAME" != "Darwin" ]]; then
  printf "\033[31mInvalid uninstaller for machine\033[0m\n"
  exit 0
fi
}

#actions
run
