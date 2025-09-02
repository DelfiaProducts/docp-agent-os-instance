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
  $sudo_cmd systemctl stop docp-manager
  $sudo_cmd systemctl disable docp-manager
}

#remove file service
function remove_file_service(){
  $sudo_cmd rm /etc/systemd/system/docp-manager.service
}

#remove work dir
function remove_work_dir() {
  $sudo_cmd rm -rf $DOCP_FILES_PATH 
}

# remove user and group
function remove_user_and_group(){
  $sudo_cmd userdel $USER_GROUP_NAME > /dev/null 2>&1
  $sudo_cmd groupdel $USER_GROUP_NAME > /dev/null 2>&1
}

# remove perm sudoers file
function remove_perm_sudoers_file(){
  $sudo_cmd sed -i '/^docp-agent ALL=(ALL) NOPASSWD: ALL/d' /etc/sudoers
}

#setup configure e verify machine
function setup(){
  verify_kernel
}

# uninstaller the manager
function _uninstaller(){
  stop_and_disable
  remove_file_service
  remove_work_dir
  remove_user_and_group
  remove_perm_sudoers_file
}

#uninstall manager
function uninstall(){
  setup
 _uninstaller 
}

#verify is linux kernel
function verify_kernel(){
if [[ "$KERNEL_NAME" != "Linux" ]]; then
  printf "\033[31mInvalid uninstaller for machine\033[0m\n"
  exit 0
fi
}


#actions
uninstall
