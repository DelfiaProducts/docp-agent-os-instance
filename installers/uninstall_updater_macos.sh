#!/bin/bash

sudo_cmd=

KERNEL_NAME=$(uname -s)
ARCHITECTURE=$(uname -m)
MANAGER_IS_RUNNING=$(ps aux | grep -v grep | grep docp-agent/bin/current/updater)
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
  launchctl bootout gui/$(id -u)/com.docp.updater 
  launchctl unload ~/Library/LaunchAgents/com.docp.updater.plist
}

#remove file service
function remove_file_service(){
  $sudo_cmd rm ~/Library/LaunchAgents/com.docp.updater.plist
}


#verify is darwin kernel
function verify_kernel(){
if [[ "$KERNEL_NAME" != "Darwin" ]]; then
  printf "\033[31mInvalid uninstaller for machine\033[0m\n"
  exit 0
fi
}
#setup configure e verify machine
function setup(){
  verify_kernel
}

# uninstaller the manager
function _uninstaller(){
  stop_and_disable
  remove_file_service
}

#uninstall manager
function uninstall(){
  setup
 _uninstaller 
}

#run uninstaller
function run(){
  already_running
}



# verify is already running manager
function already_running(){
if [[ "$MANAGER_IS_RUNNING" ]]; then
  uninstall
else
  printf "\033[31mNot running updater\033[0m\n"
  exit 0
fi
}

#actions
run
