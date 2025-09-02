#!/bin/bash

sudo_cmd=

KERNEL_NAME=$(uname -s)
ARCHITECTURE=$(uname -m)
DOCP_FILES_PATH=/opt/docp-agent
USER_GROUP_NAME=docp-agent

# Root user detection
if [ "$UID" == "0" ]; then
    sudo_cmd=''
else
    sudo_cmd='sudo'
fi

#stop and disable systemd
function stop_and_disable() {
  $sudo_cmd systemctl stop docp-agent
  $sudo_cmd systemctl disable docp-agent
}

#remove file service
function remove_file_service(){
  $sudo_cmd rm /etc/systemd/system/docp-agent.service
}
# reload daemon
function reload_daemon(){
 $sudo_cmd systemctl daemon-reload 
}

#setup configure e verify machine
function setup(){
  verify_kernel
}

# uninstaller the manager
function _uninstaller(){
  stop_and_disable
  remove_file_service
  reload_daemon
}

#uninstall manager
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
if [[ "$KERNEL_NAME" != "Linux" ]]; then
  printf "\033[31mInvalid uninstaller for machine\033[0m\n"
  exit 0
fi
}

#actions
run
